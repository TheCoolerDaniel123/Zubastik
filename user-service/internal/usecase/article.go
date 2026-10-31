package usecase

import (
	"context"

	"github.com/google/uuid"
	"shared/entity"
	sharedjwt "shared/jwt"
	"user-service/api/dto"
	"user-service/internal/storage/database"
)

type ArticleUseCase struct {
	articleRepo database.ArticleDatabase
}

func NewArticleUseCase(articleRepo database.ArticleDatabase) *ArticleUseCase {
	return &ArticleUseCase{articleRepo: articleRepo}
}

func (uc *ArticleUseCase) GetApprovedArticles(ctx context.Context) ([]entity.Article, error) {
	return uc.articleRepo.GetApproved(ctx)
}

func (uc *ArticleUseCase) CreateArticle(ctx context.Context, body dto.CreateArticleBody) (*entity.Article, error) {
	claims := sharedjwt.GetClaimsFromCtx(ctx)
	if claims == nil {
		return nil, ErrNoAuthClaims
	}
	article := entity.Article{
		AuthorUUID: claims.UserID,
		Title:      body.Title,
		Content:    body.Content,
		ImageURL:   body.ImageURL,
		Status:     entity.ArticleStatusPending,
	}
	return uc.articleRepo.Create(ctx, article)
}

func (uc *ArticleUseCase) GetAllArticles(ctx context.Context) ([]entity.Article, error) {
	return uc.articleRepo.GetAll(ctx)
}

func (uc *ArticleUseCase) ApproveArticle(ctx context.Context, id uuid.UUID) error {
	return uc.articleRepo.UpdateStatus(ctx, id, entity.ArticleStatusApproved)
}

func (uc *ArticleUseCase) RejectArticle(ctx context.Context, id uuid.UUID) error {
	return uc.articleRepo.UpdateStatus(ctx, id, entity.ArticleStatusRejected)
}
