package usecase

import (
	"context"

	"github.com/google/uuid"
	"shared/entity"
	sharedjwt "shared/jwt"
	"user-service/internal/storage/database"
)

type DoctorUseCase struct {
	userRepo  database.UserDatabase
	toothRepo database.ToothDatabase
	noteRepo  database.ToothNoteDatabase
	recRepo   database.RecommendationDatabase
}

func NewDoctorUseCase(
	userRepo database.UserDatabase,
	toothRepo database.ToothDatabase,
	noteRepo database.ToothNoteDatabase,
	recRepo database.RecommendationDatabase,
) *DoctorUseCase {
	return &DoctorUseCase{
		userRepo:  userRepo,
		toothRepo: toothRepo,
		noteRepo:  noteRepo,
		recRepo:   recRepo,
	}
}

func (uc *DoctorUseCase) GetPatientByUUID(ctx context.Context, patientUUID uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetUserByID(ctx, patientUUID)
}

func (uc *DoctorUseCase) GetPatientTeeth(ctx context.Context, patientUUID uuid.UUID) ([]entity.Tooth, error) {
	return uc.toothRepo.GetTeethByUserUUID(ctx, patientUUID)
}

func (uc *DoctorUseCase) CreateToothNote(ctx context.Context, note entity.ToothNote) (*entity.ToothNote, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	note.DoctorUUID = claims.UserID
	return uc.noteRepo.CreateToothNote(ctx, note)
}

func (uc *DoctorUseCase) GetToothNotes(ctx context.Context, patientUUID uuid.UUID) ([]entity.ToothNote, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.noteRepo.GetToothNotesByPatient(ctx, claims.UserID, patientUUID)
}

func (uc *DoctorUseCase) UpdateToothNote(ctx context.Context, noteID uuid.UUID, newNote string) error {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return ErrNoAuthClaims
	}
	return uc.noteRepo.UpdateToothNote(ctx, noteID, claims.UserID, newNote)
}

func (uc *DoctorUseCase) DeleteToothNote(ctx context.Context, noteID uuid.UUID) error {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return ErrNoAuthClaims
	}
	return uc.noteRepo.DeleteToothNote(ctx, noteID, claims.UserID)
}

func (uc *DoctorUseCase) CreateRecommendation(ctx context.Context, rec entity.Recommendation) (*entity.Recommendation, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	rec.DoctorUUID = claims.UserID
	return uc.recRepo.CreateRecommendation(ctx, rec)
}

func (uc *DoctorUseCase) GetRecommendations(ctx context.Context, patientUUID uuid.UUID) ([]entity.Recommendation, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.recRepo.GetRecommendationsByPatient(ctx, claims.UserID, patientUUID)
}
