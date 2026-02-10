package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"vault/internal/models"
	"vault/internal/repository"
)

type AuthService struct {
	users     *repository.UserRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

func (s *AuthService) Register(email, password string) (int64, error) {
	if email == "" || password == "" {
		return 0, errors.New("email and password required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	return s.users.Create(email, string(hash))
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	if email == "" || password == "" {
		return "", nil, errors.New("email and password required")
	}

	user, err := s.users.GetByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.tokenTTL).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return signed, user, nil
}
