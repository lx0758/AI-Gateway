package router

import (
	"fmt"
	"sync"
	"time"
)

const (
	CooldownThreshold = 3
	CooldownDuration  = 30 * time.Minute
	RecordInterval    = 5 * time.Second
)

type CooldownState struct {
	Consecutive429 int
	CooldownUntil  *time.Time
	Last429Time    *time.Time
}

type CooldownManager struct {
	mu     sync.RWMutex
	states map[string]*CooldownState
}

func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		states: make(map[string]*CooldownState),
	}
}

func (m *CooldownManager) getKey(providerID uint, providerModelID uint) string {
	return fmt.Sprintf("%d:%d", providerID, providerModelID)
}

func (m *CooldownManager) getState(key string) *CooldownState {
	m.mu.Lock()
	defer m.mu.Unlock()
	if state, exists := m.states[key]; exists {
		return state
	}
	state := &CooldownState{
		Consecutive429: 0,
		CooldownUntil:  nil,
		Last429Time:    nil,
	}
	m.states[key] = state
	return state
}

func (m *CooldownManager) IsCooldown(providerID uint, providerModelID uint) bool {
	key := m.getKey(providerID, providerModelID)
	m.mu.RLock()
	defer m.mu.RUnlock()
	if state, exists := m.states[key]; exists {
		if state.CooldownUntil != nil && time.Now().Before(*state.CooldownUntil) {
			return true
		}
	}
	return false
}

func (m *CooldownManager) Record429(providerID uint, providerModelID uint) {
	key := m.getKey(providerID, providerModelID)
	state := m.getState(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if state.Last429Time != nil && now.Sub(*state.Last429Time) < RecordInterval {
		return
	}

	state.Last429Time = &now
	state.Consecutive429++
	if state.Consecutive429 >= CooldownThreshold {
		cooldownUntil := now.Add(CooldownDuration)
		state.CooldownUntil = &cooldownUntil
	}
}

func (m *CooldownManager) RecordSuccess(providerID uint, providerModelID uint) {
	key := m.getKey(providerID, providerModelID)
	state := m.getState(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	state.Consecutive429 = 0
	state.Last429Time = nil
}

func (m *CooldownManager) GetCooldownEndTime(providerID uint, providerModelID uint) *time.Time {
	key := m.getKey(providerID, providerModelID)
	m.mu.RLock()
	defer m.mu.RUnlock()
	if state, exists := m.states[key]; exists {
		return state.CooldownUntil
	}
	return nil
}

func (m *CooldownManager) GetEarliestCooldownEnd(providers []RouteResult) *RouteResult {
	var earliest *RouteResult
	var earliestTime *time.Time

	for _, result := range providers {
		endTime := m.GetCooldownEndTime(result.Provider.ID, result.ProviderModel.ID)
		if endTime != nil {
			if earliestTime == nil || endTime.Before(*earliestTime) {
				earliestTime = endTime
				earliest = &result
			}
		}
	}

	return earliest
}

func (m *CooldownManager) ClearExpiredCooldowns() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, state := range m.states {
		if state.CooldownUntil != nil && now.After(*state.CooldownUntil) {
			state.CooldownUntil = nil
			state.Consecutive429 = 0
		}
	}
}

func (m *CooldownManager) ClearCooldown(providerID uint, providerModelID uint) {
	key := m.getKey(providerID, providerModelID)
	m.mu.Lock()
	defer m.mu.Unlock()
	if state, exists := m.states[key]; exists {
		state.CooldownUntil = nil
		state.Consecutive429 = 0
		state.Last429Time = nil
	}
}

func (m *CooldownManager) ClearAllForProvider(providerID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefix := fmt.Sprintf("%d:", providerID)
	for key, state := range m.states {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			state.CooldownUntil = nil
			state.Consecutive429 = 0
			state.Last429Time = nil
		}
	}
}
