package user

import (
	"context"
	"errors"
	"inventory-it/internal/auth"

	"inventory-it/internal/pkg"

	"github.com/google/uuid"
)

type Usecase interface {
	Register(ctx context.Context, username, email, password, role string) error
	Login(ctx context.Context, email, password string) (string, error)
}

type usecase struct {
	repo      Repository
	jwtConfig auth.JwtConfig
}

func NewUsecase(r Repository, jwtConfig *auth.JwtConfig) Usecase {
	return &usecase{repo: r, jwtConfig: *jwtConfig}
}
func (u *usecase) Register(ctx context.Context, username, email, password, role string) error {
	//Instatiate object user
	var userCreate User

	//Validation and verification email
	_, isEmailRegistered, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if isEmailRegistered {
		return errors.New("Email already registered")
	}

	//Validation and verification username
	_, isUsernameRegistered, err := u.repo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if isUsernameRegistered {
		return errors.New("Username already registered!")
	}
	userCreate.Email = email
	userCreate.UserId = uuid.NewString()
	userCreate.Password = pkg.NewHashingPassword(password)
	userCreate.Role = role
	userCreate.Username = username
	return u.repo.Create(ctx, &userCreate)
}

func (u *usecase) Login(ctx context.Context, username, password string) (string, error) {
	user, isUsernameRegistered, err := u.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if !isUsernameRegistered {
		return "", errors.New("Username not registered")
	}
	if !pkg.ComparePassword(user.Password, password) {
		return "", errors.New("Invalid password")
	}
	token, err := u.jwtConfig.GenerateToken(user.UserId, user.Role, user.Email, user.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}
