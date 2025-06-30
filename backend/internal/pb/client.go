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

// VerifyUserToken verifies a user token and returns the user ID on success.
func (c *Client) VerifyUserToken(userToken string) (string, error) {
	type respBody struct {
		Record struct {
			ID string `json:"id"`
		} `json:"record"`
	}

	// Use the auth-verify endpoint instead of auth-refresh for token verification
	url := fmt.Sprintf("%s/api/collections/games_accounts/auth-verify", c.baseURL)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set(aoHeader, c.aoKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try auth-refresh as fallback for backward compatibility
		url = fmt.Sprintf("%s/api/collections/games_accounts/auth-refresh", c.baseURL)
		req, _ = http.NewRequest("POST", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set(aoHeader, c.aoKey)

		resp, err = c.http.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", errors.New("token verification failed")
		}
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

func (c *Client) GetPower(id string) (models.Power, error) {
	var power models.Power

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/collections/powers/records/%s", c.baseURL, id),
		nil,
	)
	c.withHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return power, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return power, errors.New("power not found")
	}
	if err := json.NewDecoder(resp.Body).Decode(&power); err != nil {
		return power, err
	}
	return power, nil
}

// AuthWithPassword authenticates a user with PocketBase and returns user info and user ID.
func (c *Client) AuthWithPassword(email, password string) (struct {
	UserID string
	User   any
}, error) {
	type respBody struct {
		Token  string `json:"token"`
		Record struct {
			ID          string `json:"id"`
			Email       string `json:"email"`
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
			// Add more fields as needed
		} `json:"record"`
	}
	payload := map[string]string{
		"identity": email,
		"password": password,
	}
	body, _ := json.Marshal(payload)
	url := c.baseURL + "/api/collections/games_accounts/auth-with-password"

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.withHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return struct {
			UserID string
			User   any
		}{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return struct {
			UserID string
			User   any
		}{}, errors.New("auth failed: " + string(b))
	}
	var out respBody
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return struct {
			UserID string
			User   any
		}{}, err
	}
	return struct {
		UserID string
		User   any
	}{UserID: out.Record.ID, User: out.Record}, nil
}

// RegisterUser registers a new user in PocketBase and returns user info and user ID.
func (c *Client) RegisterUser(email, password, username, displayName string) (struct {
	UserID string
	User   any
}, error) {
	type respBody struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		// Add more fields as needed
	}
	payload := map[string]string{
		"email":           email,
		"password":        password,
		"passwordConfirm": password,
		"username":        username,
		"display_name":    displayName,
	}
	body, _ := json.Marshal(payload)
	url := c.baseURL + "/api/collections/games_accounts/records"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.withHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return struct {
			UserID string
			User   any
		}{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return struct {
			UserID string
			User   any
		}{}, errors.New("registration failed: " + string(b))
	}
	var out respBody
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return struct {
			UserID string
			User   any
		}{}, err
	}
	return struct {
		UserID string
		User   any
	}{UserID: out.ID, User: out}, nil
}

// AuthWithPasswordFull authenticates a user and returns userID, access/refresh tokens, and user info.
func (c *Client) AuthWithPasswordFull(email, password string) (userID, accessToken, refreshToken string, user any, err error) {
	type respBody struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
		Record       struct {
			ID          string `json:"id"`
			Email       string `json:"email"`
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
			// Add more fields as needed
		} `json:"record"`
	}
	payload := map[string]string{
		"identity": email,
		"password": password,
	}
	body, _ := json.Marshal(payload)
	url := c.baseURL + "/api/collections/users/auth-with-password"

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.withHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		err = errors.New("auth failed: " + string(b))
		return
	}
	var out respBody
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return
	}
	userID = out.Record.ID
	accessToken = out.Token
	refreshToken = out.RefreshToken
	user = out.Record
	return
}

// AuthRefresh refreshes PB tokens using the current access and refresh tokens.
func (c *Client) AuthRefresh(accessToken, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	type respBody struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	}
	payload := map[string]string{
		"token":        accessToken,
		"refreshToken": refreshToken,
	}
	body, _ := json.Marshal(payload)
	url := c.baseURL + "/api/collections/users/auth-refresh"

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.withHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		err = errors.New("refresh failed: " + string(b))
		return
	}
	var out respBody
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return
	}
	newAccessToken = out.Token
	newRefreshToken = out.RefreshToken
	return
}

// AuthLogout revokes PB tokens using the access token.
func (c *Client) AuthLogout(accessToken string) error {
	url := c.baseURL + "/api/collections/users/logout"
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	c.withHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return errors.New("logout failed: " + string(b))
	}
	return nil
}
