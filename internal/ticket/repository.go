package ticket

import (
	"context"

	"github.com/matheusantiquera/minhas-rifas/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error)
	List(ctx context.Context, userID string, filters ListFilters) ([]domain.Ticket, error)
	CountByRaffle(ctx context.Context, raffleID int) (int64, error)
	DeleteByUser(ctx context.Context, userID string) (int64, error)
}

type repository struct {
	collection *mongo.Collection
	counters   *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	coll := db.Collection("tickets")

	coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})

	return &repository{
		collection: coll,
		counters:   db.Collection("counters"),
	}
}

func (r *repository) nextID(ctx context.Context) (int, error) {
	filter := bson.M{"_id": "tickets"}
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

func (r *repository) Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	id, err := r.nextID(ctx)
	if err != nil {
		return domain.Ticket{}, err
	}

	ticket.ID = id

	_, err = r.collection.InsertOne(ctx, ticket)
	if err != nil {
		return domain.Ticket{}, err
	}

	return ticket, nil
}

func (r *repository) List(ctx context.Context, userID string, filters ListFilters) ([]domain.Ticket, error) {
	query := bson.M{"user_id": userID}
	if filters.RaffleID > 0 {
		query["raffle_id"] = filters.RaffleID
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	tickets := []domain.Ticket{}
	if err := cursor.All(ctx, &tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *repository) CountByRaffle(ctx context.Context, raffleID int) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"raffle_id": raffleID})
}

func (r *repository) DeleteByUser(ctx context.Context, userID string) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
