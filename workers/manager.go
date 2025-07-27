package workers

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/surajsub/temporal-rest-dsl/activities"
	"github.com/surajsub/temporal-rest-dsl/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerManager struct {
	client   client.Client
	// Future integration with vault
	secretID string
	roleID   string

	workers map[string]worker.Worker

	activeCount int
	mu          sync.Mutex
	workerIDs   map[string]string
}

func NewWorkerManager(c client.Client, secretid, roleid string) *WorkerManager {
	return &WorkerManager{

		client:    c,
		secretID:  secretid,
		roleID:    roleid,
		workers:   make(map[string]worker.Worker),
		workerIDs: make(map[string]string),
	}
}

func (m *WorkerManager) StartWorker(customername, queueName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workers[queueName]; exists {

		log.Printf("Worker for queue %s is already running for customer %s", queueName, customername)

		return
	}

	workerID := uuid.New().String()
	m.workerIDs[queueName] = workerID

	w := worker.New(m.client, queueName, worker.Options{})
	w.RegisterWorkflow(workflows.TemporalExecutorWorkflow) // Register your workflows
	w.RegisterActivity(activities.RunActivity) // Register your activities
	w.RegisterActivity(activities.DBActivity)

	w.RegisterActivity(activities.SaveStateToStorage) // Save it to local storage for every deployment to replay the delete flow.
	w.RegisterActivity(activities.LoadStateFromStorage)

	go func() {
		if err := w.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Worker for queue %s failed: %v", queueName, err)
		}
	}()

	m.workers[queueName] = w
	m.activeCount++
	log.Printf("Started worker %s for queue %s for customer [ %s ] ", workerID, queueName, customername)
}

func (m *WorkerManager) StopWorker(queueName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if w, exists := m.workers[queueName]; exists {
		w.Stop()
		delete(m.workers, queueName)
		m.activeCount-- // Decrement active worker count
		log.Printf("Stopped worker for queue %s", queueName)
	} else {
		log.Printf("Worker for queue %s is not running", queueName)
	}
}

// GetActiveWorkers returns the number of active workers.
func (m *WorkerManager) GetActiveWorkers() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeCount
}

// GetWorkerID returns the worker ID for a specific task queue.
func (m *WorkerManager) GetWorkerID(queueName string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.workerIDs[queueName]
}
