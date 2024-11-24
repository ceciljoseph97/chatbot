import os
import yaml
from googletrans import Translator
from spellchecker import SpellChecker
from textblob import TextBlob

def get_user_input(prompt):
    return input(prompt)

def list_yml_files():
    return [f for f in os.listdir() if f.endswith('.yml')]

def load_corpus_from_file(file_name):
    with open(file_name, 'r') as file:
        return yaml.safe_load(file)

def translate_and_correct(text, dest_lang='de'):
    translator = Translator()
    translated_text = translator.translate(text, dest=dest_lang).text

    if dest_lang == 'de':
        blob = TextBlob(translated_text)
        corrected_text = str(blob.correct())
    else:
        spell = SpellChecker()
        corrected_words = []
        for word in translated_text.split():
            correction = spell.correction(word)
            if correction is None:
                correction = word
            corrected_words.append(correction)
        corrected_text = " ".join(corrected_words)
    
    return corrected_text

def edit_corpus(corpus, is_german=False):
    lang = 'de' if is_german else 'en'
    
    print("Current categories: ", corpus.get('categories', []))
    add_categories = get_user_input("Do you want to add new categories? (yes/no): ").lower()
    if add_categories == 'yes':
        print("Enter new categories (type 'done' to finish):")
        while True:
            category = get_user_input("Category: ")
            if category.lower() == 'done':
                break
            if not is_german:
                corpus.setdefault('categories', []).append(category)
            else:
                category = translate_and_correct(category, dest_lang=lang)
                corpus.setdefault('categories', []).append(category)

    print("Current conversations: ", corpus.get('conversations', []))
    add_conversations = get_user_input("Do you want to add new conversations? (yes/no): ").lower()
    if add_conversations == 'yes':
        print("Enter new conversations (type 'done' to finish):")
        while True:
            question = get_user_input("Question: ")
            if question.lower() == 'done':
                break
            answer = get_user_input("Answer: ")
            if not is_german:
                corpus.setdefault('conversations', []).append([question, answer])
            else:
                question = translate_and_correct(question, dest_lang=lang)
                answer = translate_and_correct(answer, dest_lang=lang)
                corpus.setdefault('conversations', []).append([question, answer])

    return corpus

def create_corpus():
    categories = []
    print("Enter categories (type 'done' to finish):")
    while True:
        category = get_user_input("Category: ")
        if category.lower() == 'done':
            break
        categories.append(category)

    conversations = []
    print("Enter conversations (type 'done' to finish):")
    while True:
        question = get_user_input("Question: ")
        if question.lower() == 'done':
            break
        answer = get_user_input("Answer: ")
        conversations.append([question, answer])

    corpus = {
        'categories': categories,
        'conversations': conversations
    }

    return corpus

def save_corpus_to_file(corpus, file_name):
    with open(file_name, 'w') as file:
        yaml.dump(corpus, file, default_flow_style=False)
    print(f"Corpus saved to {file_name}")

def create_and_save_german_corpus(english_corpus, base_file_name):
    german_corpus = {
        'categories': [translate_and_correct(cat, dest_lang='de') for cat in english_corpus.get('categories', [])],
        'conversations': [
            [translate_and_correct(q, dest_lang='de'), translate_and_correct(a, dest_lang='de')]
            for q, a in english_corpus.get('conversations', [])
        ]
    }
    german_file_name = base_file_name + '_de.yml'
    save_corpus_to_file(german_corpus, german_file_name)

def ensure_german_file_exists(english_file_name):
    german_file_name = english_file_name.replace('.yml', '_de.yml')
    return german_file_name if os.path.exists(german_file_name) else None

def main():
    while True:
        yml_files = list_yml_files()
        if yml_files:
            print("Existing YAML files:")
            for idx, file in enumerate(yml_files):
                print(f"{idx + 1}. {file}")
            choice = get_user_input("Do you want to edit an existing file or create a new one? (edit/new): ").lower()
            if choice == 'edit':
                file_index = int(get_user_input(f"Enter the number of the file you want to edit (1-{len(yml_files)}): ")) - 1
                file_name = yml_files[file_index]
                is_german = file_name.endswith('_de.yml')
                corpus = load_corpus_from_file(file_name)
                corpus = edit_corpus(corpus, is_german=is_german)
                save_corpus_to_file(corpus, file_name)
                if not is_german:
                    german_file_name = ensure_german_file_exists(file_name)
                    if german_file_name:
                        create_and_save_german_corpus(corpus, file_name.split('.yml')[0])
                    else:
                        print(f"No German file found for {file_name}. Creating a new German file.")
                        create_and_save_german_corpus(corpus, file_name.split('.yml')[0])

            elif choice == 'new':
                print("Creating English corpus...")
                english_corpus = create_corpus()
                base_file_name = get_user_input("Enter the base file name for the corpus: ")

                english_file_name = base_file_name + '.yml'
                save_corpus_to_file(english_corpus, english_file_name)
                print("Creating German corpus...")
                create_and_save_german_corpus(english_corpus, base_file_name)

        else:
            print("No YAML files found. Creating a new corpus.")
            print("Creating English corpus...")
            english_corpus = create_corpus()
            base_file_name = get_user_input("Enter the base file name for the corpus: ")

            english_file_name = base_file_name + '.yml'
            save_corpus_to_file(english_corpus, english_file_name)
            print("Creating German corpus...")
            create_and_save_german_corpus(english_corpus, base_file_name)

        another = get_user_input("Do you want to create or edit another corpus? (yes/no): ").lower()
        if another != 'yes':
            break

if __name__ == "__main__":
    main()
