package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nersus15/mini-proxy/proxy"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
	"github.com/webcore-go/webcore/infra/config"
)

func main() {
	// Validasi argumen CLI
	if len(os.Args) < 2 {
		log.Fatal("Usage: webcore [proxy]")
	}

	service := os.Args[1]
	validServices := map[string]bool{
		"proxy": true,
	}

	if !validServices[service] {
		os.Exit(1)
	}

	ctx := context.Background()

	// Load configuration
	cfg := config.Config{}
	if err := config.LoadDefaultConfig(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	out.SetEnvironment(cfg.App.Environment)

	var libraries map[string]core.LibraryLoader
	var packages []core.Module

	switch service {
	case "proxy":
		if cfg.Server.Port == 2021 {
			cfg.Server.Port = 2022
		}
		libraries = proxy.APP_LIBRARIES
		packages = proxy.APP_PACKAGES
	}

	// Initialize application
	application := core.NewApp(ctx, &cfg, libraries, packages)

	// Start the application
	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	fmt.Printf("Service '%s' started successfully\n", service)
}
