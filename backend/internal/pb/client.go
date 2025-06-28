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
func NewClient(baseURL, aoKey string) *Client {
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

// InsertSession creates a new game_sessions record.
func (c *Client) InsertSession(session models.GameSession) (models.GameSession, error) {
	var out models.GameSession
	body, _ := json.Marshal(session)

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/collections/game_sessions/records", c.baseURL),
		bytes.NewReader(body),
	)
	c.withHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return out, errors.New("failed to create session")
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

// UpdateSession updates state and is_active in one patch.
func (c *Client) UpdateSession(id string, state models.GameState, active bool) error {
	payload := map[string]interface{}{
		"state":     state,
		"is_active": active,
	}
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

// UpdateSessionPlayers patches only the player_ids field.
func (c *Client) UpdateSessionPlayers(id string, players []string) error {
	payload := map[string]interface{}{
		"player_ids": players,
	}
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
		return errors.New("failed to patch players")
	}
	return nil
}

// ListCards fetches all card definitions.
func (c *Client) ListCards() ([]models.Card, error) {
	type result struct {
		Items []models.Card `json:"items"`
	}
	var r result

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/collections/cards/records?perPage=200", c.baseURL),
		nil,
	)
	c.withHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to list cards")
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.Items, nil
}

// InsertMove writes a move record to PocketBase.
func (c *Client) InsertMove(m models.Move) error {
	type payload struct {
		SessionID string `json:"session_id"`
		PlayerID  string `json:"player_id"`
		Type      string `json:"type"`
		Data      string `json:"move_data"`
	}
	body, _ := json.Marshal(payload{
		SessionID: m.SessionID,
		PlayerID:  m.PlayerID,
		Type:      m.Type,
		Data:      m.Data,
	})

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/collections/moves/records", c.baseURL),
		bytes.NewReader(body),
	)
	c.withHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New("failed to insert move")
	}
	return nil
}

func (c *Client) GetCard(id string) (models.Card, error) {
	var card models.Card

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/collections/cards/records/%s", c.baseURL, id),
		nil,
	)
	c.withHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return card, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return card, errors.New("card not found")
	}
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return card, err
	}
	return card, nil
}
