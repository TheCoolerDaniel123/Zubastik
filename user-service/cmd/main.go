package main

import (
	_ "github.com/lib/pq"
	"user-service/internal/config"
	"user-service/internal/service"
)

func main() {
	config := config.MustLoad()
	service.MustRun(config)
}
