package service

import "errors"

type UserRepository interface {
	Save(name string) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(name string) error {
	if name == "" {
		return errors.New("name is empty")
	}
	return s.repo.Save(name)
}
