package server

import (
	"context"
	"github.com/richard-on/auth-service/pkg/authService"
	"github.com/richard-on/task-service/config"
	"github.com/richard-on/task-service/internal/db"
	"github.com/richard-on/task-service/pkg/server/routes"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) Run() {
	// Waiting for quit signal on exit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	api := s.app.Group("/task")

	v1 := api.Group("/v1", func(c *fiber.Ctx) error {
		c.Set("Version", "v1.0")
		return c.Next()
	})

	cwt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(cwt, ":4000",
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		s.log.Fatal(err, "failed to connect to gRPC")
	}

	defer func(conn *grpc.ClientConn) {
		err = conn.Close()
		if err != nil {
			s.log.Fatal(err, "failed to close gRPC connection")
		}
	}(conn)

	mCtx := context.Background()

	mClient, err := mongo.Connect(mCtx,
		options.Client().ApplyURI(config.DbConnString))
	if err != nil {
		s.log.Fatal(err, "failed to connect to database")
	}

	collection := mClient.Database(config.MongoDbName).Collection(config.MongoCollection)

	defer func() {
		if err = mClient.Disconnect(mCtx); err != nil {
			s.log.Fatalf(err, "failed to disconnect db")
		}
	}()

	taskDb := db.NewDatabase(mCtx, collection)

	// Registering endpoints
	authClient := authService.NewAuthServiceClient(conn)
	routes.TaskRouter(v1, taskDb, authClient)

	go func() {
		if err = s.app.Listen(":5000"); err != nil {
			s.log.Fatalf(err, "error while listening at port 80")
		}
	}()

	<-quit

	err = s.app.Shutdown()
	if err != nil {
		s.log.Fatalf(err, "could not shutdown server")
	}
	s.log.Info("server shutdown success")
}
