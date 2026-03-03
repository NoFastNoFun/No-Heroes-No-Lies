package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"no-heroes-no-lies/internal/models"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitPostgres initializes the global PostgreSQL connection using environment variables.
func InitPostgres() error {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	sslmode := os.Getenv("POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	return DB.Ping()
}

// InsertMove inserts a move into the moves table.
func InsertMove(m models.Move) error {
	_, err := DB.Exec(`INSERT INTO moves (session_id, player_id, type, move_data, created_at) VALUES ($1, $2, $3, $4, NOW())`, m.SessionID, m.PlayerID, m.Type, m.Data)
	return err
}

// User represents a user in the system.
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    string
}

// CreateUser inserts a new user into the users table.
func CreateUser(email, username, passwordHash string) (User, error) {
	var user User
	row := DB.QueryRow(`INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id, email, username, password_hash, created_at`,
		email, username, passwordHash)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

// GetUserByEmail fetches a user by email.
func GetUserByEmail(email string) (User, error) {
	var user User
	row := DB.QueryRow(`SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1`, email)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

// GetUserByUsername fetches a user by username.
func GetUserByUsername(username string) (User, error) {
	var user User
	row := DB.QueryRow(`SELECT id, email, username, password_hash, created_at FROM users WHERE username = $1`, username)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

// ArchiveGame saves a finished game session and its moves to the archived_games table.
func ArchiveGame(session models.GameSession, moves []models.Move) error {
	movesJSON, err := json.Marshal(moves)
	if err != nil {
		return err
	}
	sessionJSON, err := json.Marshal(session.State)
	if err != nil {
		return err
	}
	_, err = DB.Exec(`INSERT INTO archived_games (id, player_ids, state, moves, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		session.ID, pqStringArray(session.PlayerIDs), sessionJSON, movesJSON, session.IsActive, session.CreatedAt, session.UpdatedAt)
	return err
}

// GetArchivedGame fetches an archived game by ID.
func GetArchivedGame(id string) (models.GameSession, []models.Move, error) {
	var stateJSON, movesJSON []byte
	var session models.GameSession
	var playerIDs []string
	var isActive bool
	var createdAt, updatedAt string
	row := DB.QueryRow(`SELECT player_ids, state, moves, is_active, created_at, updated_at FROM archived_games WHERE id = $1`, id)
	err := row.Scan(&playerIDs, &stateJSON, &movesJSON, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return session, nil, err
	}
	session.ID = id
	session.PlayerIDs = playerIDs
	session.IsActive = isActive
	session.CreatedAt = createdAt
	session.UpdatedAt = updatedAt
	json.Unmarshal(stateJSON, &session.State)
	var moves []models.Move
	json.Unmarshal(movesJSON, &moves)
	return session, moves, nil
}

// pqStringArray is a helper for []string to Postgres text[]
func pqStringArray(arr []string) interface{} {
	type stringArray []string
	return stringArray(arr)
}

// FetchMovesForSession returns all moves for a session, ordered by creation time.
func FetchMovesForSession(sessionID string) ([]models.Move, error) {
	rows, err := DB.Query(`SELECT id, session_id, player_id, type, move_data, created_at FROM moves WHERE session_id = $1 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var moves []models.Move
	for rows.Next() {
		var m models.Move
		err := rows.Scan(&m.ID, &m.SessionID, &m.PlayerID, &m.Type, &m.Data, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		moves = append(moves, m)
	}
	return moves, nil
}
