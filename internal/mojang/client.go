package mojang

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	BaseURL    = "https://api.mojang.com"
	SessionURL = "https://sessionserver.mojang.com"
)

type Client struct {
	REST     *resty.Client
	Profiles *Profiles
}

func NewClient() *Client {
	c := &Client{
		REST: resty.New().
			SetTimeout(10 * time.Second),
	}

	c.Profiles = &Profiles{client: c}

	return c
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Mojang API (%d): %s", e.StatusCode, e.Message)
}

func (c *Client) get(ctx context.Context, url string, t any) error {
	resp, err := c.REST.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetResult(t).
		Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return &APIError{
			StatusCode: resp.StatusCode(),
			Message:    resp.String(),
		}
	}

	return nil
}
