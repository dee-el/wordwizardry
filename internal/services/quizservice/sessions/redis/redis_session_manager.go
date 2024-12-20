package redissessionmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"wordwizardry/internal/pkg/models"
	"wordwizardry/internal/services/quizservice/sessions"
)

const (
	// Redis key patterns
	sessionKey     = "quiz:session:%s"         // Hash: Stores quiz session data
	playersKey     = "quiz:session:%s:players" // Hash: Stores players in a quiz
	quizIDKey      = "quiz:index:quizid:%s"    // String: Stores session ID for a quiz ID
	leaderboardKey = "quiz:session:%s:scores"  // Sorted Set: Stores scores for ranking
	sessionTTL     = 24 * time.Hour            // TTL for all keys
)

type RedisSessionManager struct {
	rdb *redis.Client
}

var _ sessions.SessionManager = (*RedisSessionManager)(nil)

func NewRedisSessionManager(redisURL string) (*RedisSessionManager, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(opt)

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisSessionManager{rdb: rdb}, nil
}

func (r *RedisSessionManager) FindQuizSession(ctx context.Context, sessionID string) (*models.Session, error) {
	sessionData, err := r.rdb.HGetAll(ctx, fmt.Sprintf(sessionKey, sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if len(sessionData) == 0 {
		return nil, nil
	}

	var session models.Session
	if err := json.Unmarshal([]byte(sessionData["data"]), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	players, err := r.getSessionPlayers(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	session.Players = players

	return &session, nil
}

func (r *RedisSessionManager) FindQuizSessionByQuizID(ctx context.Context, quizID string) (*models.Session, error) {
	key := fmt.Sprintf(quizIDKey, quizID)
	sessionID, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	return r.FindQuizSession(ctx, sessionID)
}

func (r *RedisSessionManager) CreateQuizSession(ctx context.Context, session *models.Session) error {
	session.ID = uuid.New().String()

	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store session data
	key := fmt.Sprintf(sessionKey, session.ID)
	err = r.rdb.HSet(ctx, key, "data", sessionData).Err()
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Create index from quiz ID to session ID
	indexKey := fmt.Sprintf(quizIDKey, session.Quiz.ID)
	err = r.rdb.Set(ctx, indexKey, session.ID, sessionTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to create quiz ID index: %w", err)
	}

	// Set TTL
	r.rdb.Expire(ctx, key, sessionTTL)
	r.rdb.Expire(ctx, indexKey, sessionTTL)
	r.rdb.Expire(ctx, fmt.Sprintf(playersKey, session.ID), sessionTTL)
	r.rdb.Expire(ctx, fmt.Sprintf(leaderboardKey, session.ID), sessionTTL)

	return nil
}

func (r *RedisSessionManager) AddPlayerToQuizSession(ctx context.Context, sessionID string, player models.SessionPlayer) error {
	playerData, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("failed to marshal player: %w", err)
	}

	// Add to players hash
	playersKey := fmt.Sprintf(playersKey, sessionID)
	err = r.rdb.HSet(ctx, playersKey, player.ID, playerData).Err()
	if err != nil {
		return fmt.Errorf("failed to add player: %w", err)
	}

	// Initialize score in leaderboard
	leaderboardKey := fmt.Sprintf(leaderboardKey, sessionID)
	err = r.rdb.ZAdd(ctx, leaderboardKey, redis.Z{
		Score:  0,
		Member: player.ID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to initialize player score: %w", err)
	}

	return nil
}

func (r *RedisSessionManager) FindQuizPlayerSession(ctx context.Context, sessionID, playerID string) (*models.Session, error) {
	session, err := r.FindQuizSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	playerExists := false
	for _, p := range session.Players {
		if p.ID == playerID {
			playerExists = true
			break
		}
	}

	if !playerExists {
		return nil, nil
	}

	return session, nil
}

func (r *RedisSessionManager) UpdateQuizPlayerScoreSession(ctx context.Context, sessionID, playerID string, score int, res models.Result) error {
	leaderboardKey := fmt.Sprintf(leaderboardKey, sessionID)
	err := r.rdb.ZIncrBy(ctx, leaderboardKey, float64(score), playerID).Err()
	if err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}

	playersKey := fmt.Sprintf(playersKey, sessionID)
	playerData, err := r.rdb.HGet(ctx, playersKey, playerID).Result()
	if err != nil {
		return fmt.Errorf("failed to get player data: %w", err)
	}

	var player models.SessionPlayer
	if err := json.Unmarshal([]byte(playerData), &player); err != nil {
		return fmt.Errorf("failed to unmarshal player: %w", err)
	}

	player.Score += score

	updatedPlayerData, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("failed to marshal updated player: %w", err)
	}

	err = r.rdb.HSet(ctx, playersKey, playerID, updatedPlayerData).Err()
	if err != nil {
		return fmt.Errorf("failed to save updated player: %w", err)
	}

	return nil
}

func (r *RedisSessionManager) FindLeaderboardQuizSession(ctx context.Context, sessionID string) ([]models.SessionPlayer, error) {
	leaderboardKey := fmt.Sprintf(leaderboardKey, sessionID)
	playersWithScores, err := r.rdb.ZRevRangeWithScores(ctx, leaderboardKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	playersKey := fmt.Sprintf(playersKey, sessionID)
	players := make([]models.SessionPlayer, 0, len(playersWithScores))

	for _, playerScore := range playersWithScores {
		playerID := playerScore.Member.(string)
		playerData, err := r.rdb.HGet(ctx, playersKey, playerID).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get player data: %w", err)
		}

		var player models.SessionPlayer
		if err := json.Unmarshal([]byte(playerData), &player); err != nil {
			return nil, fmt.Errorf("failed to unmarshal player: %w", err)
		}

		players = append(players, player)
	}

	return players, nil
}

func (r *RedisSessionManager) getSessionPlayers(ctx context.Context, sessionID string) ([]models.SessionPlayer, error) {
	playersData, err := r.rdb.HGetAll(ctx, fmt.Sprintf(playersKey, sessionID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	players := make([]models.SessionPlayer, 0, len(playersData))
	for _, data := range playersData {
		var player models.SessionPlayer
		if err := json.Unmarshal([]byte(data), &player); err != nil {
			return nil, fmt.Errorf("failed to unmarshal player: %w", err)
		}
		players = append(players, player)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Score > players[j].Score
	})

	return players, nil
}
