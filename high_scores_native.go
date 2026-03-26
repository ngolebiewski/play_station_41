//go:build !js
// +build !js

package main

// saveToLocalStorageImpl is a stub for native platforms (not used)
func (hsm *HighScoreManager) saveToLocalStorageImpl(initials string, score int) error {
	return nil
}

// loadFromLocalStorageImpl is a stub for native platforms (not used)
func (hsm *HighScoreManager) loadFromLocalStorageImpl() ([]HighScore, error) {
	return getDefaultHighScores(), nil
}
