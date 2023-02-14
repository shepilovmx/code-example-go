package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Leaderboard struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Year        int                `json:"year" bson:"year"`
	Month       time.Month         `json:"month" bson:"month"`
	TopCreators []TopCreator       `json:"top_creators" bson:"top_creators"`
	TopEvents   []TopEvent         `json:"top_events" bson:"top_events"`
}

type TopCreator struct {
	UserID      primitive.ObjectID `json:"-" bson:"user_id"`
	Earnings    int                `json:"earnings" bson:"earnings"`
	*UserFields `bson:"user_fields,omitempty"`
}

type TopEvent struct {
	WagerID     primitive.ObjectID `json:"-" bson:"wager_id"`
	UserID      primitive.ObjectID `json:"-" bson:"user_id,omitempty"`
	Title       string             `json:"title" bson:"title,omitempty"`
	Earnings    int                `json:"earnings" bson:"earnings"`
	*UserFields `bson:"user_fields,omitempty"`
}

type UserFields struct {
	UsernameToShow string `json:"name" bson:"-"`
	Username       string `json:"-" bson:"username,omitempty"`
	Email          string `json:"-" bson:"email,omitempty"`
	Avatar         string `json:"avatar" bson:"avatar,omitempty"`
}

func (l *Leaderboard) FormatUsername() {
	for i, _ := range l.TopCreators {
		l.TopCreators[i].SetUsernameForShow(l.TopCreators[i].UserID)
	}

	for i, _ := range l.TopEvents {
		l.TopEvents[i].SetUsernameForShow(l.TopEvents[i].UserID)
	}
}

func (uf *UserFields) SetUsernameForShow(userId primitive.ObjectID) {
	if uf.Username != "" {
		uf.UsernameToShow = uf.Username
	} else if uf.Email != "" {
		index := strings.Index(uf.Email, "@")
		emailName := uf.Email[:index]
		charactersNumber := utf8.RuneCountInString(emailName)

		uf.UsernameToShow = strings.Repeat("*", charactersNumber/2) + uf.Email[charactersNumber/2:]
	} else {
		uf.UsernameToShow = strconv.Itoa(int(userId.Timestamp().Unix()))
	}
}
