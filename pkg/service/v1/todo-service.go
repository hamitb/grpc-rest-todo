package v1

import (
	"context"
	"database/sql"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/hamitb/go-grpc-http-rest-microservice/pkg/api/v1"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

// todoServiceServer is implementation for v1.TodoServiceServer proto interface
type todoServiceServer struct {
	db *sql.DB
}

// NewTodoServiceServer creates Todo service
func NewTodoServiceServer(db *sql.DB) v1.TodoServiceServer {
	return &todoServiceServer{db: db}
}

// checkAPI checks if the API version requested by the client is supported by server
func (s *todoServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if api != apiVersion {
			return status.Errorf(codes.Unimplemented, "unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// connect returns SQL database connection from the pool
func (s *todoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to connect to database-> '%s'", err.Error())
	}
	return c, nil
}

// Create new todo task
func (s *todoServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	// check the API version
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.Todo.Reminder)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "reminder filed has invalid format-> '%s'", err.Error())
	}

	// insert Todo entity data
	res, err := c.ExecContext(ctx, "INSERT INTO todo(title, description, reminder) VALUES(?, ?, ?)",
		req.Todo.Title, req.Todo.Description, req.Todo.Reminder)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to insert into Todo-> '%s'", err.Error())
	}

	// get ID of creates Todo
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve id for created Todo-> '%s'", err.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

func (s *todoServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	// check the API version
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// read Todo entity
	rows, err := c.QueryContext(ctx, "SELECT id, title, description, reminder FROM todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Errorf(codes.Unknown, "failed to select from todo-> '%s'", err.Error())
		}
		return nil, status.Errorf(codes.NotFound, "Todo with ID='%d' is not found", req.Id)
	}

	var td v1.Todo
	var reminder time.Time
	if err := rows.Scan(&td.Id, &td.Title, &td.Description, reminder); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve field values from Todo row-> '%s'", err.Error())
	}
	td.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "reminder field has invalid format-> '%s'", err.Error())
	}

	if rows.Next() {
		return nil, status.Errorf(codes.Unknown, "found multiple Todo rows with ID='%d'", req.Id)
	}

	return &v1.ReadResponse{
		Api:  apiVersion,
		Todo: &td,
	}, nil
}