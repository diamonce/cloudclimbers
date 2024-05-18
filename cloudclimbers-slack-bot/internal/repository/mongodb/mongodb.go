package mongodb

import (
    "context"
    "cloudclimbers-slack-bot/internal/models"
    "cloudclimbers-slack-bot/internal/repository"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type MongoDBRepository struct {
    client *mongo.Client
    db     *mongo.Database
}

func NewMongoDBRepository(client *mongo.Client, db *mongo.Database) repository.EnvironmentRepository {
    return &MongoDBRepository{client: client, db: db}
}

func (r *MongoDBRepository) CreateEnvironment(env *models.Environment) error {
    _, err := r.db.Collection("environments").InsertOne(context.Background(), env)
    return err
}

func (r *MongoDBRepository) GetEnvironment(id string) (*models.Environment, error) {
    var env models.Environment
    err := r.db.Collection("environments").FindOne(context.Background(), bson.M{"id": id}).Decode(&env)
    return &env, err
}

func (r *MongoDBRepository) DeleteEnvironment(id string) error {
    _, err := r.db.Collection("environments").DeleteOne(context.Background(), bson.M{"id": id})
    return err
}
