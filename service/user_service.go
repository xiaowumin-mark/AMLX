package service

import (
	"context"
	"errors"
	"strings"

	"github.com/xiaowumin-mark/AMLX/model"
	"github.com/xiaowumin-mark/AMLX/store"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
	ErrInvalidInput = errors.New("invalid input")
)

type CreateUserRequest struct {
	Name     string
	Email    string
	Password string
	RoleID   uint
}

type UpdateUserRequest struct {
	Name     *string
	Email    *string
	Password *string
	RoleID   *uint
}

type UserService interface {
	Create(ctx context.Context, req CreateUserRequest) (*model.Users, error)
	GetByID(ctx context.Context, id uint) (*model.Users, error)
	GetByEmail(ctx context.Context, email string) (*model.Users, error)
	Update(ctx context.Context, id uint, req UpdateUserRequest) (*model.Users, error)
	SetBan(ctx context.Context, id uint, ban bool) error
}

type userService struct {
	users store.UserStore
	cost  int
}

func NewUserService(users store.UserStore, bcryptCost int) UserService {
	return &userService{users: users, cost: bcryptCost}
}

func (s *userService) Create(ctx context.Context, req CreateUserRequest) (*model.Users, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	if req.Name == "" || req.Email == "" || req.Password == "" || req.RoleID == 0 {
		return nil, ErrInvalidInput
	}

	if _, err := s.users.GetByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := HashPassword(req.Password, s.cost)
	if err != nil {
		return nil, err
	}
	user := &model.Users{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		RoleId:   req.RoleID,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uint) (*model.Users, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	user, err := s.users.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*model.Users, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	if email == "" {
		return nil, ErrInvalidInput
	}
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (s *userService) Update(ctx context.Context, id uint, req UpdateUserRequest) (*model.Users, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	if req.Name == nil && req.Email == nil && req.Password == nil && req.RoleID == nil {
		return nil, ErrInvalidInput
	}

	user, err := s.users.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)
		if value == "" {
			return nil, ErrInvalidInput
		}
		user.Name = value
	}
	if req.Email != nil {
		value := strings.TrimSpace(strings.ToLower(*req.Email))
		if value == "" {
			return nil, ErrInvalidInput
		}
		if value != user.Email {
			if _, err := s.users.GetByEmail(ctx, value); err == nil {
				return nil, ErrEmailExists
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		}
		user.Email = value
	}
	if req.Password != nil {
		value := strings.TrimSpace(*req.Password)
		if value == "" {
			return nil, ErrInvalidInput
		}
		hashedPassword, err := HashPassword(value, s.cost)
		if err != nil {
			return nil, err
		}
		user.Password = hashedPassword
	}
	if req.RoleID != nil {
		if *req.RoleID == 0 {
			return nil, ErrInvalidInput
		}
		user.RoleId = *req.RoleID
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) SetBan(ctx context.Context, id uint, ban bool) error {
	if id == 0 {
		return ErrInvalidInput
	}
	return s.users.SetBan(ctx, id, ban)
}
