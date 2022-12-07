package db

import (
	"context"
	"github.com/richard-on/task-service/config"
	"github.com/richard-on/task-service/internal/model"
	"github.com/richard-on/task-service/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DB struct {
	Db  *mongo.Collection
	Ctx context.Context
	Log logger.Logger
}

func NewDatabase(ctx context.Context, db *mongo.Collection) *DB {
	return &DB{
		Db:  db,
		Ctx: ctx,
		Log: logger.NewLogger(
			config.DefaultWriter,
			config.LogInfo.Level,
			"task-db"),
	}
}

func (db *DB) AddTask(task model.Task) (model.Task, error) {
	res, err := db.Db.InsertOne(db.Ctx, task)
	if err != nil {
		return model.Task{}, err
	}

	db.Log.Debug(res.InsertedID.(primitive.ObjectID).String())

	return task, nil
}

func (db *DB) GetAllTasks(email string) ([]model.Task, error) {
	cursor, err := db.Db.Find(db.Ctx, bson.M{"initiator": email})
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	for cursor.Next(db.Ctx) {
		var task model.Task
		if err = cursor.Decode(&task); err != nil {
			return nil, err
		}
		db.Log.Debug(task.Initiator)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (db *DB) GetTaskById(taskId string) (model.Task, error) {
	id, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		return model.Task{}, err
	}

	var task model.Task
	res := db.Db.FindOne(db.Ctx, bson.M{"_id": id})
	if err = res.Decode(&task); err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func (db *DB) DeleteTask(taskId string) error {
	id, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		return err
	}

	// TODO: check if nothing was deleted
	_, err = db.Db.DeleteOne(db.Ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateTask(task *model.Task) error {
	filter := bson.M{"_id": task.ID}
	update := bson.D{
		{"$set",
			bson.D{{"status", task.Status}, {"next", task.Next}},
		},
	}

	var updatedTask model.Task
	err := db.Db.FindOneAndUpdate(db.Ctx, filter, update).Decode(&updatedTask)
	if err != nil {
		return err
	}

	return nil
}
