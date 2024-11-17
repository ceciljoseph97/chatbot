package perichatbot

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"

	"golangChatBot/bot"
	"golangChatBot/bot/adapters/logic"
	"golangChatBot/bot/adapters/storage"
	"golangChatBot/cli/chat/nlp"
)

type Config struct {
	GreetingsFile          string `yaml:"greetings_file"`
	VocabularyFile         string `yaml:"vocabulary_file"`
	KeywordsFile           string `yaml:"keywords_file"`
	CustomDictionaryFile   string `yaml:"custom_dictionary_file"`
	WordFrequencyFile      string `yaml:"word_frequency_file"`
	DictFile               string `yaml:"dict_file"`
	IdfFile                string `yaml:"idf_file"`
	StopWordsFile          string `yaml:"stop_words_file"`
	GeneratedStopWordsFile string `yaml:"generated_stop_words_file"`
}

type Chatbot struct {
	bot           *bot.ChatBot
	greetings     []string
	keywords      []string
	config        Config
	modelLoaded   bool
	modelLoadOnce sync.Once
	dev           bool
	storeFile     string
	tops          int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewChatbot(configFile string, dev bool, storeFile string, tops int) (*Chatbot, error) {
	cb := &Chatbot{
		dev:       dev,
		storeFile: storeFile,
		tops:      tops,
	}

	if err := cb.loadConfig(configFile); err != nil {
		return nil, err
	}

	if err := cb.initializeNLP(); err != nil {
		return nil, err
	}

	cb.loadGreetings()
	cb.loadKeywords()

	if err := cb.loadModel(); err != nil {
		return nil, err
	}

	return cb, nil
}

func (cb *Chatbot) loadConfig(configFile string) error {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading config file %s: %v", configFile, err)
	}
	if err := yaml.Unmarshal(configData, &cb.config); err != nil {
		return fmt.Errorf("error parsing config file %s: %v", configFile, err)
	}
	return nil
}

func (cb *Chatbot) initializeNLP() error {
	err := nlp.Initialize(nlp.Config{
		CustomDictionaryFile: cb.config.CustomDictionaryFile,
		VocabularyFile:       cb.config.VocabularyFile,
		WordFrequencyFile:    cb.config.WordFrequencyFile,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize NLP: %v", err)
	}
	return nil
}

func (cb *Chatbot) loadGreetings() {
	cb.greetings = loadGreetings(cb.config.GreetingsFile)
	if cb.dev {
		fmt.Printf("Loaded greetings: %v\n", cb.greetings)
	}
}

func (cb *Chatbot) loadKeywords() {
	cb.keywords = loadKeywords(cb.config.VocabularyFile)
	if cb.dev {
		fmt.Printf("Loaded keywords: %v\n", cb.keywords)
	}
}

func (cb *Chatbot) loadModel() error {
	var err error
	cb.modelLoadOnce.Do(func() {
		store, e := storage.NewSeparatedMemoryStorage(cb.storeFile, storage.Config{
			DictFile:               cb.config.DictFile,
			IdfFile:                cb.config.IdfFile,
			StopWordsFile:          cb.config.StopWordsFile,
			GeneratedStopWordsFile: cb.config.GeneratedStopWordsFile,
		})
		if e != nil {
			err = e
			return
		}
		cb.bot = &bot.ChatBot{
			LogicAdapter: logic.NewTopicMatch(store, cb.tops),
		}
		if cb.dev {
			cb.bot.LogicAdapter.SetVerbose()
		}
		cb.modelLoaded = true
	})
	return err
}

func (cb *Chatbot) GetResponse(message string) string {
	correctedMessage := nlp.CorrectInput(message)
	isGreeting, greetingResponse := cb.handleGreetingsAndOneWordQuestions(correctedMessage)

	if isGreeting {
		return greetingResponse
	}

	answers := cb.bot.GetResponse(correctedMessage)
	if len(answers) == 0 {
		return "Hi there, no answer found at the moment. We'll update the developers regarding the question asked."
	}

	return answers[0].Content
}

func (cb *Chatbot) handleGreetingsAndOneWordQuestions(question string) (bool, string) {
	words := strings.Fields(question)
	wordCount := len(words)
	wordLower := strings.ToLower(strings.TrimSpace(question))

	if contains(cb.greetings, wordLower) {
		return true, "Hi there! Please ask me more about the Perinet products."
	}

	if wordCount == 1 {
		return true, fmt.Sprintf("%s, can you provide more context to this?", words[0])
	}

	return false, ""
}

func (cb *Chatbot) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break
		}
		clientMessage := string(message)
		if cb.dev {
			log.Printf("Received message: %s", clientMessage)
		}

		response := cb.GetResponse(clientMessage)

		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func loadGreetings(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Could not open greetings file '%s': %v. Using default greetings.", filename, err)
		return []string{"hi", "hello", "hey", "greetings", "sup", "yo"}
	}
	defer file.Close()

	var greetingsList []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			greetingsList = append(greetingsList, strings.ToLower(line))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading greetings file '%s': %v. Using default greetings.", filename, err)
		return []string{"hi", "hello", "hey", "greetings", "sup", "yo"}
	}

	if len(greetingsList) == 0 {
		log.Printf("Greetings file '%s' is empty. Using default greetings.", filename)
		return []string{"hi", "hello", "hey", "greetings", "sup", "yo"}
	}

	return greetingsList
}

func loadKeywords(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Could not open keywords file '%s': %v. No keywords will be used for context.", filename, err)
		return []string{}
	}
	defer file.Close()

	var keywordsList []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			keywordsList = append(keywordsList, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading keywords file '%s': %v. No keywords will be used for context.", filename, err)
		return []string{}
	}

	if len(keywordsList) == 0 {
		log.Printf("Keywords file '%s' is empty. No keywords will be used for context.", keywordsList)
	}

	return keywordsList
}

func contains(slice []string, item string) bool {
	for _, str := range slice {
		if str == item {
			return true
		}
	}
	return false
}
