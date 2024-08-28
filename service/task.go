package service

import (
	"a21hc3NpZ25tZW50/model"
	repo "a21hc3NpZ25tZW50/repository"
)

type TaskService interface {
	Store(task *model.Task) error
	Update(id int, task *model.Task) error
	Delete(id int) error
	GetByID(id int) (*model.Task, error)
	GetList() ([]model.Task, error)
	GetTaskCategory(id int) ([]model.TaskCategory, error)
}

type taskService struct {
	taskRepository repo.TaskRepository
}

func NewTaskService(taskRepository repo.TaskRepository) TaskService {
	return &taskService{taskRepository}
}

func (ts *taskService) Store(task *model.Task) error {
	if err := ts.taskRepository.Store(task); err != nil {
		return err
	}

	return nil
}

func (ts *taskService) Update(id int, task *model.Task) error {
	if err := ts.taskRepository.Update(id, task); err != nil {
		return err
	}

	return nil
}

func (ts *taskService) Delete(id int) error {
	if err := ts.taskRepository.Delete(id); err != nil {
		return err
	}

	return nil
}

func (ts *taskService) GetByID(id int) (*model.Task, error) {
	task, err := ts.taskRepository.GetByID(id)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (ts *taskService) GetList() ([]model.Task, error) {
	tasks, err := ts.taskRepository.GetList()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (ts *taskService) GetTaskCategory(id int) ([]model.TaskCategory, error) {
	task, err := ts.taskRepository.GetTaskCategory(id)
	if err != nil {
		return nil, err
	}

	return task, nil
}
