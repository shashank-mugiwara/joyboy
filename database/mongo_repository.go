package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shashank-mugiwara/joyboy/task"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository implements TaskRepository interface using MongoDB
type MongoRepository struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoRepository creates a new MongoRepository instance
func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	return &MongoRepository{
		client:   client,
		database: client.Database(dbName),
	}
}

// getTasksCollection returns the tasks collection
func (r *MongoRepository) getTasksCollection() *mongo.Collection {
	return r.database.Collection("tasks")
}

// getLocalsCollection returns the locals collection
func (r *MongoRepository) getLocalsCollection() *mongo.Collection {
	return r.database.Collection("locals")
}

// SaveTask saves or updates a task using upsert
func (r *MongoRepository) SaveTask(ctx context.Context, t *task.Task) error {
	collection := r.getTasksCollection()

	filter := bson.M{"id": t.ID}
	update := bson.M{"$set": t}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	return nil
}

// SaveLocal saves or updates a local using upsert
func (r *MongoRepository) SaveLocal(ctx context.Context, l *task.Local) error {
	collection := r.getLocalsCollection()

	filter := bson.M{"name": l.Name}
	update := bson.M{"$set": l}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save local: %w", err)
	}

	return nil
}

// FindTasks retrieves all tasks matching the filter
func (r *MongoRepository) FindTasks(ctx context.Context, filter bson.M) ([]task.Task, error) {
	collection := r.getTasksCollection()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []task.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// FindLocals retrieves all locals matching the filter
func (r *MongoRepository) FindLocals(ctx context.Context, filter bson.M) ([]task.Local, error) {
	collection := r.getLocalsCollection()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find locals: %w", err)
	}
	defer cursor.Close(ctx)

	var locals []task.Local
	if err := cursor.All(ctx, &locals); err != nil {
		return nil, fmt.Errorf("failed to decode locals: %w", err)
	}

	return locals, nil
}

// WhereTasks retrieves tasks matching the given conditions
func (r *MongoRepository) WhereTasks(ctx context.Context, conditions map[string]interface{}) ([]task.Task, error) {
	filter := bson.M(conditions)
	return r.FindTasks(ctx, filter)
}

// WhereLocals retrieves locals matching the given conditions
func (r *MongoRepository) WhereLocals(ctx context.Context, conditions map[string]interface{}) ([]task.Local, error) {
	filter := bson.M(conditions)
	return r.FindLocals(ctx, filter)
}

// FirstTask retrieves the first task matching the filter
func (r *MongoRepository) FirstTask(ctx context.Context, filter bson.M) (*task.Task, error) {
	collection := r.getTasksCollection()

	var t task.Task
	err := collection.FindOne(ctx, filter).Decode(&t)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &t, nil
}

// FirstLocal retrieves the first local matching the filter
func (r *MongoRepository) FirstLocal(ctx context.Context, filter bson.M) (*task.Local, error) {
	collection := r.getLocalsCollection()

	var l task.Local
	err := collection.FindOne(ctx, filter).Decode(&l)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find local: %w", err)
	}

	return &l, nil
}

// DeleteTask deletes a task by ID
func (r *MongoRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	collection := r.getTasksCollection()

	filter := bson.M{"id": id}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("task not found with id: %s", id.String())
	}

	return nil
}

// DeleteLocal deletes a local by name
func (r *MongoRepository) DeleteLocal(ctx context.Context, name string) error {
	collection := r.getLocalsCollection()

	filter := bson.M{"name": name}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete local: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("local not found with name: %s", name)
	}

	return nil
}

// AutoMigrate creates indexes on commonly queried fields
// MongoDB is schemaless, so no migration is needed, but we create indexes for performance
func (r *MongoRepository) AutoMigrate(ctx context.Context) error {
	// Create indexes for tasks collection
	tasksCollection := r.getTasksCollection()

	taskIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "state", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "containerId", Value: 1}},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := tasksCollection.Indexes().CreateMany(ctx, taskIndexes)
	if err != nil {
		return fmt.Errorf("failed to create task indexes: %w", err)
	}

	// Create indexes for locals collection
	localsCollection := r.getLocalsCollection()

	localIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "owner", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "containerId", Value: 1}},
		},
	}

	_, err = localsCollection.Indexes().CreateMany(ctx, localIndexes)
	if err != nil {
		return fmt.Errorf("failed to create local indexes: %w", err)
	}

	return nil
}

// GetAllTasks retrieves all tasks from the database
func (r *MongoRepository) GetAllTasks(ctx context.Context) ([]task.Task, error) {
	return r.FindTasks(ctx, bson.M{})
}

// GetAllLocals retrieves all locals from the database
func (r *MongoRepository) GetAllLocals(ctx context.Context) ([]task.Local, error) {
	return r.FindLocals(ctx, bson.M{})
}

// GetTaskByID retrieves a task by its ID
func (r *MongoRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*task.Task, error) {
	return r.FirstTask(ctx, bson.M{"id": id})
}

// GetTasksByState retrieves all tasks with a specific state
func (r *MongoRepository) GetTasksByState(ctx context.Context, state string) ([]task.Task, error) {
	return r.WhereTasks(ctx, map[string]interface{}{"state": state})
}
