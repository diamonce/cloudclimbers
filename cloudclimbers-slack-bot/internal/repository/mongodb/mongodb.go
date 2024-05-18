package mongodb

import (
    "context"
    "cloudclimbers-slack-bot/internal/models"
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

func (r *MongoDBRepository) GetEnabledPlugins() ([]models.PluginConfig, error) {
    cursor, err := r.db.Collection("plugins").Find(context.Background(), bson.M{"enabled": true})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    var plugins []models.PluginConfig
    if err := cursor.All(context.Background(), &plugins); err != nil {
        return nil, err
    }
    return plugins, nil
}

func (r *MongoDBRepository) SetPluginStatus(name string, enabled bool) error {
    _, err := r.db.Collection("plugins").UpdateOne(
        context.Background(),
        bson.M{"name": name},
        bson.M{"$set": bson.M{"enabled": enabled}},
    )
    return err
}
