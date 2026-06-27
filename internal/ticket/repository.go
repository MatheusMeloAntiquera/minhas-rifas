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
	List(ctx context.Context, userID int, filters ListFilters) ([]domain.Ticket, error)
}

type repository struct {
	collection *mongo.Collection
	counters   *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		collection: db.Collection("tickets"),
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

func (r *repository) List(ctx context.Context, userID int, filters ListFilters) ([]domain.Ticket, error) {
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
