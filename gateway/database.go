package gateway

import (
	// "database/sql"
)

type User struct {
	ID string
	Username string
}

func GetUser(id string) (User, error) {
	return User{}, nil
}