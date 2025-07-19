package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lavatee/astraltest/internal/model"
	"github.com/lavatee/astraltest/internal/repository"
	"github.com/redis/go-redis/v9"
)

type DocumentService struct {
	repo  *repository.Repository
	cache *redis.Client
}

func NewDocumentService(repo *repository.Repository, cache *redis.Client) *DocumentService {
	return &DocumentService{
		repo:  repo,
		cache: cache,
	}
}

func (s *DocumentService) invalidateUserCache(ctx context.Context, token string) error {
	pattern := "docs:" + token + "*"
	keys, err := s.cache.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		if err := s.cache.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *DocumentService) Upload(ctx context.Context, token, meta, jsonData string, file multipart.File, header *multipart.FileHeader, isFileLoaded bool) (*model.Document, error) {
	var metaData model.DocumentMeta
	if err := json.Unmarshal([]byte(meta), &metaData); err != nil {
		return nil, err
	}
	doc := &model.Document{
		ID:      uuid.New().String(),
		Name:    metaData.Name,
		Mime:    metaData.Mime,
		File:    metaData.File,
		Public:  metaData.Public,
		Created: time.Now(),
		Grant:   metaData.Grant,
	}
	var fileData []byte
	if metaData.File {
		var err error
		if !isFileLoaded {
			return nil, errors.New("File has not been loaded")
		}
		fileData, err = io.ReadAll(file)
		if err != nil {
			return nil, err
		}
	}
	if err := s.repo.Documents.Create(ctx, token, doc, jsonData, fileData); err != nil {
		return nil, err
	}
	if err := s.invalidateUserCache(ctx, token); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *DocumentService) GetAll(ctx context.Context, token, login, key, value string, limit int) ([]*model.Document, error) {
	cacheKey := "docs:" + token
	if login != "" {
		cacheKey += ":" + login
	}
	if key != "" && value != "" {
		cacheKey += ":" + key + ":" + value
	}
	if limit > 0 {
		cacheKey += ":limit:" + strconv.Itoa(limit)
	}
	if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
		var docs []*model.Document
		if err := json.Unmarshal([]byte(cached), &docs); err == nil {
			return docs, nil
		}
	}
	docs, err := s.repo.GetAll(ctx, token, login, key, value, limit)
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(docs); err == nil {
		s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}
	return docs, nil
}

func (s *DocumentService) GetByID(ctx context.Context, token, id string) (*model.Document, []byte, error) {
	cacheKey := "doc:" + id
	if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
		var doc model.Document
		if err := json.Unmarshal([]byte(cached), &doc); err == nil {
			fileData, err := s.repo.GetFileData(ctx, token, id)
			return &doc, fileData, err
		}
	}
	doc, fileData, err := s.repo.Documents.GetByID(ctx, token, id)
	if err != nil {
		return nil, nil, err
	}

	if data, err := json.Marshal(doc); err == nil {
		s.cache.Set(ctx, cacheKey, data, 10*time.Minute)
	}
	return doc, fileData, nil
}

func (s *DocumentService) Delete(ctx context.Context, token, id string) error {
	if err := s.repo.Delete(ctx, token, id); err != nil {
		return err
	}
	s.cache.Del(ctx, "doc:"+id)
	if err := s.invalidateUserCache(ctx, token); err != nil {
		return err
	}
	return nil
}
