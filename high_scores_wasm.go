//go:build js
// +build js

package main

import (
	"encoding/json"
	"sort"
	"syscall/js"
)

// ─── Web/WASM Storage (localStorage) ───────────────────────────────────────

// saveToLocalStorageImpl saves a high score to browser localStorage
func (hsm *HighScoreManager) saveToLocalStorageImpl(initials string, score int) error {
	// Get existing scores
	scores, _ := hsm.loadFromLocalStorageImpl()

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

	// Serialize and save to localStorage
	data, err := json.Marshal(scores)
	if err != nil {
		return err
	}

	// Use JS to save to localStorage
	localStorage := js.Global().Get("localStorage")
	localStorage.Call("setItem", "ps41_high_scores", string(data))

	return nil
}

// loadFromLocalStorageImpl loads high scores from browser localStorage
func (hsm *HighScoreManager) loadFromLocalStorageImpl() ([]HighScore, error) {
	localStorage := js.Global().Get("localStorage")
	jsonStr := localStorage.Call("getItem", "ps41_high_scores")

	if jsonStr.IsNull() {
		// No scores yet, return prepopulated list
		return getDefaultHighScores(), nil
	}

	var scores []HighScore
	err := json.Unmarshal([]byte(jsonStr.String()), &scores)
	if err != nil {
		// Fall back to defaults if JSON is corrupted
		return getDefaultHighScores(), nil
	}

	return scores, nil
}
