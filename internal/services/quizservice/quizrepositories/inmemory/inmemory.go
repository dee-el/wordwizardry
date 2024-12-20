package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"wordwizardry/internal/services/quizservice/quizrepositories"

	"wordwizardry/internal/pkg/models"
)

type QuizRepository struct {
	mu        sync.RWMutex
	quizzes   map[string]*models.Quiz
	questions map[string][]models.Question
	results   map[string][]*models.QuizResult
}

var _ quizrepositories.QuizReader = (*QuizRepository)(nil)
var _ quizrepositories.QuizWriter = (*QuizRepository)(nil)

func NewQuizRepository() *QuizRepository {
	repo := &QuizRepository{
		quizzes:   make(map[string]*models.Quiz),
		questions: make(map[string][]models.Question),
		results:   make(map[string][]*models.QuizResult),
	}

	// Initialize with sample data
	repo.initSampleData()
	return repo
}

func (r *QuizRepository) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.quizzes[quiz.ID]; exists {
		return fmt.Errorf("quiz already exists: %s", quiz.ID)
	}

	r.quizzes[quiz.ID] = quiz
	return nil
}

func (r *QuizRepository) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.quizzes[quiz.ID]; !exists {
		return fmt.Errorf("quiz not found: %s", quiz.ID)
	}

	r.quizzes[quiz.ID] = quiz
	return nil
}

func (r *QuizRepository) MapQuestions(ctx context.Context, quizID string, questions []models.Question) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.quizzes[quizID]; !exists {
		return fmt.Errorf("quiz not found: %s", quizID)
	}

	r.questions[quizID] = questions
	return nil
}

func (r *QuizRepository) GetQuiz(ctx context.Context, id string) (*models.Quiz, []models.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quiz, exists := r.quizzes[id]
	if !exists {
		return nil, nil, fmt.Errorf("quiz not found: %s", id)
	}

	questions := r.questions[id]
	return quiz, questions, nil
}

func (r *QuizRepository) SaveQuizResult(ctx context.Context, result *models.QuizResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.results[result.QuizID] = append(r.results[result.QuizID], result)
	return nil
}

// Sample data initialization
func (r *QuizRepository) initSampleData() {
	sampleQuizzes := []struct {
		quiz      *models.Quiz
		questions []models.Question
	}{
		{
			quiz: &models.Quiz{
				ID:        "quiz1",
				Title:     "Basic Animals",
				Status:    models.QuizStatusActive,
				CreatedAt: time.Now(),
			},
			questions: []models.Question{
				{
					ID:      "q1_1",
					QuizID:  "quiz1",
					Word:    "Platypus",
					Meaning: "A duck-billed, beaver-tailed, otter-footed, egg-laying mammal",
					Options: []string{"A marsupial", "A monotreme", "A rodent", "A reptile"},
					Correct: "A monotreme",
				},
				{
					ID:      "q1_2",
					QuizID:  "quiz1",
					Word:    "Pangolin",
					Meaning: "A scaly anteater that rolls into a ball when threatened",
					Options: []string{"A reptile", "A mammal", "An amphibian", "An insect"},
					Correct: "A mammal",
				},
			},
		},
		{
			quiz: &models.Quiz{
				ID:        "quiz2",
				Title:     "Marine Life",
				Status:    models.QuizStatusActive,
				CreatedAt: time.Now(),
			},
			questions: []models.Question{
				{
					ID:      "q2_1",
					QuizID:  "quiz2",
					Word:    "Nautilus",
					Meaning: "A living fossil with a spiral shell and up to 90 tentacles",
					Options: []string{"A cephalopod", "A crustacean", "A fish", "A mollusk"},
					Correct: "A cephalopod",
				},
				{
					ID:      "q2_2",
					QuizID:  "quiz2",
					Word:    "Dugong",
					Meaning: "A marine mammal known as sea cow",
					Options: []string{"A cetacean", "A sirenian", "A pinniped", "A fish"},
					Correct: "A sirenian",
				},
			},
		},
		{
			quiz: &models.Quiz{
				ID:        "quiz3",
				Title:     "Birds of Prey",
				Status:    models.QuizStatusActive,
				CreatedAt: time.Now(),
			},
			questions: []models.Question{
				{
					ID:      "q3_1",
					QuizID:  "quiz3",
					Word:    "Harpy Eagle",
					Meaning: "One of the most powerful eagles, native to rainforests",
					Options: []string{"A falcon", "An eagle", "A hawk", "An owl"},
					Correct: "An eagle",
				},
				{
					ID:      "q3_2",
					QuizID:  "quiz3",
					Word:    "Secretary Bird",
					Meaning: "A bird of prey that walks on long legs and kills snakes",
					Options: []string{"A stork", "A crane", "A raptor", "An ostrich"},
					Correct: "A raptor",
				},
			},
		},
		{
			quiz: &models.Quiz{
				ID:        "quiz4",
				Title:     "Nocturnal Animals",
				Status:    models.QuizStatusActive,
				CreatedAt: time.Now(),
			},
			questions: []models.Question{
				{
					ID:      "q4_1",
					QuizID:  "quiz4",
					Word:    "Aye-aye",
					Meaning: "A nocturnal primate with an unusually long middle finger",
					Options: []string{"A lemur", "A monkey", "A bat", "A sloth"},
					Correct: "A lemur",
				},
				{
					ID:      "q4_2",
					QuizID:  "quiz4",
					Word:    "Tarsier",
					Meaning: "A small primate with enormous eyes",
					Options: []string{"A rodent", "A primate", "A marsupial", "A bat"},
					Correct: "A primate",
				},
			},
		},
		{
			quiz: &models.Quiz{
				ID:        "quiz5",
				Title:     "Unusual Mammals",
				Status:    models.QuizStatusActive,
				CreatedAt: time.Now(),
			},
			questions: []models.Question{
				{
					ID:      "q5_1",
					QuizID:  "quiz5",
					Word:    "Numbat",
					Meaning: "A striped marsupial anteater",
					Options: []string{"A rodent", "A marsupial", "A monotreme", "A carnivore"},
					Correct: "A marsupial",
				},
				{
					ID:      "q5_2",
					QuizID:  "quiz5",
					Word:    "Binturong",
					Meaning: "Also known as bearcat, smells like popcorn",
					Options: []string{"A bear", "A cat", "A viverrid", "A raccoon"},
					Correct: "A viverrid",
				},
			},
		},
	}

	// Store sample data
	for _, sample := range sampleQuizzes {
		r.quizzes[sample.quiz.ID] = sample.quiz
		r.questions[sample.quiz.ID] = sample.questions
	}
}
