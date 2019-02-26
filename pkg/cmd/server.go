package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/hamitb/go-grpc-http-rest-microservice/pkg/protocol/grpc"
	v1 "github.com/hamitb/go-grpc-http-rest-microservice/pkg/service/v1"
	_ "github.com/lib/pq"
)

// Config is configuration for Server
type Config struct {
	// gRPC server start parameters section
	// GRPCPort is TCP port to listen by gRPC server
	GRPCPort string

	// DB Datastore parameters section
	// DatastoreDBHost is host of database
	DatastoreDBHost string
	// DatastoreDBUser is port to connect to database
	DatastoreDBPort int
	// DatastoreDBUser is username to connect to database
	DatastoreDBUser string
	//DatastoreDBPassword password to connect to database
	DatastoreDBPassword string
	// DatastoreDBName is schema of database
	DatastoreDBName string
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	// get configuration
	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.DatastoreDBHost, "db-host", "", "Database host")
	flag.IntVar(&cfg.DatastoreDBPort, "db-port", 5432, "Database port")
	flag.StringVar(&cfg.DatastoreDBUser, "db-user", "", "Database user")
	flag.StringVar(&cfg.DatastoreDBPassword, "db-password", "no-pass", "Database password")
	flag.StringVar(&cfg.DatastoreDBName, "db-name", "", "Database name")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.DatastoreDBHost,
		cfg.DatastoreDBPort,
		cfg.DatastoreDBUser,
		cfg.DatastoreDBPassword,
		cfg.DatastoreDBName)
	log.Println(dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	v1API := v1.NewTodoServiceServer(db)

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort)
}
