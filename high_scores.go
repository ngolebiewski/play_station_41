package main

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"sort"
)

// HighScore represents a single high score entry
type HighScore struct {
	Initials string `json:"initials"`
	Score    int    `json:"score"`
}

// HighScoreManager handles persistence across web and native platforms
// Web: localStorage (automatic via JS)
// Native: high_scores.json file in current directory
type HighScoreManager struct {
	isWasm    bool
	filePath  string
	inMemory  []HighScore // Cache for desktop
}

// NewHighScoreManager creates a new high score manager
func NewHighScoreManager() *HighScoreManager {
	hsm := &HighScoreManager{
		isWasm:   runtime.GOOS == "js",
		filePath: "high_scores.json",
	}

	// On native, try to load existing scores from file
	if !hsm.isWasm {
		hsm.loadFromFile()
	}

	return hsm
}

// SaveHighScore saves a new high score
func (hsm *HighScoreManager) SaveHighScore(initials string, score int) error {
	if hsm.isWasm {
		return hsm.saveToLocalStorageImpl(initials, score)
	}
	return hsm.saveToFile(initials, score)
}

// LoadHighScores loads the top 7 high scores
func (hsm *HighScoreManager) LoadHighScores() ([]HighScore, error) {
	if hsm.isWasm {
		return hsm.loadFromLocalStorageImpl()
	}
	// Return cached in-memory scores on native
	if len(hsm.inMemory) == 0 {
		return getDefaultHighScores(), nil
	}
	return hsm.inMemory, nil
}

// ─── Native Storage (JSON file) ────────────────────────────────────────────

func (hsm *HighScoreManager) loadFromFile() {
	data, err := os.ReadFile(hsm.filePath)
	if err != nil {
		// File doesn't exist yet, use defaults
		hsm.inMemory = getDefaultHighScores()
		return
	}

	var scores []HighScore
	err = json.Unmarshal(data, &scores)
	if err != nil {
		log.Printf("Error parsing high scores file: %v", err)
		hsm.inMemory = getDefaultHighScores()
		return
	}

	hsm.inMemory = scores
}

func (hsm *HighScoreManager) saveToFile(initials string, score int) error {
	// Get existing scores
	scores := hsm.inMemory
	if len(scores) == 0 {
		scores = getDefaultHighScores()
	}

	// Add new score
	scores = append(scores, HighScore{
		Initials: initials,
		Score:    score,
	})

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Keep top 7
	if len(scores) > 7 {
		scores = scores[:7]
	}

	// Save to file
	data, err := json.MarshalIndent(scores, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(hsm.filePath, data, 0644)
	if err != nil {
		return err
	}

	// Update in-memory cache
	hsm.inMemory = scores
	return nil
}

// getDefaultHighScores returns the prepopulated high score list
func getDefaultHighScores() []HighScore {
	return []HighScore{
		{Initials: "YAG", Score: 7000},
		{Initials: "INK", Score: 6000},
		{Initials: "CCC", Score: 5000},
		{Initials: "DDD", Score: 4000},
		{Initials: "EEE", Score: 3000},
		{Initials: "FFF", Score: 2000},
		{Initials: "GGG", Score: 1000},
	}
}
