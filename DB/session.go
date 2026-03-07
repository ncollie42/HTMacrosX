package database

import "time"

func CreateSession(sessionID string, userID int, expiresAt time.Time) error {
	_, err := sqlDB.Exec(
		`INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)`,
		sessionID, userID, expiresAt.Format(time.RFC3339),
	)
	return err
}

func GetSessionUserID(sessionID string) (int, error) {
	var userID int
	err := sqlDB.QueryRow(
		`SELECT user_id FROM sessions WHERE session_id = ? AND expires_at > ?`,
		sessionID, time.Now().Format(time.RFC3339),
	).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func SetSessionToken(sessionID string, token string) error {
	_, err := sqlDB.Exec(`UPDATE sessions SET token = ? WHERE session_id = ?`, token, sessionID)
	return err
}

func GetSessionToken(sessionID string) (string, error) {
	var token string
	err := sqlDB.QueryRow(`SELECT token FROM sessions WHERE session_id = ?`, sessionID).Scan(&token)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ClearSessionToken(sessionID string) error {
	return SetSessionToken(sessionID, "")
}

func DeleteSession(sessionID string) error {
	_, err := sqlDB.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID)
	return err
}

func CleanExpiredSessions() error {
	_, err := sqlDB.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().Format(time.RFC3339))
	return err
}
