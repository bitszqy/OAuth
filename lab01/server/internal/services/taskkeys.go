package services

import (
	"context"
	"errors"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/taskpolicy"
	"github.com/cyllective/oauth-labs/lab01/server/internal/variant"
)

var ErrTaskNotAuthorized = errors.New("task is not authorized")

type TaskGrantChecker interface {
	HasActive(ctx context.Context, grant *dto.TaskGrant) bool
}

type TaskKeyService struct {
	grants TaskGrantChecker
	crypto *AttributeCryptoService
}

func NewTaskKeyService(grants TaskGrantChecker, crypto *AttributeCryptoService) *TaskKeyService {
	return &TaskKeyService{grants: grants, crypto: crypto}
}

func (s *TaskKeyService) Release(ctx context.Context, userID, clientID string, request *dto.TaskKeyRequest) (*dto.TaskKeyResponse, error) {
	attributes := taskpolicy.AllAttributes()
	if !variant.IsVulnerable() {
		task, ok := taskpolicy.Lookup(request.TaskID)
		if !ok {
			return nil, ErrTaskNotAuthorized
		}
		grant := &dto.TaskGrant{
			UserID:   userID,
			ClientID: clientID,
			TaskID:   request.TaskID,
		}
		if !s.grants.HasActive(ctx, grant) {
			return nil, ErrTaskNotAuthorized
		}
		attributes = task.AllowedAttributes
	}

	keys := make(map[string]dto.FieldKey, len(attributes))
	for _, attribute := range attributes {
		key, err := s.crypto.Key(userID, request.ResourceID, attribute)
		if err != nil {
			return nil, err
		}
		keys[attribute] = key
	}
	return &dto.TaskKeyResponse{
		TaskID:            request.TaskID,
		AllowedAttributes: attributes,
		Keys:              keys,
	}, nil
}
