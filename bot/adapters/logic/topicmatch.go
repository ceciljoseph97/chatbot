package logic

import (
	"golangChatBot/bot/adapters/storage"
	"golangChatBot/bot/nlp"
	"sort"
	"strings"
)

// TopicScore holds the scoring information for a potential match
type TopicScore struct {
	Question   string
	TextScore  float32
	TopicScore float32
	FinalScore float32
}

// TopicMatch implements the LogicAdapter interface with topic-based matching
type TopicMatch struct {
	verbose   bool
	storage   storage.StorageAdapter
	tops      int
	stopWords map[string]bool
}

// NewTopicMatch creates a new TopicMatch instance
func NewTopicMatch(storage storage.StorageAdapter, tops int) LogicAdapter {
	return &TopicMatch{
		storage:   storage,
		tops:      tops,
		stopWords: initializeStopWords(),
	}
}

// CanProcess implements LogicAdapter interface
func (match *TopicMatch) CanProcess(text string) bool {
	return true
}

// SetVerbose implements LogicAdapter interface
func (match *TopicMatch) SetVerbose() {
	match.verbose = true
}

// Process implements LogicAdapter interface
func (match *TopicMatch) Process(text string) []Answer {
	if responses, ok := match.storage.Find(text); ok {
		return match.processExactMatch(responses)
	}
	return match.processTopicMatch(text)
}

// processExactMatch handles exact matches found in storage
func (match *TopicMatch) processExactMatch(responses map[string]int) []Answer {
	var answers []Answer

	// Find max count for normalization
	maxCount := 0
	for _, count := range responses {
		if count > maxCount {
			maxCount = count
		}
	}

	// Create answers with normalized confidence scores
	for response, count := range responses {
		normalizedConfidence := float32(count) / float32(maxCount)
		answers = append(answers, Answer{
			Content:    response,
			Confidence: normalizedConfidence,
		})
	}

	// Sort by confidence
	sort.Slice(answers, func(i, j int) bool {
		return answers[i].Confidence > answers[j].Confidence
	})

	// Return top N answers
	if len(answers) > match.tops {
		answers = answers[:match.tops]
	}

	return answers
}

// processTopicMatch handles fuzzy matching based on topic similarity
func (match *TopicMatch) processTopicMatch(text string) []Answer {
	// Extract topics from input
	inputTopics := match.extractTopics(text)

	// Get candidate matches
	candidates := match.storage.Search(text)

	// Score and rank candidates
	scores := make([]TopicScore, 0)
	for _, candidate := range candidates {
		candidateTopics := match.extractTopics(candidate)

		textScore := nlp.SimilarityForStrings(text, candidate)
		topicScore := match.calculateTopicSimilarity(inputTopics, candidateTopics)

		// Calculate weighted scores based on different factors
		lengthRatio := float32(min(len(text), len(candidate))) / float32(max(len(text), len(candidate)))
		topicRatio := float32(len(inputTopics)) / float32(max(len(inputTopics), len(candidateTopics)))

		// Weighted combination of scores
		finalScore := (textScore * 0.4) + // Direct text similarity
			(topicScore * 0.3) + // Topic overlap
			(lengthRatio * 0.15) + // Length similarity
			(topicRatio * 0.15) // Topic count similarity

		scores = append(scores, TopicScore{
			Question:   candidate,
			TextScore:  textScore,
			TopicScore: topicScore,
			FinalScore: finalScore,
		})
	}

	// Sort by final score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].FinalScore > scores[j].FinalScore
	})

	return match.convertToAnswers(scores)
}

// extractTopics extracts meaningful topics from text
func (match *TopicMatch) extractTopics(text string) []string {
	// Normalize text
	text = strings.ToLower(text)
	words := strings.Fields(text)

	// Extract meaningful topics
	topics := make([]string, 0)
	for _, word := range words {
		// Skip stop words and very short words
		if !match.stopWords[word] && len(word) > 2 {
			topics = append(topics, word)
		}
	}

	return topics
}

// calculateTopicSimilarity calculates similarity between two sets of topics
func (match *TopicMatch) calculateTopicSimilarity(topics1, topics2 []string) float32 {
	if len(topics1) == 0 || len(topics2) == 0 {
		return 0
	}

	// Create sets for comparison
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, t := range topics1 {
		set1[t] = true
	}
	for _, t := range topics2 {
		set2[t] = true
	}

	// Calculate intersection
	intersection := 0
	for topic := range set1 {
		if set2[topic] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0
	}

	return float32(intersection) / float32(union)
}

// convertToAnswers converts TopicScores to Answers
func (match *TopicMatch) convertToAnswers(scores []TopicScore) []Answer {
	tops := match.tops
	if len(scores) < tops {
		tops = len(scores)
	}

	answers := make([]Answer, 0, tops)
	if len(scores) == 0 {
		return answers
	}

	// Get highest score for normalization
	maxScore := scores[0].FinalScore

	for i := 0; i < tops; i++ {
		if responses, ok := match.storage.Find(scores[i].Question); ok {
			// Find best response by occurrence count
			var bestResponse string
			var maxCount int
			for response, count := range responses {
				if count > maxCount {
					maxCount = count
					bestResponse = response
				}
			}

			if bestResponse != "" {
				// Normalize confidence score between 0 and 1
				normalizedConfidence := scores[i].FinalScore / maxScore
				answers = append(answers, Answer{
					Content:    bestResponse,
					Confidence: normalizedConfidence,
				})
			}
		}
	}

	return answers
}

// initializeStopWords creates the initial set of stop words
func initializeStopWords() map[string]bool {
	return map[string]bool{
		"the": true, "is": true, "at": true, "which": true, "on": true,
		"a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "with": true, "to": true, "for": true, "of": true,
		"what": true, "how": true, "why": true, "when": true, "where": true,
		"who": true, "will": true, "be": true, "do": true, "does": true,
		"can": true, "could": true, "would": true, "should": true, "has": true,
		"have": true, "had": true, "are": true, "was": true, "were": true,
	}
}

// Helper function for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
