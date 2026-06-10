package main

import (
	"fmt"
	"os"

	"github.com/AlexandrKudryavtsev/go-job-queue/config"
	"github.com/AlexandrKudryavtsev/go-job-queue/internal/app"
)

func main() {
	cfg, err := config.Load("./config/config.yaml")

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed load config: %v\n", err)
		os.Exit(1)
	}
	if err = cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}

	if err = app.Run(cfg); err != nil {
		os.Exit(1)
	}
}
