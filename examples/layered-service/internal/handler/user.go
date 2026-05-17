package handler

import "github.com/tepzxl/codemap/examples/layered-service/internal/service"

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		service: service.NewUserService(),
	}
}

func (h *UserHandler) CreateUser(name string) error {
	return h.service.CreateUser(name)
}
