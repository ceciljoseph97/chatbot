import logging
import os
import sys

chat_logger = logging.getLogger('chat_logger')
chat_logger.setLevel(logging.INFO)
chat_handler = logging.FileHandler('chat_logs.log')
chat_handler.setFormatter(logging.Formatter('%(asctime)s - %(levelname)s - %(message)s'))
chat_logger.addHandler(chat_handler)

training_logger = logging.getLogger('training_logger')
training_logger.setLevel(logging.INFO)
training_handler = logging.FileHandler('training_logs.log')
training_handler.setFormatter(logging.Formatter('%(asctime)s - %(levelname)s - %(message)s'))
training_logger.addHandler(training_handler)

gui_logger = logging.getLogger('gui_logger')
gui_logger.setLevel(logging.DEBUG)
gui_file_handler = logging.FileHandler('corpus_trainer_gui.log')
gui_file_handler.setFormatter(logging.Formatter('%(asctime)s - %(levelname)s - %(message)s'))
gui_logger.addHandler(gui_file_handler)

def resource_path(relative_path):
    """ Get absolute path to resource, works for dev and for PyInstaller """
    try:
        base_path = sys._MEIPASS
    except Exception:
        base_path = os.path.abspath(".")

    return os.path.join(base_path, relative_path)
