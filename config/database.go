package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

// ConnectDB connects to MongoDB using cached configuration
func ConnectDB() {
	// Ensure configuration is loaded first
	if Config == nil {
		log.Fatal("❌ Configuration not loaded. Call LoadEnv() first before ConnectDB()")
	}

	// Get cached MongoDB URI and database name
	mongoURI := GetMongoURI()
	dbName := GetDBName()

	if mongoURI == "" {
		log.Fatal("❌ MongoDB URI is required. Please set MONGO_URI environment variable or configure Secret Manager")
	}

	log.Printf("🔗 Connecting to MongoDB database: %s", dbName)

	// Create MongoDB client options with optimized settings
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Set connection timeouts
	clientOptions.SetConnectTimeout(10 * time.Second)
	clientOptions.SetServerSelectionTimeout(5 * time.Second)
	clientOptions.SetSocketTimeout(30 * time.Second)

	// Set connection pool settings for production
	clientOptions.SetMaxPoolSize(10)
	clientOptions.SetMinPoolSize(2)
	clientOptions.SetMaxConnIdleTime(30 * time.Second)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("❌ Failed to create MongoDB client: %v", err)
	}

	// Ping the database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}

	DB = client
	configMode := "environment variables"
	if IsSecretManagerEnabled() {
		configMode = "Secret Manager (cached)"
	}

	fmt.Printf("✅ Connected to MongoDB database: %s (using %s)\n", dbName, configMode)
	log.Println("🚀 Database connection pool configured and ready")
}

// GetCollection returns a MongoDB collection using cached database name
func GetCollection(collectionName string) *mongo.Collection {
	if DB == nil {
		log.Fatal("❌ Database not connected. Call ConnectDB() first")
	}

	// Use cached database name from configuration
	dbName := GetDBName()
	return DB.Database(dbName).Collection(collectionName)
}

// DisconnectDB closes the MongoDB connection gracefully
func DisconnectDB() {
	if DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := DB.Disconnect(ctx); err != nil {
			log.Printf("⚠️  Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("✅ Disconnected from MongoDB")
		}
		DB = nil
	}
}

// GetDatabase returns the MongoDB database instance using cached name
func GetDatabase() *mongo.Database {
	if DB == nil {
		log.Fatal("❌ Database not connected. Call ConnectDB() first")
	}

	dbName := GetDBName()
	return DB.Database(dbName)
}

// HealthCheckDB performs a quick health check on the database connection
func HealthCheckDB() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return DB.Ping(ctx, nil)
}
