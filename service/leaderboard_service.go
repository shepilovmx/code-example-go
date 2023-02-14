package service

import (
	"context"
	"time"

	"dummypath/entity"
	"dummypath/go/leaderboard"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type leaderboardService struct {
	leaderboardRepo leaderboard.Repository
}

func NewLeaderboardService(lr leaderboard.Repository) leaderboard.Service {
	return &leaderboardService{
		leaderboardRepo: lr,
	}
}

func (s leaderboardService) CreateOne(ctx context.Context, newLeaderboard entity.Leaderboard) (*entity.Leaderboard, error) {
	insertedID, err := s.leaderboardRepo.InsertOne(ctx, newLeaderboard)
	if err != nil {
		return nil, err
	}

	newLeaderboard.ID = insertedID

	return &newLeaderboard, nil
}

func (s leaderboardService) UpdateTopEventsAndCreators(ctx context.Context, leaderboardID primitive.ObjectID, topCreators []entity.TopCreator, topEvents []entity.TopEvent) error {
	return s.leaderboardRepo.UpdateTopEventsAndCreators(ctx, leaderboardID, topCreators, topEvents)
}

func (s leaderboardService) FindOneByYearAndMonth(ctx context.Context, year int, month time.Month) (*entity.Leaderboard, error) {
	lboard, err := s.leaderboardRepo.FindOneByYearAndMonth(ctx, year, month)
	if err != nil {
		return lboard, err
	}

	lboard.FormatUsername()

	return lboard, nil
}
