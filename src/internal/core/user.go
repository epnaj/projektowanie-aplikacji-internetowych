package core

import (
	"context"
	"errors"
	"strings"
	"time"
)

type UserService struct {
	users  UserRepository
	hasher PasswordHasher
}

func NewUserService(users UserRepository, hasher PasswordHasher) *UserService {
	return &UserService{
		users:  users,
		hasher: hasher,
	}
}

// Register creates a new user from a plaintext password; the ID is
// assigned by the repository on insert
func (s *UserService) Register(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	_, err := s.users.GetByEmail(ctx, email)
	switch {
	case err == nil:
		return nil, ErrConflict
	case !errors.Is(err, ErrNotFound):
		return nil, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	return user, s.users.Create(ctx, user)
}

// Authenticate returns ErrInvalidCredentials for both unknown email and
// wrong password, so callers cannot probe which emails are registered.
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !s.hasher.Compare(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *UserService) Get(ctx context.Context, id ID) (*User, error) {
	return s.users.GetById(ctx, id)
}

func (s *UserService) Delete(ctx context.Context, id ID) error {
	return s.users.Delete(ctx, id)
}
