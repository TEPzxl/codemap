package service

import "github.com/tepzxl/codemap/examples/layered-service/internal/repository"

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		repo: repository.NewUserRepository(),
	}
}

func (s *UserService) CreateUser(name string) error {
	return s.repo.Save(name)
}
