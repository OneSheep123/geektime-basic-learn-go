package repository

import (
	"context"
	"ddd_demo/internal/domain"
)

//go:generate mockgen -source=./article_reader.go -package=repomocks -destination=./mocks/article_reader.mock.go ArticleReaderRepository
type ArticleReaderRepository interface {
	// Save 有则更新，无则插入，也就是 insert or update 语义
	Save(ctx context.Context, art domain.Article) error
}
