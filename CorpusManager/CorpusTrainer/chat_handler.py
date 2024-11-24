import subprocess
import threading
import logging
from utils import resource_path, chat_logger
import os

class ChatHandler:
    def __init__(self, gui_display_callback, config, update_status_callback):
        self.gui_display = gui_display_callback
        self.update_status = update_status_callback
        self.chat_process = None
        self.chat_thread = None
        self.is_ready = False
        self.config = config
        self._buffer = ''
        self.executable = 'chat.exe'

    def set_executable(self, executable):
        self.executable = executable

    def start_chat_process(self):
        try:
            chat_exe = resource_path(self.executable) if os.name == 'nt' else resource_path(self.executable.replace('.exe', ''))
            config_file = self.config.get('config')
            corpus_file = self.config.get('c')

            if not os.path.exists(chat_exe):
                self.gui_display("System", f"Chat executable not found at {chat_exe}")
                logging.getLogger('gui_logger').error(f'Chat executable not found at {chat_exe}')
                return

            if not os.path.exists(config_file):
                self.gui_display("System", f"Config file not found at {config_file}")
                logging.getLogger('gui_logger').error(f'Config file not found at {config_file}')
                return

            if not os.path.exists(corpus_file):
                self.gui_display("System", f"Corpus file not found at {corpus_file}")
                logging.getLogger('gui_logger').error(f'Corpus file not found at {corpus_file}')
                return

            cmd = [chat_exe, '-config', config_file, '-c', corpus_file]

            if self.config.get('anim'):
                cmd.append('-anim')

            if self.config.get('cmem') is not None:
                cmd.extend(['-cmem', str(self.config['cmem'])])

            if self.config.get('intro'):
                cmd.append('-intro')

            if self.config.get('t') is not None:
                cmd.extend(['-t', str(self.config['t'])])

            if self.config.get('context'):
                cmd.append('-context')

            if self.config.get('dev'):
                cmd.append('-dev')

            chat_logger.info(f'Starting {self.executable} with command: {" ".join(cmd)}')

            self.chat_process = subprocess.Popen(
                cmd,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1
            )

            self.chat_thread = threading.Thread(target=self.listen_to_chat, daemon=True)
            self.chat_thread.start()

        except Exception as e:
            self.gui_display("System", f"Error starting {self.executable}: {e}")
            logging.getLogger('gui_logger').error(f'Error starting {self.executable}: {e}')

    def listen_to_chat(self):
        try:
            for line in iter(self.chat_process.stdout.readline, ''):
                response = line.strip()
                if response:
                    if "Loading model..." in response:
                        continue
                    if "Model loaded successfully!" in response:
                        self.is_ready = True
                        self.update_status("Ready")
                        chat_logger.info("ChatBot is ready.")
                        continue
                    sender, message = self.parse_response(response)
                    self.gui_display(sender, message)
                    chat_logger.info(f'Received from {self.executable}: {response}')
        except Exception as e:
            self.gui_display("System", f"Error reading from {self.executable}: {e}")
            logging.getLogger('gui_logger').error(f'Error reading from {self.executable}: {e}')

    def parse_response(self, response):
        self.update_status("PeriChat Processed and Responded")
        message = response.replace("User: PeriChat: ", "").strip()
        return ("ChatBot", message)

    def send_message(self, message):
        if not self.chat_process:
            self.gui_display("System", "ChatBot process not started.")
            self.update_status("PeriChat Initializing")
            logging.getLogger('gui_logger').warning("Attempted to send message before starting chat process.")
            return

        if self.chat_process and self.is_ready:
            try:
                self.chat_process.stdin.write(message + "\n")
                self.chat_process.stdin.flush()
                chat_logger.info(f'Sent to {self.executable}: {message}')
            except Exception as e:
                self.gui_display("System", f"Error sending message to {self.executable}: {e}")
                logging.getLogger('gui_logger').error(f'Error sending message to {self.executable}: {e}')
        elif not self.is_ready:
            self.update_status("PeriChat Initializing")
            self.gui_display("System", "ChatBot is still loading. Please wait...")
            logging.getLogger('gui_logger').warning("Attempted to send message before ChatBot was ready.")
