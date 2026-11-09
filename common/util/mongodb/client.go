package mongo_driver

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBClient 用于封装 MongoDB 客户端
type MongoDBClient struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDBClient 初始化 MongoDB 客户端
func NewMongoDBClient(uri string, dbName string) (*MongoDBClient, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// 连接到 MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	// 获取数据库
	database := client.Database(dbName)

	return &MongoDBClient{
		client:   client,
		database: database,
	}, nil
}

// Close 关闭 MongoDB 客户端连接
func (m *MongoDBClient) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := m.client.Disconnect(ctx); err != nil {
		log.Fatal("Failed to disconnect from MongoDB:", err)
	}
}

// InsertOne 插入一个文档到指定集合
func (m *MongoDBClient) InsertOne(collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	collection := m.database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindOne 查询指定条件的一个文档
func (m *MongoDBClient) FindOne(collectionName string, filter interface{}) (*mongo.SingleResult, error) {
	collection := m.database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	return result, nil
}

// UpdateOne 更新指定条件的一个文档
func (m *MongoDBClient) UpdateOne(collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	collection := m.database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteOne 删除指定条件的一个文档
func (m *MongoDBClient) DeleteOne(collectionName string, filter interface{}) (*mongo.DeleteResult, error) {
	collection := m.database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *MongoDBClient) Find(collectionName string, filter interface{}, result interface{}) error {
	collection := m.database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return cur.All(ctx, result)
}

func (m *MongoDBClient) FindWithPagination(collectionName string, filter interface{}, result interface{}, pageNum int64, pageSize int64, sort bson.D) (int64, error) {
	collection := m.database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	skip := (pageNum - 1) * pageSize

	// 构建分页选项
	opts := options.Find().
		SetSkip(skip).
		SetLimit(pageSize).
		SetSort(sort)

	// 查询总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	// 查询数据
	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, result); err != nil {
		return 0, err
	}

	return total, nil
}

func (m *MongoDBClient) DeleteAll(collectionName string) (*mongo.DeleteResult, error) {
	collection := m.database.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 空 filter 表示匹配所有文档
	return collection.DeleteMany(ctx, bson.D{})
}
