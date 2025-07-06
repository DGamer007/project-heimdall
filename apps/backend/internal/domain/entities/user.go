package entities

import "time"

type User struct {
	Id        string
	Email     string
	UserName  string
	Password  string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	FindOneById(id string) (User, error)
	FindOneByIdentifier(identifier string) (User, error)
	CreateOne(user *User) (string, error)
	CheckIfExistsById(id string) (bool, error)
	CheckIfExistsByEmail(email string) (bool, error)
	CheckIfExistsByUserName(username string) (bool, error)
}
