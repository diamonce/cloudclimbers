package mongodb

import (
	"cloudclimbers-slack-bot/internal/models"
	"cloudclimbers-slack-bot/internal/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoDBRepository struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDBRepository(client *mongo.Client, db *mongo.Database) *MongoDBRepository {
	repo := &MongoDBRepository{client: client, db: db}
	repo.setupCollections()
	return repo
}

func (r *MongoDBRepository) setupCollections() {
	collections := []string{"environments", "plugins", "action_logs"}

	for _, collectionName := range collections {
		collection := r.db.Collection(collectionName)
		// Check if the collection exists by counting documents
		_, err := collection.EstimatedDocumentCount(context.Background(), &options.EstimatedDocumentCountOptions{})
		if err != nil {
			// If the collection does not exist, create it
			err = r.db.CreateCollection(context.Background(), collectionName)
			if err != nil {
				utils.Logger().Error("Failed to create collection", zap.String("collection", collectionName), zap.Error(err))
			} else {
				utils.Logger().Info("Created collection", zap.String("collection", collectionName))
			}
		} else {
			utils.Logger().Info("Collection already exists", zap.String("collection", collectionName))
		}
	}
}

func (r *MongoDBRepository) CreateEnvironment(env *models.Environment) error {
	utils.Logger().Info("Inserting environment", zap.String("id", env.ID))
	_, err := r.db.Collection("environments").InsertOne(context.Background(), env)
	if err != nil {
		utils.Logger().Error("Failed to insert environment", zap.Error(err))
	} else {
		utils.Logger().Info("Successfully inserted environment", zap.String("id", env.ID))
	}
	return err
}

func (r *MongoDBRepository) GetEnvironment(id string) (*models.Environment, error) {
	utils.Logger().Info("Getting environment", zap.String("id", id))
	var env models.Environment
	err := r.db.Collection("environments").FindOne(context.Background(), bson.M{"id": id}).Decode(&env)
	if err != nil {
		utils.Logger().Error("Failed to get environment", zap.Error(err))
		return nil, err
	}
	utils.Logger().Info("Successfully got environment", zap.String("id", id))
	return &env, err
}

func (r *MongoDBRepository) DeleteEnvironment(id string) error {
	utils.Logger().Info("Deleting environment", zap.String("id", id))
	_, err := r.db.Collection("environments").DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		utils.Logger().Error("Failed to delete environment", zap.Error(err))
	} else {
		utils.Logger().Info("Successfully deleted environment", zap.String("id", id))
	}
	return err
}

func (r *MongoDBRepository) GetEnabledPlugins() ([]models.PluginConfig, error) {
	utils.Logger().Info("Getting enabled plugins")
	cursor, err := r.db.Collection("plugins").Find(context.Background(), bson.M{"enabled": true})
	if err != nil {
		utils.Logger().Error("Failed to get enabled plugins", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(context.Background())

	var plugins []models.PluginConfig
	if err := cursor.All(context.Background(), &plugins); err != nil {
		utils.Logger().Error("Failed to decode enabled plugins", zap.Error(err))
		return nil, err
	}
	utils.Logger().Info("Successfully got enabled plugins", zap.Int("count", len(plugins)))
	return plugins, nil
}

func (r *MongoDBRepository) SetPluginStatus(name string, enabled bool) error {
	utils.Logger().Info("Setting plugin status", zap.String("name", name), zap.Bool("enabled", enabled))
	_, err := r.db.Collection("plugins").UpdateOne(
		context.Background(),
		bson.M{"name": name},
		bson.M{"$set": bson.M{"enabled": enabled}},
	)
	if err != nil {
		utils.Logger().Error("Failed to set plugin status", zap.Error(err))
	} else {
		utils.Logger().Info("Successfully set plugin status", zap.String("name", name), zap.Bool("enabled", enabled))
	}
	return err
}

func (r *MongoDBRepository) LogAction(actionLog *models.ActionLog) error {
	utils.Logger().Info("Logging action", zap.String("action_id", actionLog.ActionID), zap.String("user_id", actionLog.UserID), zap.String("channel_id", actionLog.ChannelID))
	_, err := r.db.Collection("action_logs").InsertOne(context.Background(), actionLog)
	if err != nil {
		utils.Logger().Error("Failed to log action", zap.Error(err))
	} else {
		utils.Logger().Info("Successfully logged action", zap.String("action_id", actionLog.ActionID), zap.String("user_id", actionLog.UserID), zap.String("channel_id", actionLog.ChannelID))
	}
	return err
}
