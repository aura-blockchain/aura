package main

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

// Queue manages verification task queue with priority
type Queue struct {
	mu    sync.RWMutex
	items PriorityQueue
	tasks map[string]*Task
}

// Task represents a verification task
type Task struct {
	ID           string
	Type         string
	Priority     int
	CredentialID string
	RequesterID  string
	AssignedTo   string
	Status       TaskStatus
	CreatedAt    time.Time
	AssignedAt   *time.Time
	CompletedAt  *time.Time
	Deadline     time.Time
	Data         map[string]interface{}
	index        int // Used by heap
}

// TaskStatus represents task status
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskAssigned   TaskStatus = "assigned"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
	TaskRejected   TaskStatus = "rejected"
)

// PriorityQueue implements heap.Interface
type PriorityQueue []*Task

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Higher priority first, then earlier deadline
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	return pq[i].Deadline.Before(pq[j].Deadline)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	task := x.(*Task)
	task.index = n
	*pq = append(*pq, task)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	task := old[n-1]
	old[n-1] = nil
	task.index = -1
	*pq = old[0 : n-1]
	return task
}

// NewQueue creates a new task queue
func NewQueue() *Queue {
	q := &Queue{
		items: make(PriorityQueue, 0),
		tasks: make(map[string]*Task),
	}
	heap.Init(&q.items)
	return q
}

// Enqueue adds a new task to the queue
func (q *Queue) Enqueue(task *Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if task.ID == "" {
		task.ID = generateTaskID()
	}

	if _, exists := q.tasks[task.ID]; exists {
		return errors.New("task already exists")
	}

	task.Status = TaskPending
	task.CreatedAt = time.Now()

	if task.Deadline.IsZero() {
		task.Deadline = time.Now().Add(24 * time.Hour)
	}

	heap.Push(&q.items, task)
	q.tasks[task.ID] = task

	return nil
}

// Dequeue gets the highest priority task
func (q *Queue) Dequeue() (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.items.Len() == 0 {
		return nil, errors.New("queue is empty")
	}

	task := heap.Pop(&q.items).(*Task)
	return task, nil
}

// Assign assigns a task to a verifier
func (q *Queue) Assign(taskID, verifierID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	if task.Status != TaskPending {
		return errors.New("task is not pending")
	}

	now := time.Now()
	task.AssignedTo = verifierID
	task.Status = TaskAssigned
	task.AssignedAt = &now

	return nil
}

// StartTask marks a task as in progress
func (q *Queue) StartTask(taskID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	if task.Status != TaskAssigned {
		return errors.New("task is not assigned")
	}

	task.Status = TaskInProgress

	return nil
}

// CompleteTask marks a task as completed
func (q *Queue) CompleteTask(taskID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	if task.Status != TaskInProgress {
		return errors.New("task is not in progress")
	}

	now := time.Now()
	task.Status = TaskCompleted
	task.CompletedAt = &now

	return nil
}

// RejectTask marks a task as rejected
func (q *Queue) RejectTask(taskID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}

	now := time.Now()
	task.Status = TaskRejected
	task.CompletedAt = &now

	return nil
}

// GetTask retrieves a task by ID
func (q *Queue) GetTask(taskID string) (*Task, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	task, exists := q.tasks[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

// GetTasksByVerifier gets all tasks assigned to a verifier
func (q *Queue) GetTasksByVerifier(verifierID string) []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range q.tasks {
		if task.AssignedTo == verifierID {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// GetTasksByStatus gets all tasks with a specific status
func (q *Queue) GetTasksByStatus(status TaskStatus) []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range q.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// GetPendingTasks returns all pending tasks
func (q *Queue) GetPendingTasks() []*Task {
	return q.GetTasksByStatus(TaskPending)
}

// GetQueueDepth returns the number of pending tasks
func (q *Queue) GetQueueDepth() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.items.Len()
}

// GetStatistics returns queue statistics
func (q *Queue) GetStatistics() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	statusCounts := make(map[TaskStatus]int)
	overdueTasks := 0
	now := time.Now()

	for _, task := range q.tasks {
		statusCounts[task.Status]++

		if task.Status != TaskCompleted && task.Status != TaskRejected {
			if now.After(task.Deadline) {
				overdueTasks++
			}
		}
	}

	return map[string]interface{}{
		"total_tasks":   len(q.tasks),
		"pending":       statusCounts[TaskPending],
		"assigned":      statusCounts[TaskAssigned],
		"in_progress":   statusCounts[TaskInProgress],
		"completed":     statusCounts[TaskCompleted],
		"rejected":      statusCounts[TaskRejected],
		"overdue_tasks": overdueTasks,
		"queue_depth":   q.items.Len(),
	}
}

// AutoAssign automatically assigns tasks to available verifiers
func (q *Queue) AutoAssign(verifiers []*Verifier) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	assigned := 0

	// Get pending tasks
	for q.items.Len() > 0 {
		task := heap.Pop(&q.items).(*Task)

		// Find best verifier for this task
		verifier := q.findBestVerifier(task, verifiers)
		if verifier == nil {
			// Put task back in queue
			heap.Push(&q.items, task)
			break
		}

		// Assign task
		now := time.Now()
		task.AssignedTo = verifier.ID
		task.Status = TaskAssigned
		task.AssignedAt = &now

		assigned++
	}

	return assigned, nil
}

// findBestVerifier finds the best verifier for a task
func (q *Queue) findBestVerifier(task *Task, verifiers []*Verifier) *Verifier {
	var bestVerifier *Verifier
	bestScore := -1

	for _, verifier := range verifiers {
		if !verifier.IsActive || verifier.Status != StatusActive {
			continue
		}

		// Check if verifier has the required specialization
		hasSpecialization := false
		for _, spec := range verifier.Specializations {
			if spec == task.Type {
				hasSpecialization = true
				break
			}
		}

		if !hasSpecialization {
			continue
		}

		// Calculate score based on reputation and current workload
		currentTasks := q.countVerifierTasks(verifier.ID)
		score := verifier.ReputationScore - (currentTasks * 10)

		if score > bestScore {
			bestScore = score
			bestVerifier = verifier
		}
	}

	return bestVerifier
}

// countVerifierTasks counts active tasks for a verifier
func (q *Queue) countVerifierTasks(verifierID string) int {
	count := 0
	for _, task := range q.tasks {
		if task.AssignedTo == verifierID &&
			(task.Status == TaskAssigned || task.Status == TaskInProgress) {
			count++
		}
	}
	return count
}

func generateTaskID() string {
	return "task_" + generateRandomString(12)
}
