package storage

import (
	"sync"

	"valersbot/internal/errs"
	"valersbot/internal/models"
)

type MemoryStore struct {
	mu            sync.RWMutex
	plans         map[string]*models.Plan
	openByChannel map[string]string // channelID -> messageID
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		plans:         make(map[string]*models.Plan),
		openByChannel: make(map[string]string),
	}
}

func (s *MemoryStore) CreateOpenPlan(plan *models.Plan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.openByChannel[plan.ChannelID]; ok {
		return errs.ErrOpenPlanExists
	}

	s.plans[plan.ID] = plan
	s.openByChannel[plan.ChannelID] = plan.ID
	return nil
}

func (s *MemoryStore) GetPlan(messageID string) (*models.Plan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	plan, exists := s.plans[messageID]
	return plan, exists
}

func (s *MemoryStore) GetOpenPlanByChannel(channelID string) (*models.Plan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messageID, ok := s.openByChannel[channelID]
	if !ok {
		return nil, false
	}
	plan, exists := s.plans[messageID]
	return plan, exists
}

// UpdatePlan runs fn while holding the write lock so concurrent votes stay safe.
func (s *MemoryStore) UpdatePlan(messageID string, fn func(*models.Plan) error) (*models.Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan, exists := s.plans[messageID]
	if !exists {
		return nil, errs.ErrPlanNotFound
	}

	if err := fn(plan); err != nil {
		return nil, err
	}

	if plan.Closed {
		delete(s.openByChannel, plan.ChannelID)
	}

	return plan, nil
}
