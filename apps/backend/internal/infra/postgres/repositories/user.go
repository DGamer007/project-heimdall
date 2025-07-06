package repositories

import (
	"database/sql"
	"fmt"

	"heimdall/backend/internal/constants"
	"heimdall/backend/internal/domain/entities"
	internal_errors "heimdall/backend/pkg/errors"
	internal_string "heimdall/backend/pkg/string"

	"github.com/lib/pq"
)

type userRepository struct {
	conn *sql.DB
}

func NewUserRepository(conn *sql.DB) *userRepository {
	return &userRepository{
		conn: conn,
	}
}

var userUniqueColumns = map[string]bool{
	"email":    true,
	"username": true,
	"id":       true,
}

func (repo *userRepository) FindOneById(id string) (entities.User, error) {
	return repo.findOneByUniqueColumn("id", id)
}

func (repo *userRepository) FindOneByIdentifier(identifier string) (entities.User, error) {
	var identifierType string
	if internal_string.IsValidEmail(identifier) {
		identifierType = "email"
	} else {
		identifierType = "username"
	}

	return repo.findOneByUniqueColumn(identifierType, identifier)
}

func (repo *userRepository) CreateOne(user *entities.User) (string, error) {
	query := `
		INSERT INTO
			core.users
			(
				email,
				username,
				password,
				first_name,
				last_name
			)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id string

	row := repo.conn.QueryRow(query, user.Email, user.UserName, user.Password, user.FirstName, user.LastName)
	err := row.Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {

			case constants.PG_UNIQUE_VIOLATION_ERR_CODE:
				if pqErr.Constraint == "users_email_key" {
					return "", internal_errors.NewDuplicateResourceError("User", "email", user.Email)
				}
				if pqErr.Constraint == "users_username_key" {
					return "", internal_errors.NewDuplicateResourceError("User", "email", user.UserName)
				}

				return "", internal_errors.NewDuplicateResourceError("User", pqErr.Column, pqErr.Detail)

			case constants.PG_CHECK_VIOLATION_ERR_CODE:
				return "", internal_errors.NewValidationError("User data violates Postgres constraints", map[string]any{
					"constraint": pqErr.Constraint,
					"detail":     pqErr.Detail,
					"table":      "users",
				})

			case constants.PG_NOT_NULL_VIOLATION_ERR_CODE:
				return "", internal_errors.NewValidationError("Required field is missing", map[string]any{
					"column": pqErr.Column,
					"table":  "users",
				})

			case constants.PG_FOREIGN_KEY_VIOLATION_ERR_CODE:
				return "", internal_errors.NewValidationError("Referenced record does not exist", map[string]any{
					"constraint": pqErr.Constraint,
					"detail":     pqErr.Detail,
				})

			case constants.PG_STRING_DATA_RIGHT_TRUNCATION_ERR_CODE:
				return "", internal_errors.NewValidationError("Data too long for Postgres column", map[string]any{
					"column": pqErr.Column,
					"table":  "users",
					"detail": pqErr.Detail,
				})

			}

		}

		return "", internal_errors.NewInfrastructureError("Failed to create user in Postgres", err)
	}

	return id, nil
}

func (repo *userRepository) CheckIfExistsByEmail(email string) (bool, error) {
	return repo.checkIfExistsByUniqueColumn("email", email)
}

func (repo *userRepository) CheckIfExistsByUserName(username string) (bool, error) {
	return repo.checkIfExistsByUniqueColumn("username", username)
}

func (repo *userRepository) CheckIfExistsById(id string) (bool, error) {
	return repo.checkIfExistsByUniqueColumn("id", id)
}

func (repo *userRepository) findOneByUniqueColumn(column string, value any) (entities.User, error) {

	if !userUniqueColumns[column] {
		return entities.User{}, internal_errors.NewValidationError("Invalid column specified for user query", map[string]any{
			"column":          column,
			"allowed_columns": []string{"id", "email", "username"},
		})
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			email,
			username,
			password,
			first_name,
			last_name,
			created_at,
			updated_at
		FROM
			core.users
		WHERE
			%s = $1
	`, column)

	var user entities.User

	row := repo.conn.QueryRow(query, value)
	err := row.Scan(&user.Id, &user.Email, &user.UserName, &user.Password, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.User{}, internal_errors.NewNotFoundError("user", fmt.Sprintf("%v", value))
		}
		return entities.User{}, internal_errors.NewInfrastructureError("Postgres query failed while fetching user", err)
	}

	return user, nil
}

func (repo *userRepository) checkIfExistsByUniqueColumn(column string, value any) (bool, error) {
	if !userUniqueColumns[column] {
		return false, internal_errors.NewValidationError("Invalid column specified for user existence check", map[string]any{
			"column":          column,
			"allowed_columns": []string{"id", "email", "username"},
		})
	}

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM
				core.users
			WHERE
				%s = $1
		)
	`, column)

	var exists bool
	row := repo.conn.QueryRow(query, value)
	err := row.Scan(&exists)
	if err != nil {
		return false, internal_errors.NewInfrastructureError("Postgres query failed while checking user existence", err)
	}

	return exists, nil
}
