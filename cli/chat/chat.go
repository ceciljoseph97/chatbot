package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

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

var chatbot *bot.ChatBot

var (
	configFile = flag.String("config", "/app/cli/config.yaml", "path to the config file")
	dev        = flag.Bool("dev", false, "developer mode")
	storeFile  = flag.String("c", "PMFuncOverView.gob", "the file to store corpora")
	tops       = flag.Int("t", 1, "the number of answers to return")
	showIntro  = flag.Bool("intro", true, "show the intro message")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
	httpPort   = flag.String("http", "", "start HTTP server on `port` for network profiling")
	context    = flag.Bool("context", true, "enable or disable context handling")
	cmem       = flag.Int("cmem", 2, "number of conversations the context remains active (2-4)")
	anim       = flag.Bool("anim", false, "enable or disable animated letter-by-letter printing")
)

type Conversation struct {
	Categories    []string   `yaml:"categories"`
	Conversations [][]string `yaml:"conversations"`
}

var greetings []string
var keywords []string

func main() {
	flag.Parse()

	if *cmem < 2 || *cmem > 4 {
		fmt.Println("Invalid cmem value. Please set it between 2 and 4.")
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *httpPort != "" {
		go func() {
			log.Println("Starting pprof server on port " + *httpPort)
			log.Println(http.ListenAndServe(":"+*httpPort, nil))
		}()
	}

	var config Config
	configData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Error reading config file %s: %v", *configFile, err)
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Error parsing config file %s: %v", *configFile, err)
	}
	err = nlp.Initialize(nlp.Config{
		CustomDictionaryFile: config.CustomDictionaryFile,
		VocabularyFile:       config.VocabularyFile,
		WordFrequencyFile:    config.WordFrequencyFile,
	})
	if err != nil {
		log.Fatalf("Failed to initialize NLP: %v", err)
	}
	greetings = loadGreetings(config.GreetingsFile)
	if *dev {
		fmt.Printf("Loaded greetings: %v\n", greetings)
	}

	keywords = loadKeywords(config.VocabularyFile)
	if *dev {
		fmt.Printf("Loaded keywords: %v\n", keywords)
	}

	var wg sync.WaitGroup
	modelLoaded := make(chan bool)

	wg.Add(1)
	go loadModel(&wg, modelLoaded, config)

	go showLoading(modelLoaded)
	wg.Wait()

	scanner := bufio.NewScanner(os.Stdin)
	var conversationData Conversation
	contextCategories := make(map[string]int)

	if *showIntro {
		printIntro(*dev)
	}

	for {
		fmt.Print("User: ")
		if !scanner.Scan() {
			fmt.Println()
			if err := scanner.Err(); err != nil {
				log.Printf("Error reading input: %v", err)
			}
			fmt.Println("Exiting chat.")
			break
		}
		question := scanner.Text()

		if *dev {
			if question == "/exit" {
				break
			} else if question == "/save" {
				saveConversation(&conversationData)
				conversationData = Conversation{}
				contextCategories = make(map[string]int)
				fmt.Println("Conversation saved and data cleared.")
				continue
			}
		}

		// WildCard Exit
		if question == "/geronimo" {
			break
		}

		correctedQuestion := nlp.CorrectInput(question)
		if *dev && correctedQuestion != question {
			fmt.Printf("Corrected Input: %s\n", correctedQuestion)
		}

		startTime := time.Now()

		isGreeting, greetingResponse := handleGreetingsAndOneWordQuestions(correctedQuestion)
		if isGreeting {
			fmt.Print("PeriChat: ")
			typeOutText(greetingResponse)
			continue
		}

		if *context {
			extractCategoriesForContext(correctedQuestion, contextCategories)
		}

		extractCategoriesForSaving(correctedQuestion, &conversationData)

		var contextString string
		if *context && len(contextCategories) > 0 {
			categories := make([]string, 0, len(contextCategories))
			for category, age := range contextCategories {
				if age > 0 {
					categories = append(categories, category)
				}
			}
			if len(categories) > 0 {
				contextString = strings.Join(categories, ", ")
			}
		}

		if *dev && *context {
			fmt.Printf("Current context: %s\n", contextString)
		}

		var questionToAsk string
		if *context && contextString != "" {
			questionToAsk = fmt.Sprintf("%s [Context: %s]", correctedQuestion, contextString)
		} else {
			questionToAsk = correctedQuestion
		}

		if *dev {
			fmt.Printf("Question to ask: %s\n", questionToAsk)
		}

		answers := chatbot.GetResponse(questionToAsk)
		var answerContent string
		if len(answers) == 0 {
			fmt.Println("PeriChat: Hi There, No answer found at the moment... We will update the developers regarding the question asked...")
			answerContent = "No answer!"
		} else {
			if *tops == 1 {
				answerContent = answers[0].Content
				fmt.Print("PeriChat: ")
				typeOutText(answerContent)
				if *dev {
					fmt.Printf("\nConfidence: %.3f\tTime: %s", answers[0].Confidence, time.Since(startTime))
				}
				fmt.Println()
			} else {
				for i, answer := range answers {
					fmt.Printf("%d: ", i+1)
					typeOutText(answer.Content)
					if *dev {
						fmt.Printf("\nConfidence: %.3f\tTime: %s", answer.Confidence, time.Since(startTime))
					}
					fmt.Println()
				}
				answerContent = answers[0].Content
			}
		}

		conversationData.Conversations = append(conversationData.Conversations, []string{correctedQuestion, answerContent})

		extractCategoriesForSaving(answerContent, &conversationData)

		if *context {
			updateCategoryAges(contextCategories)
		}

		if *dev {
			fmt.Println("Time taken:", time.Since(startTime))
		}
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func loadModel(wg *sync.WaitGroup, modelLoaded chan bool, config Config) {
	defer wg.Done()

	store, err := storage.NewSeparatedMemoryStorage(*storeFile, storage.Config{
		DictFile:               config.DictFile,
		IdfFile:                config.IdfFile,
		StopWordsFile:          config.StopWordsFile,
		GeneratedStopWordsFile: config.GeneratedStopWordsFile,
	})
	if err != nil {
		log.Fatal(err)
	}
	chatbot = &bot.ChatBot{
		LogicAdapter: logic.NewTopicMatch(store, *tops),
	}
	if *dev {
		chatbot.LogicAdapter.SetVerbose()
	}
	modelLoaded <- true
	close(modelLoaded)
}

func showLoading(modelLoaded chan bool) {
	spinner := []string{"|", "/", "-", "\\"}
	i := 0
	for {
		select {
		case <-modelLoaded:
			fmt.Println("\nModel loaded successfully!")
			return
		default:
			fmt.Printf("\rLoading model... %s", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func extractCategoriesForContext(text string, contextCategories map[string]int) {
	textLower := strings.ToLower(text)

	if *dev {
		fmt.Printf("Analyzing text for context: %s\n", text)
	}
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			if *dev {
				if _, exists := contextCategories[keyword]; exists {
					fmt.Printf("Context category '%s' already exists. Resetting age to %d.\n", keyword, *cmem)
				} else {
					fmt.Printf("Adding new context category '%s' with age %d.\n", keyword, *cmem)
				}
			}
			contextCategories[keyword] = *cmem
		}
	}

	if len(contextCategories) == 0 && *dev {
		fmt.Println("No keywords matched for context in this text.")
	}
}

func extractCategoriesForSaving(text string, conversationData *Conversation) {
	textLower := strings.ToLower(text)
	if *dev {
		fmt.Printf("Analyzing text for saving categories: %s\n", text)
	}
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			if !contains(conversationData.Categories, keyword) {
				conversationData.Categories = append(conversationData.Categories, keyword)
				if *dev {
					fmt.Printf("Added category '%s' to conversation data.\n", keyword)
				}
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, str := range slice {
		if str == item {
			return true
		}
	}
	return false
}

func saveConversation(conversationData *Conversation) {
	if len(conversationData.Conversations) == 0 {
		fmt.Println("No conversations to save.")
		return
	}

	dir := "recent_chats"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			log.Fatalf("Error creating directory %s: %v", dir, err)
		}
	}
	category := "PerinetGenericConversation"
	if len(conversationData.Categories) > 0 {
		category = conversationData.Categories[0]
	}

	timestamp := time.Now().Format("20060102T150405")
	filename := fmt.Sprintf("perichat%s_%s.yml", timestamp, category)
	filepath := filepath.Join(dir, filename)

	data, err := yaml.Marshal(&conversationData)
	if err != nil {
		log.Fatalf("Error marshaling YAML: %v", err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		log.Fatalf("Error writing file %s: %v", filepath, err)
	}

	if *dev {
		fmt.Printf("Conversation saved to %s\n", filepath)
	}
}

func typeOutText(text string) {
	if *anim {
		rand.Seed(time.Now().UnixNano())
		for _, char := range text {
			fmt.Printf("%c", char)
			sleepDuration := time.Duration(rand.Intn(31)+20) * time.Millisecond
			time.Sleep(sleepDuration)
		}
		fmt.Println()
	} else {
		fmt.Println(text)
	}
}

func printIntro(devMode bool) {
	intro := `
***************************************
*                                     *
*        Welcome to PeriChat!         *
*                                     *
***************************************

Our purpose is to provide easy answers to your questions about Perinet products, APIs, and any general Perinet-related queries.
`

	if devMode {
		intro += "Type '/exit' to end the session, '/save' to save the conversation and '/geronimo' for special exit.\n"
	} else {
		intro += "Type '/geronimo' for special exit.\n"
	}

	fmt.Println(intro)
}

func updateCategoryAges(contextCategories map[string]int) {
	for category := range contextCategories {
		if contextCategories[category] > 0 {
			contextCategories[category]--
			if *dev {
				fmt.Printf("Decreased age of context category '%s' to %d.\n", category, contextCategories[category])
			}
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
		log.Printf("Keywords file '%s' is empty. No keywords will be used for context.", filename)
	}

	return keywordsList
}

func handleGreetingsAndOneWordQuestions(question string) (bool, string) {
	words := strings.Fields(question)
	wordCount := len(words)
	wordLower := strings.ToLower(strings.TrimSpace(question))

	if contains(greetings, wordLower) {
		return true, "Hi There please ask me more about the Perinet products"
	}

	if wordCount == 1 {
		return true, fmt.Sprintf("%s, can you provide more context to this", words[0])
	}

	return false, ""
}
