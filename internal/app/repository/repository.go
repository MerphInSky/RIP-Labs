package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	minioClient "web_backend/internal/app/minioClient"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAllowed    = errors.New("not allowed")
	ErrNoDraft       = errors.New("no draft for this user")
)

type Repository struct {
	db *gorm.DB
	mc *minio.Client
	rd *redis.Client
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	mc, err := minioClient.InitMinio()
	if err != nil {
		return nil, err
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "",
		DB:       0,
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &Repository{
		db: db,
		mc: mc,
		rd: redisClient,
	}, nil
}

func blacklistKeyForToken(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return "blacklist:" + hex.EncodeToString(h[:])
}

func (r *Repository) AddTokenToBlacklist(ctx context.Context, tokenString string, ttl time.Duration, userID string) error {
	if ttl <= 0 {
		return nil
	}
	key := blacklistKeyForToken(tokenString)
	value := "user_id:" + userID
	return r.rd.Set(key, value, ttl).Err()
}

func (r *Repository) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := blacklistKeyForToken(tokenString)
	n, err := r.rd.Exists(key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
