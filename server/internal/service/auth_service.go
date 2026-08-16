package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

var (
	ErrUsernameTaken = errors.New("用户名已被占用")
	ErrEmailTaken    = errors.New("邮箱已被注册")
	ErrInvalidCreds  = errors.New("用户名或密码错误")
)

// AuthService 认证服务
type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

// Register 注册
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*model.User, string, error) {
	// 校验用户名
	if exists, err := s.userRepo.UsernameExists(ctx, username); err != nil {
		return nil, "", err
	} else if exists {
		return nil, "", ErrUsernameTaken
	}
	if exists, err := s.userRepo.EmailExists(ctx, email); err != nil {
		return nil, "", err
	} else if exists {
		return nil, "", ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(user.ID)
	return user, token, err
}

// Login 登录
func (s *AuthService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCreds
	}

	token, err := s.generateToken(user.ID)
	return user, token, err
}

// GetProfile 获取用户信息 + 统计
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}
	stats, err := s.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &model.UserProfile{User: *user, UserStats: *stats}, nil
}

// UpdateProfile 更新用户信息
func (s *AuthService) UpdateProfile(ctx context.Context, userID string, username, email, password, avatarURL string) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}

	if username != "" {
		if exists, err := s.userRepo.UsernameExists(ctx, username); err != nil {
			return nil, err
		} else if exists && user.Username != username {
			return nil, ErrUsernameTaken
		}
		user.Username = username
	}
	if email != "" {
		if exists, err := s.userRepo.EmailExists(ctx, email); err != nil {
			return nil, err
		} else if exists && user.Email != email {
			return nil, ErrEmailTaken
		}
		user.Email = email
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hash)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken 校验 token 并返回 user_id（供中间件使用）
func (s *AuthService) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if uid, ok := claims["user_id"].(string); ok {
			return uid, nil
		}
	}
	return "", errors.New("invalid token claims")
}
