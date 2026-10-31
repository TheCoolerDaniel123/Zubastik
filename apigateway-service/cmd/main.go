package main

import (
	"apigateway-service/internal/config"
	"apigateway-service/internal/service"
)

func main() {
	cfg := config.MustLoad()
	service.MustRun(cfg)
}
