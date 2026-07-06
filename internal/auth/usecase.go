package auth

import (
	"context"
	"errors"
	"strings"

	"inventory-it/internal/pkg"

	"github.com/google/uuid"
)

type Usecase interface {
	Register(ctx context.Context, user User) (User, error)
	Login(ctx context.Context, email, password string) (User, string, error)
	UpdatePasswordById(ctx context.Context, userId string, newPassword string) error
}

type usecase struct {
	repo      Repository
	jwtConfig pkg.JwtConfig
}

func NewUsecase(r Repository, jwtConfig *pkg.JwtConfig) Usecase {
	return &usecase{repo: r, jwtConfig: *jwtConfig}
}
func (u *usecase) Register(ctx context.Context, user User) (User, error) {
	//Instatiate object user
	var userCreate User

	//Validation and verification email
	_, isEmailRegistered, err := u.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		return User{}, err
	}
	if isEmailRegistered {
		return User{}, errors.New("Email already registered")
	}

	//Validation and verification username
	_, isUsernameRegistered, err := u.repo.FindByUsername(ctx, user.Username)
	if err != nil {
		return User{}, err
	}
	if isUsernameRegistered {
		return User{}, errors.New("Username already registered!")
	}
	userCreate.Email = user.Email
	userCreate.UserId = uuid.NewString()
	userCreate.Password = pkg.NewHashingPassword(user.Password)
	userCreate.Role = user.Role
	userCreate.Department_id = strings.ToUpper(user.Department_id)
	userCreate.Username = user.Username
	return userCreate, u.repo.Create(ctx, &userCreate)

}

func (u *usecase) Login(ctx context.Context, usernameOrEmail, password string) (User, string, error) {
	var userRegistered User
	var isRegistered bool

	if pkg.IsValidEmail(usernameOrEmail) {
		user, registered, err := u.repo.FindByEmail(ctx, usernameOrEmail)
		if err != nil {
			return User{}, "", err
		}
		isRegistered = registered
		userRegistered = user
	} else {
		user, registered, err := u.repo.FindByUsername(ctx, usernameOrEmail)
		if err != nil {
			return User{}, "", err
		}

		isRegistered = registered
		userRegistered = user
	}

	if !isRegistered {
		return User{}, "", errors.New("invalid username or password")
	}

	if !pkg.ComparePassword(userRegistered.Password, password) {
		return User{}, "", errors.New("invalid username or password")
	}
	var loginResp userResponse

	token, err := u.jwtConfig.GenerateToken(
		userRegistered.UserId,
		userRegistered.Role,
		userRegistered.Email,
		userRegistered.Username,
	)
	loginResp.User = userRegistered
	loginResp.Token = token

	if err != nil {
		return User{}, "", err
	}

	return userRegistered, token, nil
}

func (u *usecase) UpdatePasswordById(ctx context.Context, userId string, newPassword string) error {
	hashedNewPassword := pkg.NewHashingPassword(newPassword)
	return u.repo.UpdatePasswordById(ctx, userId, hashedNewPassword)
}
