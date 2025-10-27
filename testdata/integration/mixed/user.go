package main

import (
    "database/sql"
    "fmt"
)

type User struct {
    ID    int
    Name  string
    Email string
}

func (u *User) Validate() error {
    if u.ID <= 0 {
        return fmt.Errorf("invalid ID: %d", u.ID)
    }
    if u.Name == "" {
        return fmt.Errorf("name cannot be empty")
    }
    if u.Email == "" {
        return fmt.Errorf("email cannot be empty")
    }
    return nil
}

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(id int) (*User, error) {
    var user User
    err := r.db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
