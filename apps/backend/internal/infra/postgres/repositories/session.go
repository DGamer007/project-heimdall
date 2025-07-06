package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"heimdall/backend/internal/constants"
	"heimdall/backend/internal/domain/entities"
	internal_errors "heimdall/backend/pkg/errors"

	"github.com/lib/pq"
)

type sessionRepository struct {
	conn *sql.DB
}

func NewSessionRepository(conn *sql.DB) *sessionRepository {
	return &sessionRepository{
		conn: conn,
	}
}

var sessionUniqueColumns = map[string]bool{
	"id":               true,
	"access_token_jti": true,
	"refresh_token":    true,
}

func (repo *sessionRepository) CreateOne(session *entities.Session) (string, error) {
	query := `
		INSERT INTO
			core.sessions
			(
				user_id,
				access_token_jti,
				refresh_token,
				expires_at
			)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id string

	// Convert Unix timestamp (int64) to time.Time for Postgres
	expiresAtTime := time.Unix(session.ExpiresAt, 0)

	row := repo.conn.QueryRow(query, session.UserId, session.AccessTokenJTI, session.RefreshToken, expiresAtTime)
	err := row.Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {

			case constants.PG_UNIQUE_VIOLATION_ERR_CODE:
				if pqErr.Constraint == "sessions_refresh_token_key" {
					return "", internal_errors.NewDuplicateResourceError("Session", "refresh_token", session.RefreshToken)
				}
				if pqErr.Constraint == "sessions_access_token_jti_key" {
					return "", internal_errors.NewDuplicateResourceError("Session", "access_token_jti", session.AccessTokenJTI)
				}
				return "", internal_errors.NewDuplicateResourceError("Session", pqErr.Column, pqErr.Detail)

			case constants.PG_CHECK_VIOLATION_ERR_CODE:
				return "", internal_errors.NewValidationError("Session data violates Postgres constraints", map[string]any{
					"constraint": pqErr.Constraint,
					"detail":     pqErr.Detail,
				})

			case constants.PG_NOT_NULL_VIOLATION_ERR_CODE:
				return "", internal_errors.NewValidationError("Required session field cannot be null", map[string]any{
					"column": pqErr.Column,
					"table":  "sessions",
				})

			case constants.PG_FOREIGN_KEY_VIOLATION_ERR_CODE:
				if pqErr.Constraint == "sessions_user_id_fkey" {
					return "", internal_errors.NewValidationError("Referenced user does not exist", map[string]any{
						"user_id": session.UserId,
					})
				}

				return "", internal_errors.NewValidationError("Referenced record does not exist", map[string]any{
					"constraint": pqErr.Constraint,
					"detail":     pqErr.Detail,
				})
			}
		}

		return "", internal_errors.NewInfrastructureError("Failed to create session in database", err)
	}

	return id, nil
}

func (repo *sessionRepository) DeleteOneByAccessTokenJTI(accessTokenJTI string) (bool, error) {
	return repo.deleteOneByUniqueColumn("access_token_jti", accessTokenJTI)
}

func (repo *sessionRepository) CheckIfExistsByIdentifier(userId string, accessTokenJTI string, refreshToken string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM
				core.sessions
			WHERE
				user_id = $1
				AND
				access_token_jti = $2
				AND
				refresh_token = $3
		)
	`

	var exists bool
	row := repo.conn.QueryRow(query, userId, accessTokenJTI, refreshToken)
	err := row.Scan(&exists)
	if err != nil {
		return false, internal_errors.NewInfrastructureError("Postgres query failed while checking session existence", err)
	}

	return exists, nil
}

func (repo *sessionRepository) ReplaceSession(oldAccessTokenJTI string, newSession *entities.Session) (string, error) {
	tx, err := repo.conn.Begin()
	if err != nil {
		return "", internal_errors.NewInfrastructureError("Failed to begin transaction", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	deleteQuery := `
		DELETE FROM
			core.sessions
		WHERE
			access_token_jti = $1
	`
	_, err = tx.Exec(deleteQuery, oldAccessTokenJTI)
	if err != nil {
		return "", internal_errors.NewInfrastructureError("Failed to delete old session", err)
	}

	insertQuery := `
		INSERT INTO
			core.sessions (
				user_id,
				access_token_jti,
				refresh_token,
				expires_at
			)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	// Convert Unix timestamp (int64) to time.Time for Postgres
	expiresAtTime := time.Unix(newSession.ExpiresAt, 0)

	var id string
	row := tx.QueryRow(insertQuery, newSession.UserId, newSession.AccessTokenJTI, newSession.RefreshToken, expiresAtTime)
	err = row.Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {

			case constants.PG_UNIQUE_VIOLATION_ERR_CODE:
				if pqErr.Constraint == "sessions_refresh_token_key" {
					return "", internal_errors.NewDuplicateResourceError("Session", "refresh_token", newSession.RefreshToken)
				}
				if pqErr.Constraint == "sessions_access_token_jti_key" {
					return "", internal_errors.NewDuplicateResourceError("Session", "access_token_jti", newSession.AccessTokenJTI)
				}
				return "", internal_errors.NewDuplicateResourceError("Session", pqErr.Column, pqErr.Detail)

			case constants.PG_CHECK_VIOLATION_ERR_CODE:
				return "", internal_errors.NewValidationError("Session data violates Postgres constraints", map[string]any{
					"constraint": pqErr.Constraint,
					"detail":     pqErr.Detail,
				})

			case constants.PG_NOT_NULL_VIOLATION_ERR_CODE:
				return "", internal_errors.NewValidationError("Required session field cannot be null", map[string]any{
					"column": pqErr.Column,
					"table":  "sessions",
				})

			case constants.PG_FOREIGN_KEY_VIOLATION_ERR_CODE:
				if pqErr.Constraint == "sessions_user_id_fkey" {
					return "", internal_errors.NewValidationError("Referenced user does not exist", map[string]any{
						"user_id": newSession.UserId,
					})
				}

				return "", internal_errors.NewValidationError("Referenced record does not exist", map[string]any{
					"constraint": pqErr.Constraint,
					"detail":     pqErr.Detail,
				})
			}
		}

		return "", internal_errors.NewInfrastructureError("Failed to create new session", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", internal_errors.NewInfrastructureError("Failed to commit transaction", err)
	}

	return id, nil
}

func (repo *sessionRepository) deleteOneByUniqueColumn(column string, value any) (bool, error) {
	if !sessionUniqueColumns[column] {
		return false, internal_errors.NewValidationError("Invalid column specified for session deletion", map[string]any{
			"column":          column,
			"allowed_columns": []string{"id", "access_token_jti", "refresh_token"},
		})
	}

	query := fmt.Sprintf(`
		DELETE FROM
			core.sessions
		WHERE
			%s = $1
	`, column)

	result, err := repo.conn.Exec(query, value)
	if err != nil {
		return false, internal_errors.NewInfrastructureError("Failed to delete session from Postgres", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, internal_errors.NewInfrastructureError("Failed to get rows affected count", err)
	}

	return rowsAffected > 0, nil
}
