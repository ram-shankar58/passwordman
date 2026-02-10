package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"vault/internal/repository"
)

// AuditEvent represents an audit log event
type AuditEvent struct {
	UserID  int64
	EntryID int64
	Action  string
}

// AuditService handles background audit logging using goroutines and channels
type AuditService struct {
	eventChan chan AuditEvent
	repo      *repository.VaultRepository
	done      chan struct{}
	wg        sync.WaitGroup
}

func NewAuditService(repo *repository.VaultRepository) *AuditService {
	svc := &AuditService{
		eventChan: make(chan AuditEvent, 100), // buffered channel
		repo:      repo,
		done:      make(chan struct{}),
	}
	// Start background worker goroutine
	svc.wg.Add(1)
	go svc.auditWorker()
	return svc
}

// LogEvent sends an audit event to the channel (non-blocking)
func (s *AuditService) LogEvent(userID, entryID int64, action string) {
	select {
	case s.eventChan <- AuditEvent{UserID: userID, EntryID: entryID, Action: action}:
	case <-s.done:
		// Service is shutting down
	default:
		// Channel full, skip to avoid blocking
	}
}

// auditWorker is a background goroutine that processes audit events
func (s *AuditService) auditWorker() {
	defer s.wg.Done()
	for {
		select {
		case event := <-s.eventChan:
			// Process audit event (would save to DB in production)
			fmt.Printf("[AUDIT] User %d accessed entry %d: %s at %s\n",
				event.UserID, event.EntryID, event.Action, time.Now().Format(time.RFC3339))
		case <-s.done:
			return
		}
	}
}

// Shutdown gracefully stops the audit service
func (s *AuditService) Shutdown(ctx context.Context) error {
	close(s.done)
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WorkerPool processes multiple vault entries concurrently
type WorkerPool struct {
	workers  int
	jobsChan chan func()
	wg       sync.WaitGroup
	done     chan struct{}
}

func NewWorkerPool(workers int) *WorkerPool {
	wp := &WorkerPool{
		workers:  workers,
		jobsChan: make(chan func(), workers*2),
		done:     make(chan struct{}),
	}

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}

	return wp
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case job := <-wp.jobsChan:
			job()
		case <-wp.done:
			return
		}
	}
}

// Submit adds a job to the worker pool
func (wp *WorkerPool) Submit(job func()) {
	select {
	case wp.jobsChan <- job:
	case <-wp.done:
		// Pool is shutting down
	}
}

// Wait blocks until all jobs are done
func (wp *WorkerPool) Wait() {
	close(wp.jobsChan)
	wp.wg.Wait()
}

// Shutdown stops the worker pool
func (wp *WorkerPool) Shutdown() {
	close(wp.done)
	wp.Wait()
}
