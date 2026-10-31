package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "shared/api/pb"
	"shared/entity"
	"user-service/internal/usecase"
)

type UserGRPCController struct {
	pb.UnimplementedUserServiceServer
	uc *usecase.UserUseCase
}

func NewUserGRPCController(uc *usecase.UserUseCase) *UserGRPCController {
	return &UserGRPCController{
		uc: uc,
	}
}

func (c *UserGRPCController) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	id, err := uuid.Parse(req.Uuid)
	if err != nil {
		return &pb.CreateUserResponse{
			Ok:           false,
			ErrorMessage: "invalid uuid",
		}, nil
	}

	user := entity.User{
		UUID:  id,
		Phone: req.Phone,
		Role:  entity.UserRole(req.Role),
	}

	err = c.uc.CreateUserInitially(ctx, user)
	if err != nil {
		return &pb.CreateUserResponse{
			Ok:           false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.CreateUserResponse{Ok: true}, nil
}
