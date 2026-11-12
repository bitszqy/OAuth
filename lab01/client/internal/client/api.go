package client

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab01/client/internal/config"
	"github.com/cyllective/oauth-labs/lab01/client/internal/constants"
	"github.com/cyllective/oauth-labs/lab01/client/internal/dto"
)

type APIClient struct {
	ctx       context.Context
	config    *oauth2.Config
	client    *http.Client
	baseURL   string
	userAgent string
}

func NewAPIClient(ctx context.Context, tokens *oauth2.Token) *APIClient {
	cfg := config.Get()
	client := &http.Client{Timeout: time.Duration(3) * time.Second}
	if proxyURL := os.Getenv("BURP_PROXY"); proxyURL != "" {
		proxy, _ := url.Parse(proxyURL)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	oacfg := config.GetOAuthConfig()
	oactx := context.WithValue(ctx, oauth2.HTTPClient, client)
	client = oauth2.NewClient(oactx, oacfg.TokenSource(oactx, tokens))
	return &APIClient{
		ctx:       oactx,
		config:    oacfg,
		client:    client,
		baseURL:   cfg.GetString("resource_server.base_url"),
		userAgent: fmt.Sprintf("oauth-labs/client-%s", constants.LabNumber),
	}
}

func (o *APIClient) GetProfile() (*dto.Profile, error) {
	url := fmt.Sprintf("%s/api/users/me", o.baseURL)
	var profile dto.Profile
	if err := o.getJSON(o.ctx, url, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (o *APIClient) GetCalendarTask(taskID string) (*dto.CalendarView, error) {
	var calendar dto.CalendarResponse
	if err := o.getJSON(o.ctx, fmt.Sprintf("%s/api/calendar/events", o.baseURL), &calendar); err != nil {
		return nil, err
	}

	if len(calendar.Events) > 0 {
		view := &dto.CalendarView{TaskID: taskID, Mode: "native-oauth-plaintext"}
		for _, event := range calendar.Events {
			view.Events = append(view.Events, dto.CalendarEventView{
				ID:         event.ID,
				ResourceID: event.ResourceID,
				Values: map[string]string{
					"start":       event.Start,
					"end":         event.End,
					"title":       event.Title,
					"location":    event.Location,
					"attendees":   stringsJoin(event.Attendees),
					"description": event.Description,
				},
			})
		}
		return view, nil
	}

	keys := make(map[string]dto.TaskKeyResponse)
	view := &dto.CalendarView{TaskID: taskID, Mode: "field-encryption"}
	for _, event := range calendar.Encrypted {
		taskKeys, ok := keys[event.ResourceID]
		if !ok {
			request := &dto.TaskKeyRequest{TaskID: taskID, ResourceID: event.ResourceID}
			response := new(dto.TaskKeyResponse)
			if err := o.postJSON(o.ctx, fmt.Sprintf("%s/api/task-keys", o.baseURL), request, response); err != nil {
				return nil, err
			}
			keys[event.ResourceID] = *response
			taskKeys = *response
		}

		values := make(map[string]string)
		locked := make([]string, 0)
		for attribute, field := range event.Fields {
			key, ok := taskKeys.Keys[attribute]
			if !ok {
				locked = append(locked, attribute)
				continue
			}
			value, err := decryptField(key, field)
			if err != nil {
				return nil, fmt.Errorf("decrypt %s for %s: %w", attribute, event.ID, err)
			}
			values[attribute] = value
		}
		view.Events = append(view.Events, dto.CalendarEventView{
			ID:         event.ID,
			ResourceID: event.ResourceID,
			Values:     values,
			Locked:     locked,
		})
	}
	return view, nil
}

func (o *APIClient) getJSON(ctx context.Context, url string, model any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", o.userAgent)
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("api request to %#v failed: %w", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("api request to %#v failed: invalid status code", url)
	}
	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("api request to %#v failed: reading response body: %w", url, err)
	}
	if err := json.Unmarshal(rawBody, &model); err != nil {
		return fmt.Errorf("api request to %#v failed: unmarshal failed: %w", url, err)
	}
	return nil
}

func (o *APIClient) postJSON(ctx context.Context, url string, body, model any) error {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request for %#v: %w", url, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("create request for %#v: %w", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", o.userAgent)
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("api request to %#v failed: %w", url, err)
	}
	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response from %#v: %w", url, err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("api request to %#v failed with status %d: %s", url, res.StatusCode, string(responseBody))
	}
	if err := json.Unmarshal(responseBody, model); err != nil {
		return fmt.Errorf("unmarshal response from %#v: %w", url, err)
	}
	return nil
}

func decryptField(key dto.FieldKey, field dto.EncryptedField) (string, error) {
	if key.Algorithm != "AES-256-GCM" || field.Algorithm != key.Algorithm {
		return "", fmt.Errorf("unsupported algorithm")
	}
	if key.KeyID != field.KeyID {
		return "", fmt.Errorf("field key id mismatch")
	}
	keyBytes, err := base64.RawURLEncoding.DecodeString(key.Key)
	if err != nil {
		return "", fmt.Errorf("decode key: %w", err)
	}
	nonce, err := base64.RawURLEncoding.DecodeString(field.Nonce)
	if err != nil {
		return "", fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(field.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	aad, err := base64.RawURLEncoding.DecodeString(key.AAD)
	if err != nil {
		return "", fmt.Errorf("decode aad: %w", err)
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func stringsJoin(values []string) string {
	result := ""
	for i, value := range values {
		if i > 0 {
			result += ", "
		}
		result += value
	}
	return result
}
