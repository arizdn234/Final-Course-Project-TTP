package service

import (
	"a21hc3NpZ25tZW50/model"
	repo "a21hc3NpZ25tZW50/repository"
)

type CategoryService interface {
	Store(category *model.Category) error
	Update(id int, category model.Category) error
	Delete(id int) error
	GetByID(id int) (*model.Category, error)
	GetList() ([]model.Category, error)
}

type categoryService struct {
	categoryRepository repo.CategoryRepository
}

func NewCategoryService(categoryRepository repo.CategoryRepository) CategoryService {
	return &categoryService{categoryRepository}
}

func (cs *categoryService) Store(category *model.Category) error {
	if err := cs.categoryRepository.Store(category); err != nil {
		return err
	}

	return nil
}

func (cs *categoryService) Update(id int, category model.Category) error {
	if err := cs.categoryRepository.Update(id, category); err != nil {
		return err
	}

	return nil
}

func (cs *categoryService) Delete(id int) error {
	if err := cs.categoryRepository.Delete(id); err != nil {
		return err
	}

	return nil
}

func (cs *categoryService) GetByID(id int) (*model.Category, error) {
	category, err := cs.categoryRepository.GetByID(id)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (cs *categoryService) GetList() ([]model.Category, error) {
	categories, err := cs.categoryRepository.GetList()
	if err != nil {
		return nil, err
	}

	return categories, nil
}
