package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"golangChatBot/bot"
	"golangChatBot/bot/adapters/storage"
)

type Config struct {
	DictFile               string `yaml:"dict_file"`
	IdfFile                string `yaml:"idf_file"`
	StopWordsFile          string `yaml:"stop_words_file"`
	GeneratedStopWordsFile string `yaml:"generated_stop_words_file"`
}

var (
	configFile    = flag.String("config", "/app/cli/config.yaml", "path to the config file")
	dir           = flag.String("d", "", "the directory to look for corpora files")
	corpora       = flag.String("i", "", "the corpora files, comma to separate multiple files")
	storeFile     = flag.String("o", "corpus.gob", "the file to store corpora")
	printMemStats = flag.Bool("m", false, "enable printing memory stats")
	logFile       = flag.String("log", "train.log", "the file to write logs to")
	extensions    = flag.String("ext", "json,yml,yaml", "file extensions to look for, separated by commas")
)

// for handling special Characters in Conversation areas
func preprocessYAMLFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	inConversations := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "conversations:") {
			inConversations = true
			continue
		}
		if inConversations {
			if strings.HasPrefix(trimmed, "-") {
				index := strings.Index(line, "- ")
				if index >= 0 {
					content := line[index+2:]
					content = strings.TrimSpace(content)
					if strings.HasPrefix(content, "- ") {
						continue
					}
					if !strings.HasPrefix(content, "\"") && strings.Contains(content, ":") {
						content = "\"" + content + "\""
						line = line[:index+2] + content
						lines[i] = line
					}
				}
			} else if len(trimmed) == 0 {
				inConversations = false
			}
		}
	}

	newData := strings.Join(lines, "\n")
	return newData, nil
}

func main() {
	flag.Parse()

	var config Config
	configData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Error reading config file %s: %v", *configFile, err)
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Error parsing config file %s: %v", *configFile, err)
	}

	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)
	exts := strings.Split(*extensions, ",")
	for i, ext := range exts {
		exts[i] = strings.TrimPrefix(ext, ".")
		exts[i] = strings.ToLower(exts[i])
	}

	log.Printf("Config File: %s", *configFile)
	log.Printf("Directory: %s", *dir)
	log.Printf("Corpora: %s", *corpora)
	log.Printf("Store File: %s", *storeFile)
	log.Printf("Print Memory Stats: %v", *printMemStats)
	log.Printf("Log File: %s", *logFile)
	log.Printf("Extensions: %v", exts)

	var corporaFiles []string
	if len(*dir) > 0 {
		files := findCorporaFiles(*dir, exts)
		corporaFiles = append(corporaFiles, files...)
	}

	if len(*corpora) > 0 {
		corporaFiles = append(corporaFiles, strings.Split(*corpora, ",")...)
	}

	if len(corporaFiles) == 0 {
		flag.Usage()
		return
	}

	log.Printf("Training on corpora files: %v", corporaFiles)

	store, err := storage.NewSeparatedMemoryStorage(*storeFile, storage.Config{
		DictFile:               config.DictFile,
		IdfFile:                config.IdfFile,
		StopWordsFile:          config.StopWordsFile,
		GeneratedStopWordsFile: config.GeneratedStopWordsFile,
	})
	if err != nil {
		log.Fatal(err)
	}

	chatbot := &bot.ChatBot{
		PrintMemStats:  *printMemStats,
		Trainer:        bot.NewCorpusTrainer(store),
		StorageAdapter: store,
	}

	processedFiles := []string{}
	for _, filename := range corporaFiles {
		newData, err := preprocessYAMLFile(filename)
		if err != nil {
			log.Fatalf("Error preprocessing file %s: %v", filename, err)
		}

		tempFile, err := os.CreateTemp("", "corpus_*.yaml")
		if err != nil {
			log.Fatalf("Error creating temporary file for %s: %v", filename, err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(newData)
		if err != nil {
			log.Fatalf("Error writing to temporary file %s: %v", tempFile.Name(), err)
		}
		tempFile.Close()
		processedFiles = append(processedFiles, tempFile.Name())
	}

	startTime := time.Now()
	if err := chatbot.Train(processedFiles); err != nil {
		log.Fatal(err)
	}

	elapsedTime := time.Since(startTime)
	log.Printf("Training completed successfully in %s.", elapsedTime)
}

func findCorporaFiles(dir string, extensions []string) []string {
	var files []string
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap["."+strings.ToLower(ext)] = true
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v", path, err)
			return nil // Continue walking
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if extMap[ext] {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking the path %s: %v", dir, err)
	}

	log.Printf("Found corpora files: %v", files)
	return files
}
