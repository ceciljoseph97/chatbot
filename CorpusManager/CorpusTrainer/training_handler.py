import subprocess
import logging
from utils import resource_path
import os

class TrainingHandler:
    def __init__(self, gui_display_callback):
        self.gui_display = gui_display_callback
        self.train_process = None

    def run_training(self, config_file, corpus_dir, store_file):
        try:
            train_exe = resource_path("train.exe") if os.name == 'nt' else resource_path("train")
            if not train_exe or not os.path.exists(train_exe):
                self.gui_display("System", f"Training executable not found at {train_exe}")
                logging.error(f'Training executable not found at {train_exe}')
                return

            cmd = [
                train_exe,
                "--config", config_file,
                "-d", corpus_dir,
                "-o", store_file,
                "-m"

            ]

            logging.info(f'Executing train.exe with command: {" ".join(cmd)}')

            self.train_process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )

            stdout, stderr = self.train_process.communicate()

            if self.train_process.returncode != 0:
                self.gui_display("System", f"Training failed with error:\n{stderr}")
                logging.error(f'Training failed with error:\n{stderr}')
            else:
                self.gui_display("System", "Training completed successfully.")
                logging.info("Training completed successfully.")

        except Exception as e:
            self.gui_display("System", f"Error during training: {e}")
            logging.error(f'Error during training: {e}')
