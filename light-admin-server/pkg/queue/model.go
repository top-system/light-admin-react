package queue

import (
	"time"

	"github.com/top-system/light-admin/pkg/uuid"
)

// TaskModel represents the task model in database.
//
// It is a plain domain struct: pkg/queue stays framework-agnostic and persists it
// through the TaskRepository interface. Column mapping (including the sys_tasks
// table name and the public_ prefixed columns) lives in the injected sqlc-backed
// implementation in the application layer.
type TaskModel struct {
	ID            uint64          `json:"id"`
	Type          string          `json:"type"`
	Status        Status          `json:"status"`
	CorrelationID uuid.UUID       `json:"correlationId"`
	OwnerID       string          `json:"ownerId"`
	PrivateState  string          `json:"privateState"`
	PublicState   TaskPublicState `json:"publicState"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// TaskPublicState represents the public state of a task
type TaskPublicState struct {
	RetryCount       int           `json:"retryCount"`
	ExecutedDuration time.Duration `json:"executedDuration"`
	Error            string        `json:"error"`
	ErrorHistory     StringSlice   `json:"errorHistory"`
	ResumeTime       int64         `json:"resumeTime"`
}

// TaskOwner represents the owner of a task (simplified user interface)
type TaskOwner struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// StringSlice is a custom type for storing string slices in database
type StringSlice []string

// Scan implements the sql.Scanner interface
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return s.unmarshal(v)
	case string:
		return s.unmarshal([]byte(v))
	}

	*s = []string{}
	return nil
}

// Value implements the driver.Valuer interface
func (s StringSlice) Value() (interface{}, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return s.marshal()
}

func (s *StringSlice) unmarshal(data []byte) error {
	if len(data) == 0 || string(data) == "[]" {
		*s = []string{}
		return nil
	}

	// Simple JSON-like parsing
	str := string(data)
	if str[0] == '[' && str[len(str)-1] == ']' {
		str = str[1 : len(str)-1]
	}
	if str == "" {
		*s = []string{}
		return nil
	}

	// Split by comma and trim quotes
	var result []string
	current := ""
	inQuote := false
	for _, c := range str {
		if c == '"' {
			inQuote = !inQuote
		} else if c == ',' && !inQuote {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	*s = result
	return nil
}

func (s StringSlice) marshal() (string, error) {
	if len(s) == 0 {
		return "[]", nil
	}

	result := "["
	for i, str := range s {
		if i > 0 {
			result += ","
		}
		result += `"` + str + `"`
	}
	result += "]"
	return result, nil
}
