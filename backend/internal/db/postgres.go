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

// GetAllCards fetches all cards from the database and maps them to the Card struct.
func GetAllCards() ([]models.Card, error) {
	rows, err := DB.Query(`SELECT id, name, strength, default_spawn, rarity, power1, power2, notes FROM cards`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []models.Card
	for rows.Next() {
		var c models.Card
		var defaultSpawn sql.NullInt64
		var notes sql.NullString
		// Only map available fields
		err := rows.Scan(&c.ID, &c.Name, &c.Strength, &defaultSpawn, &c.Type, &c.Description, &c.CoverURL, &notes)
		if err != nil {
			return nil, err
		}
		if defaultSpawn.Valid {
			c.DefaultAmountPerSession = int(defaultSpawn.Int64)
		}
		if notes.Valid {
			c.Description = notes.String
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// GetCard fetches a single card by ID.
func GetCard(id string) (models.Card, error) {
	var c models.Card
	row := DB.QueryRow(`SELECT id, name, strength, default_spawn, rarity, power1, power2, notes FROM cards WHERE id = $1`, id)
	var defaultSpawn sql.NullInt64
	var notes sql.NullString
	err := row.Scan(&c.ID, &c.Name, &c.Strength, &defaultSpawn, &c.Type, &c.Description, &c.CoverURL, &notes)
	if err != nil {
		return c, err
	}
	if defaultSpawn.Valid {
		c.DefaultAmountPerSession = int(defaultSpawn.Int64)
	}
	if notes.Valid {
		c.Description = notes.String
	}
	return c, nil
}

// GetPower fetches a single power by name.
func GetPower(name string) (models.Power, error) {
	var p models.Power
	row := DB.QueryRow(`SELECT name, cost, target, description, is_passive, trigger FROM powers WHERE name = $1`, name)
	err := row.Scan(&p.Name, &p.Cost, &p.Target, &p.Description, &p.Type, &p.Trigger)
	if err != nil {
		return p, err
	}
	return p, nil
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
