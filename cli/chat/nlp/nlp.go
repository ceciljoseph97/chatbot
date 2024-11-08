// primary for Spell Check now. Compartmentalization Purpose
package nlp

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/agnivade/levenshtein"
)

var (
	customDictionary map[string]string
	vocabulary       []string
	wordFrequency    map[string]int
	maxEditDistance  = 3
)

type Config struct {
	CustomDictionaryFile string `yaml:"custom_dictionary_file"`
	VocabularyFile       string `yaml:"vocabulary_file"`
	WordFrequencyFile    string `yaml:"word_frequency_file"`
}

func Initialize(config Config) error {
	err := loadCustomDictionary(config.CustomDictionaryFile)
	if err != nil {
		return fmt.Errorf("error loading custom dictionary: %w", err)
	}

	err = loadVocabulary(config.VocabularyFile)
	if err != nil {
		return fmt.Errorf("error loading vocabulary: %w", err)
	}

	err = loadWordFrequency(config.WordFrequencyFile)
	if err != nil {
		fmt.Printf("Warning: Could not load word frequency data: %v\n", err)
		wordFrequency = make(map[string]int)
	}

	return nil
}

func CorrectInput(input string) string {
	input = correctCustomTerms(input)
	input = correctSpelling(input)
	return input
}

func correctCustomTerms(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if corrected, exists := customDictionary[lowerWord]; exists {
			words[i] = corrected
		}
	}
	return strings.Join(words, " ")
}

func correctSpelling(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if _, exists := customDictionary[lowerWord]; exists {
			continue
		}
		if containsWord(vocabulary, word) {
			continue
		}

		if _, exists := wordFrequency[lowerWord]; exists {
			continue
		}
		suggestions := getSuggestions(lowerWord)
		if len(suggestions) > 0 {
			words[i] = suggestions[0]
		}
	}
	return strings.Join(words, " ")
}

func getSuggestions(word string) []string {
	candidates := []string{}
	minDistance := maxEditDistance + 1
	for _, vocabWord := range vocabulary {
		distance := levenshtein.ComputeDistance(word, strings.ToLower(vocabWord))
		if distance <= maxEditDistance {
			if distance < minDistance {
				minDistance = distance
				candidates = []string{vocabWord}
			} else if distance == minDistance {
				candidates = append(candidates, vocabWord)
			}
		}
	}

	for freqWord := range wordFrequency {
		distance := levenshtein.ComputeDistance(word, freqWord)
		if distance <= maxEditDistance {
			if distance < minDistance {
				minDistance = distance
				candidates = []string{freqWord}
			} else if distance == minDistance {
				candidates = append(candidates, freqWord)
			}
		}
	}

	if len(candidates) > 1 {
		highestFreq := -1
		bestCandidate := ""
		for _, candidate := range candidates {
			freq := getWordFrequency(candidate)
			if freq > highestFreq {
				highestFreq = freq
				bestCandidate = candidate
			}
		}
		return []string{bestCandidate}
	}

	return candidates
}

func getWordFrequency(word string) int {
	if freq, exists := wordFrequency[strings.ToLower(word)]; exists {
		return freq
	}
	return 1
}

func loadCustomDictionary(filePath string) error {
	customDictionary = make(map[string]string)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			continue
		}
		misspelling := strings.TrimSpace(parts[0])
		correctTerm := strings.TrimSpace(parts[1])
		customDictionary[misspelling] = correctTerm
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func loadVocabulary(filePath string) error {
	vocabulary = []string{}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		vocabulary = append(vocabulary, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func loadWordFrequency(filePath string) error {
	wordFrequency = make(map[string]int)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		word := strings.ToLower(parts[0])
		var freq int
		fmt.Sscanf(parts[1], "%d", &freq)
		wordFrequency[word] = freq
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func containsWord(slice []string, word string) bool {
	lowerWord := strings.ToLower(word)
	for _, str := range slice {
		if strings.ToLower(str) == lowerWord {
			return true
		}
	}
	return false
}
