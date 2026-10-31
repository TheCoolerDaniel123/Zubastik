package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"shared/entity"
	sharedjwt "shared/jwt"
	"user-service/api/dto"
	"user-service/internal/storage/database"
)

type UserUseCase struct {
	userRepo     database.UserDatabase
	toothRepo    database.ToothDatabase
	brushRepo    database.BrushingTimeDatabase
	calendarRepo database.CalendarDatabase
}

func NewUserUseCase(db database.UserDatabase, toothRepo database.ToothDatabase, brushRepo database.BrushingTimeDatabase, calendarRepo database.CalendarDatabase) *UserUseCase {
	return &UserUseCase{
		userRepo:     db,
		toothRepo:    toothRepo,
		brushRepo:    brushRepo,
		calendarRepo: calendarRepo,
	}
}

func (uc *UserUseCase) CreateUserInitially(ctx context.Context, user entity.User) error {
	// Initialize 32 default teeth
	for i := 0; i < 32; i++ {
		user.Teeth = append(user.Teeth, entity.Tooth{
			UserUUID:   user.UUID,
			ToothIndex: int16(i),
			ToothType:  entity.ToothTypeMolar,
			ToothState: entity.ToothStateHealthy,
		})
	}
	return uc.userRepo.CreateUser(ctx, user)
}

func (uc *UserUseCase) GetUserByUUID(ctx context.Context) (*entity.User, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}

	return uc.userRepo.GetUserByID(ctx, claims.UserID)
}

func (uc *UserUseCase) UpdateProfile(ctx context.Context, body dto.UpdateProfileBody) error {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return ErrNoAuthClaims
	}
	return uc.userRepo.UpdateUser(ctx, claims.UserID, body)
}

func (uc *UserUseCase) GetTeeth(ctx context.Context) ([]entity.Tooth, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.toothRepo.GetTeethByUserUUID(ctx, claims.UserID)
}

func (uc *UserUseCase) UpdateTeeth(ctx context.Context, body dto.UpdateTeethBody) error {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return ErrNoAuthClaims
	}
	for i, item := range body.Teeth {
		tooth := entity.Tooth{
			UserUUID:   claims.UserID,
			ToothIndex: int16(i),
			ToothState: entity.ToothState(item.ToothState),
			ToothNote:  item.ToothNote,
		}
		if err := uc.toothRepo.UpdateTooth(ctx, tooth); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UserUseCase) GetBrushSchedule(ctx context.Context) ([]entity.BrushTime, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.brushRepo.GetBrushingTimesByUserUUID(ctx, claims.UserID)
}

func (uc *UserUseCase) GetCalendarEvents(ctx context.Context) ([]entity.CalendarEvent, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	return uc.calendarRepo.GetCalendarEventsByUserUUID(ctx, claims.UserID)
}

func (uc *UserUseCase) CreateCalendarEvent(ctx context.Context, body dto.CreateCalendarEventBody) (*entity.CalendarEvent, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	event := entity.CalendarEvent{
		UserUUID:    claims.UserID,
		Date:        body.Date,
		Time:        body.Time,
		Description: body.Description,
		Status:      entity.CalendarEventStatus(body.Status),
	}
	return uc.calendarRepo.CreateCalendarEvent(ctx, event)
}

func (uc *UserUseCase) DeleteCalendarEvent(ctx context.Context, eventID uuid.UUID) error {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return ErrNoAuthClaims
	}
	return uc.calendarRepo.DeleteCalendarEvent(ctx, eventID, claims.UserID)
}

func (uc *UserUseCase) UpdateBrushSchedule(ctx context.Context, body dto.UpdateBrushScheduleBody) error {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return ErrNoAuthClaims
	}
	brushTimes := make([]entity.BrushTime, 0, len(body.TimesBrush))
	for _, ts := range body.TimesBrush {
		t, err := time.Parse("15:04", ts)
		if err != nil {
			return err
		}
		brushTimes = append(brushTimes, entity.BrushTime{UserUUID: claims.UserID, Time: t})
	}
	return uc.brushRepo.ReplaceAllBrushingTimes(ctx, claims.UserID, brushTimes)
}
