package main

import (
	"github.com/tepzxl/codemap/examples/interface-call/repository"
	"github.com/tepzxl/codemap/examples/interface-call/service"
)

func main() {
	repo := repository.NewMemoryUserRepository()
	svc := service.NewUserService(repo)
	_ = svc.CreateUser("alice")
}
