package users

import (
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (us *UserService) RegisterUser(name, email, password string, role string, bio string) (*User, error) {
	var count int
	err := us.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	_, err = us.db.Exec("INSERT INTO users (id, name, email, password, role, bio) VALUES ($1, $2, $3, $4, $5, $6)",
		id, name, email, string(hashed), role, bio)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:    id,
		Name:  name,
		Email: email,
		Role:  role,
		Bio:   bio,
	}, nil
}

func (us *UserService) LoginUser(email, password string) (*User, error) {
	row := us.db.QueryRow("SELECT id, name, password, role, bio FROM users WHERE email = $1", email)
	var id, name, hashedPwd, role, bio string
	if err := row.Scan(&id, &name, &hashedPwd, &role, &bio); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid email or password")
		}
		log.Printf("Database error during login: %v", err)
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: hashedPwd,
		Role:     role,
		Bio:      bio,
	}, nil
}
