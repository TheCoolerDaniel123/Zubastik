package grpc

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/entity"
	pb "shared/api/pb"

	"google.golang.org/grpc"
)

type GRPCClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewGRPCClient(address string) (*GRPCClient, error) {

	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	client := pb.NewUserServiceClient(conn)

	return &GRPCClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *GRPCClient) CreateUser(ctx context.Context, user entity.User) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.CreateUser(ctx, &pb.CreateUserRequest{
		Uuid:  user.UUID.String(),
		Phone: user.Phone,
		Role:  int32(user.Role),
	})
	if err != nil {
		return err
	}

	if !resp.Ok {
		return errors.New(resp.ErrorMessage)
	}

	return nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}
