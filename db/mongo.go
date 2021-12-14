package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	options2 "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoStorage struct {
	Instance *mongo.Database
	DBName   string
}

type MongoClusterConfig struct {
	Database []*MongoConnectConfig
	ReplUri  string
	DbName   string
	User     string
	Password string
}

type MongoConnectConfig struct {
	Port string
	Host string
}

func OpenMongoConnect(config *MongoClusterConfig) (*MongoStorage, error) {
	var hosts []string

	s := &MongoStorage{}

	for _, value := range config.Database {
		dsn := fmt.Sprintf("mongodb://%s:%s@%s:%s",
			config.User,
			config.Password,
			value.Host,
			value.Port,
		)

		hosts = append(hosts, dsn)
	}

	options := &options2.ClientOptions{
		Auth: &options2.Credential{
			AuthSource: config.DbName,
			Username:   config.User,
			Password:   config.Password,
		},
		Hosts:      hosts,
		ReplicaSet: &config.ReplUri,
	}

	client, err := mongo.Connect(context.TODO(), options)

	if err != nil {
		return nil, ErrMongoDbConnect(err)
	}

	s.Instance = client.Database(config.DbName)

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, ErrMongoNoPing(err)
	}

	return s, nil
}

func (m *MongoStorage) Close() {
	db := m.Instance.Client()
	_ = db.Disconnect(context.TODO())
}
