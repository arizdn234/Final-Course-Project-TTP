package service

import (
	"a21hc3NpZ25tZW50/model"
	repo "a21hc3NpZ25tZW50/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type UserService interface {
	Register(user *model.User) (model.User, error)
	Login(user *model.User) (token *string, err error)
	GetUserTaskCategory() ([]model.UserTaskCategory, error)
}

type userService struct {
	userRepo     repo.UserRepository
	sessionsRepo repo.SessionRepository
}

func NewUserService(userRepository repo.UserRepository, sessionsRepo repo.SessionRepository) UserService {
	return &userService{userRepository, sessionsRepo}
}

func (us *userService) Register(user *model.User) (model.User, error) {
	dbUser, err := us.userRepo.GetUserByEmail(user.Email)
	if err != nil {
		return *user, err
	}

	if dbUser.Email != "" || dbUser.ID != 0 {
		return *user, errors.New("email already exists")
	}

	user.CreatedAt = time.Now()

	newUser, err := us.userRepo.CreateUser(*user)
	if err != nil {
		return *user, err
	}

	return newUser, nil
}

func (us *userService) Login(user *model.User) (token *string, err error) {
	dbUser, err := us.userRepo.GetUserByEmail(user.Email)
	if err != nil {
		return nil, err
	}

	if dbUser.Email == "" || dbUser.ID == 0 {
		return nil, errors.New("user not found")
	}

	if user.Password != dbUser.Password {
		return nil, errors.New("wrong email or password")
	}

	expirationTime := time.Now().Add(20 * time.Minute)
	claims := &model.Claims{
		Email: dbUser.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := t.SignedString(model.JwtKey)
	if err != nil {
		return nil, err
	}

	session := model.Session{
		Token:  tokenString,
		Email:  user.Email,
		Expiry: expirationTime,
	}

	_, err = us.sessionsRepo.SessionAvailEmail(session.Email)
	if err != nil {
		err = us.sessionsRepo.AddSessions(session)
	} else {
		err = us.sessionsRepo.UpdateSessions(session)
	}

	return &tokenString, nil
}

func (us *userService) GetUserTaskCategory() ([]model.UserTaskCategory, error) {
	userTasks, err := us.userRepo.GetUserTaskCategory()
	if err != nil {
		return nil, err
	}

	return userTasks, nil
}
