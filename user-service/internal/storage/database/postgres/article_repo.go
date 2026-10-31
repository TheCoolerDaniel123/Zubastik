package postgres

import (
	"context"

	"github.com/google/uuid"
	pgclient "shared/clients/postgres"
	"shared/entity"
)

type ArticleRepo struct {
	cli *pgclient.PostgresClient
}

func NewArticleRepo(cli *pgclient.PostgresClient) *ArticleRepo {
	return &ArticleRepo{cli: cli}
}

func (r *ArticleRepo) Create(ctx context.Context, article entity.Article) (*entity.Article, error) {
	err := r.cli.DB.QueryRowContext(ctx,
		`INSERT INTO articles (author_uuid, title, content, status, image_url)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at::text`,
		article.AuthorUUID,
		article.Title,
		article.Content,
		article.Status,
		article.ImageURL,
	).Scan(&article.ID, &article.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepo) GetApproved(ctx context.Context) ([]entity.Article, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT a.id, a.author_uuid, COALESCE(u.name, '') as author_name, a.title, a.content, a.status, a.image_url, a.created_at::text
		 FROM articles a
		 LEFT JOIN users u ON u.uuid = a.author_uuid
		 WHERE a.status = 'approved'
		 ORDER BY a.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]entity.Article, 0)
	for rows.Next() {
		var a entity.Article
		if err := rows.Scan(&a.ID, &a.AuthorUUID, &a.AuthorName, &a.Title, &a.Content, &a.Status, &a.ImageURL, &a.CreatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (r *ArticleRepo) GetAll(ctx context.Context) ([]entity.Article, error) {
	rows, err := r.cli.DB.QueryContext(ctx,
		`SELECT a.id, a.author_uuid, COALESCE(u.name, '') as author_name, a.title, a.content, a.status, a.image_url, a.created_at::text
		 FROM articles a
		 LEFT JOIN users u ON u.uuid = a.author_uuid
		 ORDER BY a.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]entity.Article, 0)
	for rows.Next() {
		var a entity.Article
		if err := rows.Scan(&a.ID, &a.AuthorUUID, &a.AuthorName, &a.Title, &a.Content, &a.Status, &a.ImageURL, &a.CreatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (r *ArticleRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ArticleStatus) error {
	_, err := r.cli.DB.ExecContext(ctx,
		`UPDATE articles SET status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}
