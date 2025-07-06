package main

import (
	"fmt"
	"log"
	"os"

	"heimdall/backend/internal/config"
	"heimdall/backend/internal/infra"
	"heimdall/backend/internal/infra/postgres"
	"heimdall/backend/internal/interfaces/web"
)

func printBanner() {
	fmt.Println(`
       ██╗  ██╗███████╗██╗███╗   ███╗██████╗  █████╗ ██╗     ██╗
██╗    ██║  ██║██╔════╝██║████╗ ████║██╔══██╗██╔══██╗██║     ██║         ██╗
╚═╝    ███████║█████╗  ██║██╔████╔██║██║  ██║███████║██║     ██║         ╚═╝
██╗    ██╔══██║██╔══╝  ██║██║╚██╔╝██║██║  ██║██╔══██║██║     ██║         ██╗
╚═╝    ██║  ██║███████╗██║██║ ╚═╝ ██║██████╔╝██║  ██║███████╗███████╗    ╚═╝
       ╚═╝  ╚═╝╚══════╝╚═╝╚═╝     ╚═╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝
	`)
}

func main() {
	printBanner()

	// Load configuration
	AppConfig, err := config.LoadConfig()
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	log.Println("Configuration loaded successfully")

	// Connect to Postgres
	Postgres, err := postgres.NewConnection(postgres.Config{
		Host:     AppConfig.Database.Postgres.Host,
		Port:     AppConfig.Database.Postgres.Port,
		User:     AppConfig.Database.Postgres.User,
		Password: AppConfig.Database.Postgres.Password,
		DBName:   AppConfig.Database.Postgres.DBName,
		SSLMode:  AppConfig.Database.Postgres.SSLMode,
	})

	if err != nil {
		log.Printf("Failed to connect to Postgres: %v", err)
		os.Exit(1)
	}

	log.Println("Postgres connection established successfully")

	// Start server
	server := web.NewServer(&infra.DataStore{
		Postgres: Postgres,
	})
	log.Printf("Starting server on port %s", AppConfig.Server.Port)

	if err := server.Run(":" + AppConfig.Server.Port); err != nil {
		log.Printf("Failed to start server: %v", err)

		// Clean up Postgres connection before exit
		if Postgres != nil {
			Postgres.Close()
			log.Println("Postgres connection closed")
		}

		os.Exit(1)
	}
}
