package models

import (
	"database/sql"
	"errors"

	"backend/db"
)

// data-structure for user information  :
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// helper fn. for gettting user details by user_id : 
func GetUserByID(id int) (*User, error) {
	row := db.DB.QueryRow(`
        SELECT id, username, email, role
        FROM users
        WHERE id = $1
    `, id)

	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return u, nil
}
