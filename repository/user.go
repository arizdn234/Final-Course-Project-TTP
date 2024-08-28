package repository

import (
	"a21hc3NpZ25tZW50/db/filebased"
	"a21hc3NpZ25tZW50/model"
)

type UserRepository interface {
	GetUserByEmail(email string) (model.User, error)
	CreateUser(user model.User) (model.User, error)
	GetUserTaskCategory() ([]model.UserTaskCategory, error)
}

type userRepository struct {
	filebasedDb *filebased.Data
}

func NewUserRepo(filebasedDb *filebased.Data) *userRepository {
	return &userRepository{filebasedDb}
}

func (ur *userRepository) GetUserByEmail(email string) (model.User, error) {
	user, err := ur.filebasedDb.GetUserByEmail(email)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (ur *userRepository) CreateUser(user model.User) (model.User, error) {
	createdUser, err := ur.filebasedDb.CreateUser(user)
	if err != nil {
		return model.User{}, err
	}

	return createdUser, nil
}

func (ur *userRepository) GetUserTaskCategory() ([]model.UserTaskCategory, error) {
	userTasks, err := ur.filebasedDb.GetUserTaskCategory()
	if err != nil {
		return nil, err
	}

	return userTasks, nil
}
