package repository

import (
	"context"
	"fmt"
	"time"

	"dummypath/entity"
	"dummypath/go/leaderboard"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongodbLeaderboardRepository struct {
	db *mongo.Database
}

func NewMongodbLeaderboardRepository(DB *mongo.Database) leaderboard.Repository {
	return &mongodbLeaderboardRepository{
		db: DB,
	}
}

func (r mongodbLeaderboardRepository) InsertOne(ctx context.Context, lboard entity.Leaderboard) (primitive.ObjectID, error) {
	if lboard.ID.IsZero() {
		lboard.ID = primitive.NewObjectID()
	}

	_, err := r.db.Collection("leaderboard").InsertOne(ctx, lboard)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return lboard.ID, nil
}

func (r mongodbLeaderboardRepository) UpdateTopEventsAndCreators(ctx context.Context, leaderboardID primitive.ObjectID, topCreators []entity.TopCreator, topEvents []entity.TopEvent) error {
	if leaderboardID.IsZero() {
		return errors.Errorf("missing leaderboard id")
	}

	filter := bson.M{"_id": leaderboardID}

	update := bson.M{
		"$set": bson.M{
			"top_creators": topCreators,
			"top_events":   topEvents,
		},
	}

	res, err := r.db.Collection("leaderboard").UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.WithMessagef(err, `r.db.Collection("leaderboard").UpdateOne(ctx, %+v, %+v) failed`, filter, update)
	}

	if res.MatchedCount == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r mongodbLeaderboardRepository) FindOneByYearAndMonth(ctx context.Context, year int, month time.Month) (*entity.Leaderboard, error) {
	lboard := entity.Leaderboard{
		TopCreators: []entity.TopCreator{},
		TopEvents:   []entity.TopEvent{},
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"year":  year,
				"month": month,
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$top_events",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "wagers",
				"localField":   "top_events.wager_id",
				"foreignField": "_id",
				"as":           "top_events.wager",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$top_events.wager",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "top_events.wager.author._id",
				"foreignField": "_id",
				"as":           "top_events.wager.author",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$top_events.wager.author",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
				"year": bson.M{
					"$first": "$year",
				},
				"month": bson.M{
					"$first": "$month",
				},
				"top_creators": bson.M{
					"$first": "$top_creators",
				},
				"top_events": bson.M{
					"$push": bson.M{
						"wager_id": "$top_events.wager_id",
						"user_id":  "$top_events.wager.author._id",
						"title":    "$top_events.wager.title",
						"earnings": "$top_events.earnings",
						"user_fields": bson.M{
							"avatar":   "$top_events.wager.author.avatar",
							"username": "$top_events.wager.author.username",
							"email":    "$top_events.wager.author.email",
						},
					},
				},
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$top_creators",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "top_creators.user_id",
				"foreignField": "_id",
				"as":           "top_creators.user",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$top_creators.user",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
				"year": bson.M{
					"$first": "$year",
				},
				"month": bson.M{
					"$first": "$month",
				},
				"top_events": bson.M{
					"$first": "$top_events",
				},
				"top_creators": bson.M{
					"$push": bson.M{
						"user_id":  "$top_creators.user_id",
						"earnings": "$top_creators.earnings",
						"user_fields": bson.M{
							"avatar":   "$top_creators.user.avatar",
							"username": "$top_creators.user.username",
							"email":    "$top_creators.user.email",
						},
					},
				},
			},
		}
	}

	cursor, err := r.db.Collection("leaderboard").Aggregate(ctx, pipeline)
	if err != nil {
		return &lboard, fmt.Errorf("aggregate failed: %v", err)
	}

	if !cursor.Next(ctx) {
		return &lboard, entity.ErrNotFound
	}

	if err := cursor.Decode(&lboard); err != nil {
		return &lboard, fmt.Errorf("failed to decode result: %v", err)
	}

	return &lboard, nil
}
