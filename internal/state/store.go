package state

import (
	"sync"

	"release-bot/internal/models"
)

// Store is a concurrency-safe wrapper around the active CurrentRelease.
type Store struct {
	mu      sync.RWMutex
	release models.CurrentRelease
}

// IsActive returns true when a release thread has been initialised.
func (s *Store) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.release.ThreadID != ""
}

// Get returns a snapshot copy of the current release.
func (s *Store) Get() models.CurrentRelease {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.release
}

// Set replaces the entire release state (used on /release-init).
func (s *Store) Set(r models.CurrentRelease) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.release = r
}

// AppendItem adds a new ReleaseItem to the current release.
func (s *Store) AppendItem(item models.ReleaseItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.release.Items = append(s.release.Items, item)
}

// UpdateItem finds the item by developerID+branch, applies updates, and returns
// true if a match was found.
func (s *Store) UpdateItem(developerID, developerName, branch, newStatus, newPR, newBlocker string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.release.Items {
		item := &s.release.Items[i]
		if item.DeveloperID == developerID && item.Branch == branch {
			item.Status = newStatus
			item.DeveloperName = developerName
			if newPR != "" {
				item.PRLink = newPR
			}
			item.Blocker = newBlocker
			return true
		}
	}
	return false
}
