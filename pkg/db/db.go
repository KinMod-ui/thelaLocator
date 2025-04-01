package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KinMod-ui/thelaLocator/helper"
	"github.com/KinMod-ui/thelaLocator/proto"
)

func config() *pgxpool.Config {
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	// Your own Database URL
	const DATABASE_URL string = "postgres://postgres:root@localhost:5432/postgres?"

	dbConfig, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		helper.Mylog.Fatal("Failed to create a config, error: ", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	dbConfig.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		helper.Mylog.Println("Before acquiring the connection pool to the database!!")
		return true
	}

	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		helper.Mylog.Println("After releasing the connection pool to the database!!")
		return true
	}

	dbConfig.BeforeClose = func(c *pgx.Conn) {
		helper.Mylog.Println("Closed the connection pool to the database!!")
	}

	return dbConfig
}

type DatabaseClient interface {
	GetUser(ctx context.Context, id string) (*protobuf.User, error)
	CreateUser(ctx context.Context, user *protobuf.CreateUserRequest) (*protobuf.User, error)
}

type databaseClient struct {
	db *pgxpool.Pool
}

// NewDatabaseClient creates a new database client.
func NewDatabaseClient() (DatabaseClient, error) {
	db, err := pgxpool.NewWithConfig(context.Background(), config())
	if err != nil {
		return nil, fmt.Errorf("error connecting to PostgreSQL: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error pinging PostgreSQL: %w", err)
	}
	return &databaseClient{db: db}, nil
}

// GetUser retrieves a user from the database.
func (c *databaseClient) GetUser(ctx context.Context, id string) (*protobuf.User, error) {
	var user protobuf.User
	err := c.db.QueryRow(ctx, "SELECT id, lat, long FROM users WHERE id = $1", id).
		Scan(&user.Id, &user.Lat, &user.Long)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %s not found", id)
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	return &user, nil
}

// CreateUser creates a new user in the database.
func (c *databaseClient) CreateUser(ctx context.Context, user *protobuf.CreateUserRequest) (*protobuf.User, error) {
	var userID string
	err := c.db.QueryRow(ctx, "INSERT INTO users (username , lat , long) VALUES ($1 , $2 , $3) RETURNING id", user.Id, user.Lat, user.Long).
		Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return &protobuf.User{
		Id:   userID,
		Lat:  user.Lat,
		Long: user.Long,
	}, nil
}
