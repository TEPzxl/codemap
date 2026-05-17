package main

import "github.com/tepzxl/codemap/examples/layered-service/internal/handler"

func main() {
	h := handler.NewUserHandler()
	_ = h.CreateUser("alice")
}
