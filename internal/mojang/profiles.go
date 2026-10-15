package mojang

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Profiles struct {
	client *Client
}

func (s *Profiles) GetByUsername(ctx context.Context, username string) (*Profile, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("Missing username")
	}

	var profile Profile
	endpoint := BaseURL + "/users/profiles/minecraft/" + username

	if err := s.client.get(ctx, endpoint, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (s *Profiles) GetByUUID(ctx context.Context, uuid string) (*Profile, error) {
	uuid = strings.ReplaceAll(strings.TrimSpace(uuid), "-", "")
	if uuid == "" {
		return nil, fmt.Errorf("Invalid UUID")
	}

	var profile Profile
	endpoint := SessionURL + "/session/minecraft/profile/" + uuid

	if err := s.client.get(ctx, endpoint, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

var _ = http.MethodGet
