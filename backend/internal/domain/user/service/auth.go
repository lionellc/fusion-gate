package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lionellc/fusion-gate/internal/config"
	"github.com/lionellc/fusion-gate/internal/domain/user/client"
	"github.com/lionellc/fusion-gate/internal/domain/user/entity"
	appErrs "github.com/lionellc/fusion-gate/internal/errs"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, email, password, name string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, string, error)
	GenerateToken(user *entity.User) (string, error)
	ValidateToken(tokenString string) (jwt.MapClaims, error)
	GetById(ctx context.Context, id int64) (*entity.User, error)
}

type authService struct {
	cfg        *config.Config
	userClient client.UserClient
}

func NewAuthService(cfg *config.Config, userClient client.UserClient) AuthService {
	return &authService{cfg: cfg, userClient: userClient}
}

func (s *authService) Register(ctx context.Context, email, password, name string) (*entity.User, error) {
	// check is email already exists
	existing, err := s.userClient.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, appErrs.ErrUserEmailAlreadyExist
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		Role:      entity.RoleUser,
		Status:    entity.StatusActive,
		Balance:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.userClient.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*entity.User, string, error) {
	user, err := s.userClient.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", appErrs.ErrUserLogin
	}
	if user == nil {
		return nil, "", appErrs.ErrUserLogin
	}

	// verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", appErrs.ErrUserLogin
	}

	// verify status
	if !user.IsActive() {
		return nil, "", appErrs.ErrUserInactive
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *authService) GenerateToken(user *entity.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(s.cfg.JWT.Expire) * time.Second).Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *authService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, appErrs.ErrUserInvalidToken
}

func (s *authService) GetById(ctx context.Context, id int64) (*entity.User, error) {
	return s.userClient.GetById(ctx, id)
}
