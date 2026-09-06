package coordinate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var ErrRevisionConflict = errors.New("plan revision conflict")

type PlanStore struct {
	mu   sync.RWMutex
	path string
	plan Plan
}

func NewPlanStore(path string, plan Plan) (*PlanStore, error) {
	if err := ValidatePlan(plan); err != nil {
		return nil, err
	}
	return &PlanStore{path: path, plan: plan}, nil
}

func (store *PlanStore) Current() Plan {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return clonePlan(store.plan)
}

func (store *PlanStore) Update(next Plan) (Plan, error) {
	if err := ValidatePlan(next); err != nil {
		return Plan{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if next.Revision != store.plan.Revision {
		return Plan{}, ErrRevisionConflict
	}
	next.Revision++
	content, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return Plan{}, err
	}
	content = append(content, '\n')
	directory := filepath.Dir(store.path)
	temporary, err := os.CreateTemp(directory, ".spidey-plan-*")
	if err != nil {
		return Plan{}, fmt.Errorf("create temporary plan: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return Plan{}, err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return Plan{}, err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return Plan{}, err
	}
	if err := temporary.Close(); err != nil {
		return Plan{}, err
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return Plan{}, fmt.Errorf("replace plan: %w", err)
	}
	store.plan = clonePlan(next)
	return clonePlan(next), nil
}

func clonePlan(plan Plan) Plan {
	content, _ := json.Marshal(plan)
	var cloned Plan
	_ = json.Unmarshal(content, &cloned)
	return cloned
}
