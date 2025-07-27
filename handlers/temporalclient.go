package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

var (
	tClient client.Client
	mu      sync.RWMutex
)

// StartTemporalClient starts the client with reconnect monitoring
func StartTemporalClient() {
	go func() {
		for {
			c, err := client.Dial(client.Options{})
			if err != nil {
				log.Println("Temporal unavailable, retrying in 5s:", err)
				time.Sleep(5 * time.Second)
				continue
			}
			replaceClient(c)
			log.Println("Connected to Temporal")

			// Monitor health every 10 seconds
			for {
				time.Sleep(10 * time.Second)
				if err := healthCheck(c); err != nil {
					log.Println("Temporal connection unhealthy, reconnecting...", err)
					break
				}
			}
		}
	}()
}

// GetClient returns the current Temporal client or nil
func GetClient() client.Client {
	mu.RLock()
	defer mu.RUnlock()
	return tClient
}

// replaceClient safely swaps the current Temporal client
func replaceClient(c client.Client) {
	mu.Lock()
	if tClient != nil {
		tClient.Close()
	}
	tClient = c
	mu.Unlock()
}

// healthCheck pings Temporal for connection health
func healthCheck(c client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.WorkflowService().GetSystemInfo(ctx, &workflowservice.GetSystemInfoRequest{})
	return err
}
