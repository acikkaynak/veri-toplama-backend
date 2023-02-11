package locations

import (
	"context"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type Repository interface {
	GetLocations(ctx context.Context) ([]*LocationDB, error)

	ResolveLocation(ctx context.Context, location *LocationDB) error

	IsResolved(ctx context.Context, locationID int) (bool, error)

	IsDuplicate(ctx context.Context, tweetContents string) (bool, error)
}

type repository struct {
	mongo sources.MongoClient
}

func NewRepository(mongo sources.MongoClient) Repository {
	return &repository{
		mongo: mongo,
	}
}

type Location struct {
	EntryID          int       `json:"entry_id"`
	Loc              []float64 `json:"loc"`
	OriginalMessage  string    `json:"original_message"`
	OriginalLocation string    `json:"original_location"`
}

const (
	TypeWreckage   = 1
	TypeSupplyHelp = 2
)

type LocationDB struct {
	EntryID          int       `json:"entry_id" bson:"entry_id"`
	Location         []float64 `json:"location" bson:"location"`
	Corrected        bool      `json:"corrected" bson:"corrected"`
	Verified         bool      `json:"verified" bson:"verified"`
	OriginalAddress  string    `json:"original_address" bson:"original_address"`
	CorrectedAddress string    `json:"corrected_address" bson:"corrected_address"`
	OpenAddress      string    `json:"open_address" bson:"open_address"`
	Apartment        string    `json:"apartment" bson:"apartment"`
	Type             int       `json:"type" bson:"type"`
	Reason           string    `json:"reason" bson:"reason"`
	TweetContents    string    `json:"tweet_contents" bson:"tweet_contents"`
}

func (r *repository) GetLocations(ctx context.Context) ([]*LocationDB, error) {
	cur, err := r.mongo.Find(ctx, "locations", bson.D{})
	if err != nil {
		return nil, err
	}

	locs := make([]*LocationDB, 0)
	if err := cur.All(ctx, &locs); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return locs, nil
}

func (r *repository) ResolveLocation(ctx context.Context, location *LocationDB) error {
	if err := r.mongo.DeleteOne(ctx, "locations", bson.D{{
		Key:   "entry_id",
		Value: location.EntryID,
	}}); err != nil {
		logrus.Errorln(err)

		return err
	}

	if err := r.mongo.InsertOne(ctx, "locations", location); err != nil {
		logrus.Errorln(err)

		return err
	}

	return nil
}

func (r *repository) IsResolved(ctx context.Context, locationID int) (bool, error) {
	exists, err := r.mongo.DoesExist(ctx, "locations", bson.D{{
		Key:   "entry_id",
		Value: locationID,
	}})
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *repository) IsDuplicate(ctx context.Context, tweetContents string) (bool, error) {
	exists, err := r.mongo.DoesExist(ctx, "locations", bson.D{{
		Key:   "tweet_contents",
		Value: tweetContents,
	}})
	if err != nil {
		return false, err
	}

	return exists, nil
}
