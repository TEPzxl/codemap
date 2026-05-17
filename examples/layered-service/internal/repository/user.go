package repository

import "errors"

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Save(name string) error {
	if name == "" {
		return errors.New("name is empty")
	}
	return nil
}
