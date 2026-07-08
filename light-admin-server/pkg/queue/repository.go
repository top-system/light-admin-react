package queue

import (
	"context"
	"errors"

	"github.com/top-system/light-admin/pkg/uuid"
)

// ErrTaskNotFound is returned by TaskRepository implementations when no task
// matches the requested ID. It keeps pkg/queue free of any storage-framework
// error type (e.g. gorm.ErrRecordNotFound); callers match with errors.Is.
var ErrTaskNotFound = errors.New("queue: task not found")

// TaskRepository interface for task persistence
type TaskRepository interface {
	// Create creates a new task in database
	Create(ctx context.Context, task *TaskModel) error
	// Update updates an existing task in database
	Update(ctx context.Context, task *TaskModel) error
	// GetByID gets a task by ID
	GetByID(ctx context.Context, id uint64) (*TaskModel, error)
	// GetPendingTasks gets all pending tasks by types
	GetPendingTasks(ctx context.Context, types ...string) ([]*TaskModel, error)
	// Delete deletes a task by ID
	Delete(ctx context.Context, id uint64) error
}

// TaskArgs represents arguments for creating or updating a task
type TaskArgs struct {
	Status        Status
	Type          string
	PublicState   *TaskPublicState
	PrivateState  string
	OwnerID       string
	CorrelationID uuid.UUID
}

// InMemoryTaskRepository implements TaskRepository using in-memory storage
type InMemoryTaskRepository struct {
	tasks  map[uint64]*TaskModel
	nextID uint64
}

// NewInMemoryTaskRepository creates a new in-memory task repository
func NewInMemoryTaskRepository() TaskRepository {
	return &InMemoryTaskRepository{
		tasks:  make(map[uint64]*TaskModel),
		nextID: 1,
	}
}

func (r *InMemoryTaskRepository) Create(ctx context.Context, task *TaskModel) error {
	task.ID = r.nextID
	r.nextID++
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepository) Update(ctx context.Context, task *TaskModel) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepository) GetByID(ctx context.Context, id uint64) (*TaskModel, error) {
	if task, ok := r.tasks[id]; ok {
		return task, nil
	}
	return nil, ErrTaskNotFound
}

func (r *InMemoryTaskRepository) GetPendingTasks(ctx context.Context, types ...string) ([]*TaskModel, error) {
	var result []*TaskModel
	for _, task := range r.tasks {
		if task.Status == StatusQueued || task.Status == StatusProcessing || task.Status == StatusSuspending {
			if len(types) == 0 {
				result = append(result, task)
			} else {
				for _, t := range types {
					if task.Type == t {
						result = append(result, task)
						break
					}
				}
			}
		}
	}
	return result, nil
}

func (r *InMemoryTaskRepository) Delete(ctx context.Context, id uint64) error {
	delete(r.tasks, id)
	return nil
}
