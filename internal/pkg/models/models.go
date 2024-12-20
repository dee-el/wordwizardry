package models

import (
	"time"
)

type QuizStatus string

const (
	QuizStatusActive   QuizStatus = "active"
	QuizStatusInActive QuizStatus = "inactive"
)

// Add validation method
func (s QuizStatus) IsValid() bool {
	switch s {
	case QuizStatusActive, QuizStatusInActive:
		return true
	}
	return false
}

// String representation
func (s QuizStatus) String() string {
	return string(s)
}

type Quiz struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    QuizStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type Question struct {
	ID      string   `json:"id"`
	QuizID  string   `json:"quiz_id"`
	Word    string   `json:"word"`
	Meaning string   `json:"meaning"`
	Options []string `json:"options"`
	Correct string   `json:"correct"`
}

type QuizResult struct {
	ID             string    `json:"id"`
	QuizID         string    `json:"quiz_id"`
	PlayerID       string    `json:"player_id"`
	FinalScore     int       `json:"final_score"`
	CompletionTime time.Time `json:"completion_time"`
	Position       int       `json:"position"`
}

type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type SessionPlayer struct {
	Player
	QuizID string `json:"quiz_id"`
	Score  int    `json:"score"`
}

type Session struct {
	ID        string `json:"id"`
	Quiz      *Quiz  `json:"quiz"`
	Questions []Question
	Players   []SessionPlayer
	Result    Result
}

// Result is a map of questionID:PlayerID to the player's answer
type Result map[string]Answer

type Answer struct {
	PlayerChoice  string `json:"player_choice"`
	CorrectAnswer string `json:"correct_answer"`
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
