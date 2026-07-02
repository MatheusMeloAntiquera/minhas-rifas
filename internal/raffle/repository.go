package raffle

import (
	"context"

	"github.com/matheusantiquera/minhas-rifas/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Create(ctx context.Context, raffle domain.Raffle) (domain.Raffle, error)
	FindByID(ctx context.Context, id int) (*domain.Raffle, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Raffle, error)
	DeleteByUser(ctx context.Context, userID string) (int64, error)
}

type repository struct {
	collection *mongo.Collection
	counters   *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	coll := db.Collection("raffles")

	coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})

	return &repository{
		collection: coll,
		counters:   db.Collection("counters"),
	}
}

func (r *repository) nextID(ctx context.Context) (int, error) {
	filter := bson.M{"_id": "raffles"}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}

	err := r.counters.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.Seq, nil
}

func (r *repository) Create(ctx context.Context, raffle domain.Raffle) (domain.Raffle, error) {
	id, err := r.nextID(ctx)
	if err != nil {
		return domain.Raffle{}, err
	}

	raffle.ID = id

	_, err = r.collection.InsertOne(ctx, raffle)
	if err != nil {
		return domain.Raffle{}, err
	}

	return raffle, nil
}

func (r *repository) FindByID(ctx context.Context, id int) (*domain.Raffle, error) {
	filter := bson.M{"_id": id}

	var raffle domain.Raffle
	err := r.collection.FindOne(ctx, filter).Decode(&raffle)
	if err != nil {
		return nil, err
	}

	return &raffle, nil
}

func (r *repository) ListByUser(ctx context.Context, userID string) ([]domain.Raffle, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	raffles := []domain.Raffle{}
	if err := cursor.All(ctx, &raffles); err != nil {
		return nil, err
	}

	return raffles, nil
}

func (r *repository) DeleteByUser(ctx context.Context, userID string) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
