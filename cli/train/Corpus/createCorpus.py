import os
import yaml

def get_user_input(prompt):
    return input(prompt)

def list_yml_files():
    return [f for f in os.listdir() if f.endswith('.yml')]

def load_corpus_from_file(file_name):
    with open(file_name, 'r') as file:
        return yaml.safe_load(file)

def edit_corpus(corpus):
    print("Current categories: ", corpus['categories'])
    add_categories = get_user_input("Do you want to add new categories? (yes/no): ").lower()
    if add_categories == 'yes':
        print("Enter new categories (type 'done' to finish):")
        while True:
            category = get_user_input("Category: ")
            if category.lower() == 'done':
                break
            corpus['categories'].append(category)

    print("Current conversations: ", corpus['conversations'])
    add_conversations = get_user_input("Do you want to add new conversations? (yes/no): ").lower()
    if add_conversations == 'yes':
        print("Enter new conversations (type 'done' to finish):")
        while True:
            question = get_user_input("Question: ")
            if question.lower() == 'done':
                break
            answer = get_user_input("Answer: ")
            corpus['conversations'].append([question, answer])

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
                corpus = load_corpus_from_file(file_name)
                corpus = edit_corpus(corpus)
                save_corpus_to_file(corpus, file_name)
            elif choice == 'new':
                corpus = create_corpus()
                base_file_name = get_user_input("Enter the base file name for the corpus: ")
                english_file_name = base_file_name + '.yml'
                save_corpus_to_file(corpus, english_file_name)
        else:
            print("No YAML files found. Creating a new corpus.")
            print("Creating English corpus...")
            corpus = create_corpus()
            base_file_name = get_user_input("Enter the base file name for the corpus: ")
            english_file_name = base_file_name + '.yml'
            save_corpus_to_file(corpus, english_file_name)
        
        another = get_user_input("Do you want to create or edit another corpus? (yes/no): ").lower()
        if another != 'yes':
            break

if __name__ == "__main__":
    main()
