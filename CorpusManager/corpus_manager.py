import os
import yaml
import logging
from googletrans import Translator
from spellchecker import SpellChecker

def list_yml_files(directory):
    try:
        yml_files = [f for f in os.listdir(directory) if f.endswith(('.yml', '.yaml'))]
        return yml_files
    except Exception as e:
        logging.error(f'Error listing YAML files in {directory}: {e}')
        return []

def load_corpus_from_file(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            corpus = yaml.safe_load(file)
            return corpus
    except Exception as e:
        logging.error(f'Error loading corpus from {file_path}: {e}')
        return None

def save_corpus_to_file(corpus, file_path):
    try:
        with open(file_path, 'w', encoding='utf-8') as file:
            yaml.dump(corpus, file, allow_unicode=True)
        logging.info(f"Corpus saved to {file_path}")
        return f"Corpus saved successfully to {file_path}."
    except Exception as e:
        logging.error(f'Error saving corpus to {file_path}: {e}')
        return f"Failed to save corpus to {file_path}: {e}"

def translate_and_correct(text, dest_lang='de'):
    try:
        translator = Translator()
        translated = translator.translate(text, dest=dest_lang).text
        logging.debug(f'Translated "{text}" to "{translated}"')

        spell = SpellChecker(language=dest_lang)
        corrected_words = []
        for word in translated.split():
            correction = spell.correction(word)
            corrected_words.append(correction if correction else word)
        corrected = " ".join(corrected_words)
        logging.debug(f'Spell-corrected translation: "{corrected}"')

        return corrected
    except Exception as e:
        logging.error(f'Error translating and correcting text: {e}')
        return text

def create_german_corpus(corpus):
    try:
        german_corpus = {
            'categories': [translate_and_correct(cat, dest_lang='de') for cat in corpus.get('categories', [])],
            'conversations': [
                [translate_and_correct(q, dest_lang='de'), translate_and_correct(a, dest_lang='de')]
                for q, a in corpus.get('conversations', [])
            ]
        }
        logging.info('Created German corpus')
        return german_corpus
    except Exception as e:
        logging.error(f'Error creating German corpus: {e}')
        return {}
