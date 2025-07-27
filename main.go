package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/surajsub/temporal-rest-dsl/db"
	"github.com/surajsub/temporal-rest-dsl/handlers"
	"github.com/surajsub/temporal-rest-dsl/workers"

	"go.temporal.io/sdk/client"

	"gopkg.in/yaml.v3"
)

func main() {
	handlers.StartTemporalClient()
	stopQueue := flag.String("stop", "", "Task queue to stop (optional)")
	flag.Parse()

	secretID, secretIDExists := os.LookupEnv("SECRET_ID") // Read "SECRET_ID" and check if it exists
	if !secretIDExists {

		fmt.Println("Error: secret_id environment variable is not set")
		os.Exit(1) // Exit the program with an error code
	} else {
		//fmt.Println("Secret ID:", secretID)
	}

	roleID, roleIDExists := os.LookupEnv("ROLE_ID") // Read "ROLE_ID" and check if it exists
	if !roleIDExists {
		fmt.Println("Error: role_id environment variable is not set")
		os.Exit(1) // Exit the program with an error code
	} else {
		//fmt.Println("Role ID:", roleID)
	}
	dbUser, dbUserExists := os.LookupEnv("POSTGRES_DB_USER")
	dbPassword, dbPasswordExists := os.LookupEnv("POSTGRES_DB_PASSWORD")
	dbName, dbNameExists := os.LookupEnv("POSTGRES_DB_NAME")
	
	if !dbUserExists {
		fmt.Println("Error: POSTGRES_DB_USER environment variable is not set")
		os.Exit(1) // Exit the program with an error code
	} else {
		//fmt.Println("Role ID:", roleID)
	}

	if !dbPasswordExists {
		fmt.Println("Error: POSTGRES_DB_PASSWORD environment variable is not set")
		os.Exit(1) // Exit the program with an error code
	} else {
		//fmt.Println("Role ID:", roleID)
	}
	if !dbNameExists {
		fmt.Println("Error: POSTGRES_DB_NAME environment variable is not set")
		os.Exit(1) // Exit the program with an error code
	} else {
		//fmt.Println("Role ID:", roleID)
	}

	// Load customer configuration
	config, err := LoadConfig("customers.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Postgres DB
	if err := db.InitDB(dbUser,dbPassword, dbName); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Connect to Temporal

	c, err := client.Dial(client.Options{})

	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Initialize WorkerManager
	manager := workers.NewWorkerManager(c, secretID, roleID)

	// Handle worker stop request
	if *stopQueue != "" {
		manager.StopWorker(*stopQueue)
		log.Printf("Stopped worker for task queue: %s\n", *stopQueue)
		os.Exit(0)
	}

	// Start workers for each customer
	for _, customer := range config.Customers {
		manager.StartWorker(customer.Name, customer.TaskQueue)
	}

	// Start the Echo REST API server
	go func() {
		e := echo.New()
		e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler
		e.Use(handlers.RequestIDMiddleware)
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		handlers.RegisterRoutes(e, func() client.Client {
			return handlers.GetClient()
		})

		if err := e.Start(":8080"); err != nil {
			log.Fatalf("Failed to start Echo server: %v", err)
		}
	}()

	// Wait for shutdown signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	log.Println("Shutting down gracefully...")
	// You can add cleanup here if needed
}

type CustomerConfig struct {
	Customers []struct {
		Name      string `yaml:"name"`
		TaskQueue string `yaml:"task_queue"`
	} `yaml:"customers"`
}

func LoadConfig(filePath string) (*CustomerConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read YAML file")
		return nil, err
	}

	var config CustomerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to parse YAML file")
		return nil, err
	}

	return &config, nil
}
