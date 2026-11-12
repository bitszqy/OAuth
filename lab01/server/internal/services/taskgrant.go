package services

import (
	"context"
	"errors"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/repositories"
)

type TaskGrantService struct {
	repository *repositories.TaskGrantRepository
}

func NewTaskGrantService(repository *repositories.TaskGrantRepository) *TaskGrantService {
	return &TaskGrantService{repository: repository}
}

func (s *TaskGrantService) Create(ctx context.Context, grant *dto.TaskGrant) error {
	if grant.UserID == "" || grant.ClientID == "" || grant.TaskID == "" {
		return errors.New("invalid task grant")
	}
	return s.repository.Upsert(ctx, grant)
}

func (s *TaskGrantService) HasActive(ctx context.Context, grant *dto.TaskGrant) bool {
	if grant.UserID == "" || grant.ClientID == "" || grant.TaskID == "" {
		return false
	}
	return s.repository.Exists(ctx, grant)
}

func (s *TaskGrantService) RevokeAll(ctx context.Context, clientID, userID string) error {
	return s.repository.RevokeAll(ctx, clientID, userID)
}
