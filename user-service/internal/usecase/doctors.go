package usecase

import (
	"context"

	"shared/entity"
	"user-service/internal/storage/database"
)

type DoctorsUseCase struct {
	profileRepo database.DoctorProfileDatabase
}

func NewDoctorsUseCase(profileRepo database.DoctorProfileDatabase) *DoctorsUseCase {
	return &DoctorsUseCase{profileRepo: profileRepo}
}

func (uc *DoctorsUseCase) GetAll(ctx context.Context) ([]entity.DoctorProfile, error) {
	return uc.profileRepo.GetAll(ctx)
}
