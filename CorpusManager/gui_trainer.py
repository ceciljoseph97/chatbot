import tkinter as tk
from tkinter import ttk, messagebox, filedialog
from corpus_manager import (
    list_yml_files, load_corpus_from_file, save_corpus_to_file,
    translate_and_correct, create_german_corpus
)
import os
import logging
import sys

from utils import resource_path
from chat_handler import ChatHandler
from training_handler import TrainingHandler
from advanced_commands import AdvancedCommands
from PIL import Image, ImageTk

logging.basicConfig(
    filename='corpus_trainer_gui.log',
    filemode='a',
    format='%(asctime)s - %(levelname)s - %(message)s',
    level=logging.DEBUG
)

class CorpusTrainerApp:
    def __init__(self, root):
        self.root = root
        self.root.title("Corpus Trainer Application")
        self.current_directory = os.getcwd()
        self.current_file = None

        self.advanced_commands = AdvancedCommands()
        self.chat_config = {
            'anim': False,
            'c': resource_path("gob/PMFuncOverView.gob"),
            'cmem': 2,
            'config': resource_path("config_local.yaml"),
            'context': False,
            'intro': True,
            'dev': False,
            't': 1
        }

        self.chat_exe_var = tk.StringVar(value='chat.exe')

        self.chat_handler = ChatHandler(self.display_message, self.chat_config, self.update_status)
        self.training_handler = TrainingHandler(self.display_message)

        self.create_widgets()

    def create_widgets(self):
        notebook = ttk.Notebook(self.root)
        notebook.pack(fill="both", expand=True, padx=10, pady=10)
        home_frame = ttk.Frame(notebook)
        notebook.add(home_frame, text="Home")
        self.create_home_tab(home_frame)
        corpus_frame = ttk.Frame(notebook)
        notebook.add(corpus_frame, text="Corpus Management")
        self.create_corpus_management_tab(corpus_frame)
        chat_frame = ttk.Frame(notebook)
        notebook.add(chat_frame, text="Chat")
        self.create_chat_tab(chat_frame)
        training_frame = ttk.Frame(notebook)
        notebook.add(training_frame, text="Training")
        self.create_training_tab(training_frame)
        about_frame = ttk.Frame(notebook)
        notebook.add(about_frame, text="About")
        self.create_about_tab(about_frame)
        self.status_var = tk.StringVar()
        self.status_var.set("Ready")
        status_bar = ttk.Label(self.root, textvariable=self.status_var, relief='sunken', anchor='w')
        status_bar.pack(side='bottom', fill='x')

    def create_home_tab(self, parent):
        logo_path = resource_path('logo.png')
        try:
            logo_image = Image.open(logo_path)
            logo_image = logo_image.resize((800, 400))
            self.logo_photo = ImageTk.PhotoImage(logo_image)
            logo_label = ttk.Label(parent, image=self.logo_photo)
            logo_label.pack(pady=10)
        except Exception as e:
            logging.error(f'Error loading logo image: {e}')
            messagebox.showerror("Image Load Error", f"Failed to load logo image: {e}")

        description = (
            "Welcome to the CorpusManager Perichat Tool!\n\n"
            "This application is designed to assist in creating, managing, and translating corpora "
            "while providing a chat interface for testing. Train your corpus models, configure "
            "chat settings, and start interacting with your models seamlessly."
        )
        description_label = ttk.Label(parent, text=description, wraplength=700, justify="center")
        description_label.pack(pady=10)

    def create_about_tab(self, parent):
        about_text = (
            "CorpusManager Perichat Tool\n"
            "Version: 1.0.0\n\n"
            "This tool combines corpus management and chat functionality for efficient model training "
            "and testing.\n" 
            "Use it to:\n"
            "- Create and edit corpora\n"
            "- Translate corpora into multiple languages\n"
            "- Train models using advanced configurations\n"
            "- Interact with trained models via chat\n\n"
            "Developed by Cecil Joseph\n"
            ""
        )
        about_label = ttk.Label(parent, text=about_text, wraplength=700, justify="left")
        about_label.pack(pady=20)

    def create_corpus_management_tab(self, parent):

        file_frame = ttk.LabelFrame(parent, text="YAML Files")
        file_frame.pack(fill="x", padx=10, pady=5)

        dir_select_frame = ttk.Frame(file_frame)
        dir_select_frame.pack(fill="x", padx=10, pady=5)

        dir_label = ttk.Label(dir_select_frame, text="Corpus Directory:")
        dir_label.pack(side="left")

        self.dir_display = ttk.Entry(dir_select_frame, width=50)
        self.dir_display.pack(side="left", padx=5)
        self.dir_display.insert(0, self.current_directory)

        browse_dir_btn = ttk.Button(dir_select_frame, text="Browse", command=self.browse_directory)
        browse_dir_btn.pack(side="left", padx=5)

        self.file_listbox = tk.Listbox(file_frame, height=6, selectmode='extended')
        self.file_listbox.pack(side="left", fill="both", expand=True, padx=(10, 0), pady=10)

        scrollbar = ttk.Scrollbar(file_frame, orient="vertical", command=self.file_listbox.yview)
        scrollbar.pack(side="left", fill="y", pady=10)
        self.file_listbox.config(yscrollcommand=scrollbar.set)

        btn_frame = ttk.Frame(file_frame)
        btn_frame.pack(side="left", fill="y", padx=10, pady=10)

        edit_btn = ttk.Button(btn_frame, text="Edit Selected", command=self.edit_file)
        edit_btn.pack(fill="x", pady=(0, 5))

        new_btn = ttk.Button(btn_frame, text="Create New", command=self.new_file)
        new_btn.pack(fill="x")

        edit_frame = ttk.LabelFrame(parent, text="Edit Corpus")
        edit_frame.pack(fill="both", expand=True, padx=10, pady=5)

        cat_label = ttk.Label(edit_frame, text="Categories:")
        cat_label.grid(row=0, column=0, sticky="nw", padx=10, pady=5)

        self.cat_text = tk.Text(edit_frame, height=10, width=40)
        self.cat_text.grid(row=1, column=0, padx=10, pady=5)

        conv_label = ttk.Label(edit_frame, text="Conversations (/Q: Question /A: Answer):")
        conv_label.grid(row=0, column=1, sticky="nw", padx=10, pady=5)

        self.conv_text = tk.Text(edit_frame, height=10, width=40)
        self.conv_text.grid(row=1, column=1, padx=10, pady=5)

        action_frame = ttk.Frame(edit_frame)
        action_frame.grid(row=2, column=0, columnspan=2, pady=10)

        translate_btn = ttk.Button(action_frame, text="Translate Selected", command=self.translate_corpus)
        translate_btn.pack(side="left", padx=5)

        save_btn = ttk.Button(action_frame, text="Save Corpus", command=self.save_corpus)
        save_btn.pack(side="left", padx=5)

    def create_chat_tab(self, parent):
        config_frame = ttk.LabelFrame(parent, text="Chat Configuration")
        config_frame.pack(fill="x", padx=10, pady=5)

        exe_label = ttk.Label(config_frame, text="Select Chat Executable:")
        exe_label.grid(row=0, column=0, sticky="w", padx=10, pady=5)

        exe_frame = ttk.Frame(config_frame)
        exe_frame.grid(row=0, column=1, padx=10, pady=5, sticky="w")

        chat_exe_rb = ttk.Radiobutton(
            exe_frame, text="chat.exe", variable=self.chat_exe_var, value='chat.exe'
        )
        chat_exe_rb.pack(side="left", padx=(0, 10))

        chat_closest_exe_rb = ttk.Radiobutton(
            exe_frame, text="chatClosest.exe", variable=self.chat_exe_var, value='chatClosest.exe'
        )
        chat_closest_exe_rb.pack(side="left")

        row = 1

        for cmd in self.advanced_commands.get_commands():
            label = ttk.Label(config_frame, text=f"{cmd['name']} ({cmd['flag']}):")
            label.grid(row=row, column=0, sticky="w", padx=10, pady=5)

            flag = cmd['flag'].strip('-')

            if cmd['type'] == 'boolean':
                var = tk.BooleanVar(value=self.chat_config.get(flag, False))
                setattr(self, f"{flag}_var", var)
                checkbutton = ttk.Checkbutton(config_frame, variable=var)
                checkbutton.grid(row=row, column=1, padx=10, pady=5, sticky="w")

            elif cmd['type'] == 'string':
                entry = ttk.Entry(config_frame, width=40)
                entry.grid(row=row, column=1, padx=10, pady=5)
                current_value = self.chat_config.get(flag, "")
                entry.insert(0, current_value)
                setattr(self, f"{flag}_entry", entry)
                if cmd['flag'] in ['-c', '-config']:
                    browse_btn = ttk.Button(
                        config_frame, text="Browse", command=lambda c=flag: self.browse_file(c)
                    )
                    browse_btn.grid(row=row, column=2, padx=5, pady=5)
            elif cmd['type'] == 'integer':
                entry = ttk.Entry(config_frame, width=40)
                entry.grid(row=row, column=1, padx=10, pady=5)
                current_value = self.chat_config.get(flag, "")
                entry.insert(0, current_value)
                setattr(self, f"{flag}_entry", entry)
            row += 1

        execute_btn = ttk.Button(config_frame, text="Execute", command=self.execute_chat)
        execute_btn.grid(row=row, column=1, padx=10, pady=10, sticky="e")

        self.chat_display_frame = ttk.Frame(parent)
        self.chat_display_frame.pack(fill="both", expand=True, padx=10, pady=5)
        self.chat_display_frame.pack_forget()

        self.chat_display = tk.Text(self.chat_display_frame, state='disabled', wrap='word')
        self.chat_display.pack(fill="both", expand=True, side="left", padx=(0, 5), pady=5)

        scrollbar = ttk.Scrollbar(self.chat_display_frame, orient="vertical", command=self.chat_display.yview)
        scrollbar.pack(side="left", fill="y", pady=5)
        self.chat_display.config(yscrollcommand=scrollbar.set)

        self.input_frame = ttk.Frame(parent)
        self.input_frame.pack(fill="x", padx=10, pady=5)
        self.input_frame.pack_forget()

        self.chat_entry = ttk.Entry(self.input_frame, width=80)
        self.chat_entry.pack(side="left", padx=(0, 5), pady=5, fill="x", expand=True)
        self.chat_entry.bind("<Return>", self.send_chat_message)

        send_btn = ttk.Button(self.input_frame, text="Send", command=self.send_chat_message)
        send_btn.pack(side="left", padx=(5, 0), pady=5)

    def create_training_tab(self, parent):
        training_input_frame = ttk.LabelFrame(parent, text="Training Session")
        training_input_frame.pack(fill="both", expand=True, padx=10, pady=5)

        config_label = ttk.Label(training_input_frame, text="Config File:")
        config_label.grid(row=0, column=0, sticky="w", padx=10, pady=5)

        self.config_entry = ttk.Entry(training_input_frame, width=50)
        self.config_entry.grid(row=0, column=1, padx=10, pady=5)
        self.config_entry.insert(0, resource_path("config_local.yaml"))

        config_browse_btn = ttk.Button(training_input_frame, text="Browse", command=self.browse_config_file)
        config_browse_btn.grid(row=0, column=2, padx=5, pady=5)

        dir_label = ttk.Label(training_input_frame, text="Corpus Directory:")
        dir_label.grid(row=1, column=0, sticky="w", padx=10, pady=5)

        self.training_dir_entry = ttk.Entry(training_input_frame, width=50)
        self.training_dir_entry.grid(row=1, column=1, padx=10, pady=5)

        dir_browse_btn = ttk.Button(training_input_frame, text="Browse", command=self.browse_training_directory)
        dir_browse_btn.grid(row=1, column=2, padx=5, pady=5)

        files_label = ttk.Label(training_input_frame, text="Corpus (GOB):")
        files_label.grid(row=2, column=0, sticky="w", padx=10, pady=5)

        self.store_file_entry = ttk.Entry(training_input_frame, width=50)
        self.store_file_entry.grid(row=2, column=1, padx=10, pady=5)

        files_browse_btn = ttk.Button(training_input_frame, text="Browse", command=self.save_gob)
        files_browse_btn.grid(row=2, column=2, padx=5, pady=5)

        train_btn = ttk.Button(training_input_frame, text="Train", command=self.run_training)
        train_btn.grid(row=4, column=1, pady=10)

    def refresh_file_list(self):
        self.file_listbox.delete(0, tk.END)
        yml_files = list_yml_files(directory=self.current_directory)
        for file in yml_files:
            self.file_listbox.insert(tk.END, file)
        logging.debug(f'YAML file list refreshed for directory: {self.current_directory}')

    def browse_directory(self):
        directory = filedialog.askdirectory()
        if directory:
            self.current_directory = directory
            self.dir_display.delete(0, tk.END)
            self.dir_display.insert(0, directory)
            self.refresh_file_list()
            logging.info(f'Selected directory: {directory}')

    def save_gob(self):
        file_name = filedialog.asksaveasfilename(defaultextension=".gob",
                                                 initialdir=self.current_directory,
                                                 filetypes=[("GOB files", "*.gob")])
        if file_name:
            self.store_file_entry.delete(0, tk.END)
            self.store_file_entry.insert(0, file_name)
            logging.info(f'Selected storage file: {file_name}')
        return file_name

    def browse_config_file(self):
        file = filedialog.askopenfilename(defaultextension=".yaml",
                                          filetypes=[("YAML files", "*.yml"), ("YAML files", "*.yaml")])
        if file:
            self.config_entry.delete(0, tk.END)
            self.config_entry.insert(0, file)
            logging.info(f'Selected config file: {file}')

    def browse_file(self, flag):
        if flag == 'c':
            file = filedialog.askopenfilename(defaultextension=".gob",
                                              filetypes=[("GOB files", "*.gob")])
        elif flag == 'config':
            file = filedialog.askopenfilename(defaultextension=".yaml",
                                              filetypes=[("YAML files", "*.yml"), ("YAML files", "*.yaml")])
        else:
            file = None

        if file:
            entry = getattr(self, f"{flag}_entry")
            entry.delete(0, tk.END)
            entry.insert(0, file)
            logging.info(f'Selected {flag} file: {file}')

    def browse_training_directory(self):
        directory = filedialog.askdirectory()
        if directory:
            self.training_dir_entry.delete(0, tk.END)
            self.training_dir_entry.insert(0, directory)
            logging.info(f'Selected training directory: {directory}')

    def execute_chat(self):
        try:
            for cmd in self.advanced_commands.get_commands():
                flag = cmd['flag'].strip('-')
                if cmd['type'] == 'boolean':
                    var = getattr(self, f"{flag}_var").get()
                    self.chat_config[flag] = var
                elif cmd['type'] == 'string':
                    entry = getattr(self, f"{flag}_entry")
                    self.chat_config[flag] = entry.get().strip()
                elif cmd['type'] == 'integer':
                    entry = getattr(self, f"{flag}_entry")
                    value = entry.get().strip()
                    self.chat_config[flag] = int(value) if value.isdigit() else self.chat_config.get(flag, None)

            if not self.chat_config.get('c'):
                messagebox.showwarning("Input Error", "Please specify the corpus file (-c).")
                return
            if not self.chat_config.get('config'):
                messagebox.showwarning("Input Error", "Please specify the config file (-config).")
                return

            self.chat_handler.set_executable(self.chat_exe_var.get())

            self.chat_handler.start_chat_process()

            self.status_var.set("ChatBot is initializing...")
            logging.info("ChatBot initialization started.")

            self.chat_display_frame.pack_forget()
            self.input_frame.pack_forget()

        except Exception as e:
            self.display_message("System", f"Error executing chat: {e}")
            logging.error(f'Error executing chat: {e}')

    def run_training(self):
        try:
            config_file = self.config_entry.get().strip()
            corpus_dir = self.training_dir_entry.get().strip()
            store_file = self.store_file_entry.get().strip()

            if not config_file:
                messagebox.showwarning("Input Error", "Please specify the config file.")
                return

            if not corpus_dir:
                messagebox.showwarning("Input Error", "Please specify the corpus directory.")
                return

            if not store_file:
                messagebox.showwarning("Input Error", "Please specify the store file.")
                return

            if not store_file.lower().endswith('.gob'):
                messagebox.showwarning("Input Error", "Store file must have a .gob extension.")
                return

            self.status_var.set("Training in progress...")
            logging.info("Initiating training process.")

            self.training_handler.run_training(
                config_file=config_file,
                corpus_dir=corpus_dir,
                store_file=store_file,
            )

            self.status_var.set("Training completed.")
            logging.info("Training process finished.")

        except Exception as e:
            self.display_message("System", f"Error initiating training: {e}")
            logging.error(f'Error initiating training: {e}')

    def edit_file(self):
        selected_indices = self.file_listbox.curselection()
        if not selected_indices:
            messagebox.showwarning("No selection", "Please select at least one file to edit.")
            return

        if len(selected_indices) > 1:
            messagebox.showinfo("Multiple Selection", "Please select only one file to edit at a time.")
            return

        idx = selected_indices[0]
        file_name = self.file_listbox.get(idx)
        full_path = os.path.join(self.current_directory, file_name)
        corpus = load_corpus_from_file(full_path)
        if not corpus:
            messagebox.showerror("Load Error", f"Failed to load corpus from {file_name}.")
            return

        categories = "\n".join(corpus.get('categories', []))
        conversations = "\n".join([f"/Q: {qa[0]}\n/A: {qa[1]}" for qa in corpus.get('conversations', [])])
        self.cat_text.delete('1.0', tk.END)
        self.cat_text.insert(tk.END, categories)
        self.conv_text.delete('1.0', tk.END)
        self.conv_text.insert(tk.END, conversations)
        self.current_file = full_path
        messagebox.showinfo("Edit Mode", f"Editing {file_name}")
        logging.info(f'Editing file: {full_path}')

    def new_file(self):
        self.cat_text.delete('1.0', tk.END)
        self.conv_text.delete('1.0', tk.END)
        self.current_file = None
        messagebox.showinfo("New Corpus", "Enter details for the new corpus.")
        logging.info('Creating a new corpus')

    def save_corpus(self):
        try:
            categories = [cat.strip() for cat in self.cat_text.get('1.0', tk.END).strip().split('\n') if cat.strip()]
            conversations = []
            conv_lines = self.conv_text.get('1.0', tk.END).strip().split('\n')
            temp_q = None
            temp_a = None
            for idx, line in enumerate(conv_lines, start=1):
                if line.startswith("/Q:"):
                    temp_q = line[3:].strip()
                elif line.startswith("/A:"):
                    temp_a = line[3:].strip()
                    if temp_q and temp_a:
                        conversations.append([temp_q, temp_a])
                        temp_q = None
                        temp_a = None
                    else:
                        messagebox.showwarning("Format Error", f"Empty question or answer on line {idx}.")
                        return
                else:
                    messagebox.showwarning("Format Error", f"Lines must start with '/Q:' or '/A:'. Error on line {idx}.")
                    return

            if not categories:
                messagebox.showwarning("Input Error", "Please enter at least one category.")
                return
            if not conversations:
                messagebox.showwarning("Input Error", "Please enter at least one conversation.")
                return

            corpus = {
                'categories': categories,
                'conversations': conversations
            }

            if self.current_file:
                file_name = self.current_file
            else:
                file_name = filedialog.asksaveasfilename(defaultextension=".yml",
                                                         initialdir=self.current_directory,
                                                         filetypes=[("YAML files", "*.yml"), ("YAML files", "*.yaml")])
                if not file_name:
                    return

            save_message = save_corpus_to_file(corpus, file_name)
            messagebox.showinfo("Saved", save_message)
            self.refresh_file_list()
            self.current_file = file_name
            logging.info(f'Corpus saved to {file_name}')
        except Exception as e:
            logging.error(f'Error saving corpus: {e}')
            messagebox.showerror("Save Error", f"An error occurred while saving: {e}")

    def translate_corpus(self):
        try:
            selected_indices = self.file_listbox.curselection()
            if not selected_indices:
                messagebox.showwarning("No selection", "Please select at least one file to translate.")
                return

            translated_files = []

            for idx in selected_indices:
                file_name = self.file_listbox.get(idx)
                full_path = os.path.join(self.current_directory, file_name)
                corpus = load_corpus_from_file(full_path)
                if not corpus:
                    logging.warning(f'Skipping file due to load failure: {full_path}')
                    continue

                german_corpus = create_german_corpus(corpus)
                if not german_corpus:
                    messagebox.showerror("Translation Error", f"Failed to create German corpus for {file_name}.")
                    logging.error(f'Failed to create German corpus for {full_path}')
                    continue

                base_name = os.path.splitext(os.path.basename(file_name))[0]
                german_file_name = base_name + '_de.yml'
                german_full_path = os.path.join(self.current_directory, german_file_name)
                save_message = save_corpus_to_file(german_corpus, german_full_path)
                translated_files.append(german_file_name)
                logging.info(f'German corpus saved as {german_full_path}')

            if translated_files:
                messagebox.showinfo("Translation Complete", f"Translated files:\n" + "\n".join(translated_files))
            else:
                messagebox.showinfo("No Translations", "No files were translated.")

        except Exception as e:
            logging.error(f'Error translating corpus: {e}')
            messagebox.showerror("Translation Error", f"An error occurred during translation: {e}")

    def send_chat_message(self, event=None):
        user_message = self.chat_entry.get().strip()
        if not user_message:
            return

        self.chat_entry.delete(0, tk.END)
        self.display_message("You", user_message)

        if user_message.startswith('/'):
            processed_message = user_message
        else:
            processed_message = f"PeriChat: {user_message}"

        logging.info(f'Processed message to send: {processed_message}')

        self.chat_handler.send_message(processed_message)

    def display_message(self, sender, message):
        self.chat_display.config(state='normal')
        self.chat_display.insert(tk.END, f"{sender}: {message}\n")
        self.chat_display.config(state='disabled')
        self.chat_display.see(tk.END)

        if sender in ["ChatBot", "PeriChat"] and not self.chat_display_frame.winfo_viewable():
            self.chat_display_frame.pack(fill="both", expand=True, padx=10, pady=5)
            self.input_frame.pack(fill="x", padx=10, pady=5)
            self.status_var.set("Ready")
            logging.info("Chat display area is now visible.")

    def update_status(self, message):
        self.status_var.set(message)
        logging.info(f'Status updated to: {message}')

    def on_close(self):
        if self.chat_handler.chat_process:
            try:
                self.chat_handler.chat_process.terminate()
                self.chat_handler.chat_process.wait(timeout=5)
                logging.info("Chat process terminated successfully.")
            except Exception as e:
                logging.error(f'Error terminating chat process: {e}')
        self.root.destroy()

def main():
    try:
        root = tk.Tk()
        app = CorpusTrainerApp(root)
        root.protocol("WM_DELETE_WINDOW", app.on_close)
        root.mainloop()
    except Exception as e:
        logging.exception("An unexpected error occurred:")
        messagebox.showerror("Unexpected Error", f"An unexpected error occurred:\n{e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
