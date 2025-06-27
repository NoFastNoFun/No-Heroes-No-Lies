package pb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"no-heroes-no-lies/internal/models"
)

const aoHeader = "AO_KEY"

// Client wraps basic PocketBase REST calls.
type Client struct {
	baseURL string
	aoKey   string
	http    *http.Client
}

// NewClient constructs a PocketBase client.
func NewClient(baseURL string, aoKey string) *Client {
	return &Client{
		baseURL: baseURL,
		aoKey:   aoKey,
		http:    &http.Client{},
	}
}

// withHeaders applies common headers to every request.
func (c *Client) withHeaders(req *http.Request) {
	req.Header.Set(aoHeader, c.aoKey)
}

// FetchSession retrieves a session by ID.
func (c *Client) FetchSession(id string) (models.GameSession, error) {
	var session models.GameSession

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/collections/game_sessions/records/%s", c.baseURL, id),
		nil,
	)
	c.withHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return session, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return session, errors.New("failed to fetch session")
	}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return session, err
	}
	return session, nil
}

// UpdateSession patches the session state.
func (c *Client) UpdateSession(id string, state models.GameState) error {
	payload := map[string]interface{}{"state": state}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		"PATCH",
		fmt.Sprintf("%s/api/collections/game_sessions/records/%s", c.baseURL, id),
		bytes.NewReader(body),
	)
	c.withHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed: %s", string(b))
	}
	return nil
}

// VerifyUserToken hits the auth-refresh route and returns the user ID on success.
func (c *Client) VerifyUserToken(userToken string) (string, error) {
	type respBody struct {
		Record struct {
			ID string `json:"id"`
		} `json:"record"`
	}

	url := fmt.Sprintf("%s/api/collections/games_accounts/auth-refresh", c.baseURL)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set(aoHeader, c.aoKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("token verification failed")
	}

	var body respBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Record.ID, nil
}
