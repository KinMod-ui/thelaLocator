package grpc

import (
	"context"
	"log"
	"time"

	"github.com/KinMod-ui/thelaLocator/pkg/db"
	"github.com/KinMod-ui/thelaLocator/pkg/redis"
	"github.com/KinMod-ui/thelaLocator/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	protobuf.UnimplementedUserServiceServer
	RedisClient redis.RedisClient
	DbClient    db.DatabaseClient
}

// GetUser retrieves a user from Redis or PostgreSQL.
func (s *UserService) GetUser(ctx context.Context, req *protobuf.GetUserRequest) (*protobuf.User, error) {
	// 1. Try to retrieve user from Redis
	user, err := s.RedisClient.GetUser(ctx, req.Id)
	if err == nil {
		return user, nil
	}

	// 2. If not found in Redis, retrieve from PostgreSQL
	user, err = s.DbClient.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
	}

	// 3. Cache user in Redis
	if err := s.RedisClient.SetUser(ctx, req.Id, user, time.Hour); err != nil {
		// Error caching, but still return user
		log.Printf("Error caching user in Redis: %v", err)
	}

	return user, nil
}

// CreateUser creates a new user and stores it in Redis and PostgreSQL.
func (s *UserService) CreateUser(ctx context.Context, req *protobuf.CreateUserRequest) (*protobuf.User, error) {
	// 1. Create user in PostgreSQL
	user, err := s.DbClient.CreateUser(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating user: %v", err)
	}

	// 2. Cache user in Redis
	if err := s.RedisClient.SetUser(ctx, req.Id, user, time.Hour); err != nil {
		return nil, status.Errorf(codes.Internal, "error caching user in Redis: %v", err)
	}

	return user, nil
}
