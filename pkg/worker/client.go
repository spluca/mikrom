package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client for enqueueing tasks
type Client struct {
	client *asynq.Client
}

// NewClient creates a new worker client
func NewClient(redisAddr, redisPassword string, redisDB int) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return &Client{
		client: client,
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// EnqueueCreateVM enqueues a task to create a VM
func (c *Client) EnqueueCreateVM(payload *CreateVMPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeCreateVM, data)

	// Enqueue with retry options and timeout
	_, err = c.client.Enqueue(
		task,
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
		asynq.Queue("critical"), // High priority queue for VM creation
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueueDeleteVM enqueues a task to delete a VM
func (c *Client) EnqueueDeleteVM(payload *DeleteVMPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeDeleteVM, data)

	_, err = c.client.Enqueue(
		task,
		asynq.MaxRetry(3),
		asynq.Timeout(3*time.Minute),
		asynq.Queue("critical"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueueStartVM enqueues a task to start a VM
func (c *Client) EnqueueStartVM(payload *StartVMPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeStartVM, data)

	_, err = c.client.Enqueue(
		task,
		asynq.MaxRetry(3),
		asynq.Timeout(2*time.Minute),
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueueStopVM enqueues a task to stop a VM
func (c *Client) EnqueueStopVM(payload *StopVMPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeStopVM, data)

	_, err = c.client.Enqueue(
		task,
		asynq.MaxRetry(3),
		asynq.Timeout(2*time.Minute),
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueueRestartVM enqueues a task to restart a VM
func (c *Client) EnqueueRestartVM(payload *RestartVMPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeRestartVM, data)

	_, err = c.client.Enqueue(
		task,
		asynq.MaxRetry(3),
		asynq.Timeout(4*time.Minute), // Stop + Start
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}
