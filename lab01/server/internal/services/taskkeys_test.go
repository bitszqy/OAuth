package services

import (
	"bytes"
	"context"
	"testing"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/taskpolicy"
)

type fakeGrantChecker struct {
	allowed bool
}

func (f fakeGrantChecker) HasActive(context.Context, *dto.TaskGrant) bool {
	return f.allowed
}

func TestTaskKeysFixedVariantLimitsAttributes(t *testing.T) {
	t.Setenv("LAB_VARIANT", "fixed")
	service := NewTaskKeyService(fakeGrantChecker{allowed: true}, NewAttributeCryptoService(bytes.Repeat([]byte{7}, 32)))
	response, err := service.Release(context.Background(), "user-1", "client-1", &dto.TaskKeyRequest{
		TaskID:     "calendar.freebusy",
		ResourceID: "calendar:user-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.AllowedAttributes) != 2 || response.AllowedAttributes[0] != taskpolicy.Start || response.AllowedAttributes[1] != taskpolicy.End {
		t.Fatalf("unexpected allowed attributes: %#v", response.AllowedAttributes)
	}
	if _, ok := response.Keys[taskpolicy.Title]; ok {
		t.Fatal("fixed variant must not release the title key")
	}
}

func TestTaskKeysFixedVariantRejectsMissingGrant(t *testing.T) {
	t.Setenv("LAB_VARIANT", "fixed")
	service := NewTaskKeyService(fakeGrantChecker{allowed: false}, NewAttributeCryptoService(bytes.Repeat([]byte{7}, 32)))
	_, err := service.Release(context.Background(), "user-1", "client-1", &dto.TaskKeyRequest{
		TaskID:     "calendar.freebusy",
		ResourceID: "calendar:user-1",
	})
	if err != ErrTaskNotAuthorized {
		t.Fatalf("expected ErrTaskNotAuthorized, got %v", err)
	}
}

func TestTaskKeysVulnerableVariantReleasesAllKeys(t *testing.T) {
	t.Setenv("LAB_VARIANT", "vulnerable")
	service := NewTaskKeyService(fakeGrantChecker{allowed: false}, NewAttributeCryptoService(bytes.Repeat([]byte{7}, 32)))
	response, err := service.Release(context.Background(), "user-1", "client-1", &dto.TaskKeyRequest{
		TaskID:     "calendar.freebusy",
		ResourceID: "calendar:user-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.AllowedAttributes) != len(taskpolicy.AllAttributes()) {
		t.Fatalf("unexpected vulnerable attribute count: %d", len(response.AllowedAttributes))
	}
	if _, ok := response.Keys[taskpolicy.Title]; !ok {
		t.Fatal("vulnerable variant must expose the title key")
	}
}
