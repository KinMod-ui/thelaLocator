package main

import (
	"log"
	"net"
	"os"

	"github.com/KinMod-ui/thelaLocator/grpc"
	"github.com/KinMod-ui/thelaLocator/helper"
	"github.com/KinMod-ui/thelaLocator/pkg/db"
	"github.com/KinMod-ui/thelaLocator/pkg/redis"
	"github.com/KinMod-ui/thelaLocator/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"

	"go.uber.org/zap"
	googleGrpc "google.golang.org/grpc"
)

type user struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

func main() {

	port := os.Args[1]
	helper.Mylog.Println("Entering app")

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	redisClient, err := redis.NewRedisClient("localhost:6379", "", 0)

	if err != nil {
		logger.Sugar().Fatalf("Cannot open redis client: %s", err.Error())
		return
	}

	dbClient, err := db.NewDatabaseClient()
	if err != nil {
		logger.Sugar().Fatalf("Cannot open db connection: %s", err.Error())
	}

	server := googleGrpc.NewServer(
		googleGrpc.ChainUnaryInterceptor(
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_validator.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
		),
	)

	protobuf.RegisterUserServiceServer(server, &grpc.UserService{
		RedisClient: redisClient,
		DbClient:    dbClient,
	})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Server listening on port: %s", ":50051")

	// Serve gRPC requests
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
