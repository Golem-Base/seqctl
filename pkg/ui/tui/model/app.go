package model

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/sequencer"
)

// AppModel represents the application state
type AppModel struct {
	network         *network.Network
	sequencers      []*sequencer.Sequencer
	selectedIndex   int
	lastUpdate      time.Time
	refreshInterval time.Duration
	autoRefresh     bool

	// Listeners
	listeners []AppListener

	// Thread safety
	mu sync.RWMutex
}

// AppListener defines the interface for listening to model changes
type AppListener interface {
	OnDataChanged(sequencers []*sequencer.Sequencer)
	OnSelectionChanged(seq *sequencer.Sequencer)
	OnError(error)
	OnRefreshCompleted(time.Time)
}

// NewAppModel creates a new application model
func NewAppModel(network *network.Network) *AppModel {
	return &AppModel{
		network:         network,
		selectedIndex:   -1,
		refreshInterval: 5 * time.Second,
		autoRefresh:     true,
		listeners:       make([]AppListener, 0),
	}
}

// AddListener adds a listener for model changes
func (m *AppModel) AddListener(listener AppListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicates
	if slices.Contains(m.listeners, listener) {
		return
	}

	m.listeners = append(m.listeners, listener)
}

// RemoveListener removes a listener
func (m *AppModel) RemoveListener(listener AppListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, l := range m.listeners {
		if l == listener {
			m.listeners = slices.Delete(m.listeners, i, i+1)
			break
		}
	}
}

// Refresh updates the network data
func (m *AppModel) Refresh(ctx context.Context) error {
	// Check for nil network
	if m.network == nil {
		err := fmt.Errorf("network is nil")
		m.notifyError(err)
		return err
	}

	if err := m.network.Update(ctx); err != nil {
		m.notifyError(err)
		return err
	}

	m.mu.Lock()
	m.sequencers = m.network.Sequencers()
	m.lastUpdate = time.Now()

	// Validate selected index
	if m.selectedIndex >= len(m.sequencers) {
		m.selectedIndex = len(m.sequencers) - 1
	}
	if m.selectedIndex < 0 && len(m.sequencers) > 0 {
		m.selectedIndex = 0
	}
	m.mu.Unlock()

	// Notify listeners of changes
	sequencers := make([]*sequencer.Sequencer, len(m.sequencers))
	copy(sequencers, m.sequencers)
	m.notifyListeners(func(l AppListener) {
		select {
		case <-ctx.Done():
			return
		default:
			l.OnDataChanged(sequencers)
			l.OnRefreshCompleted(m.lastUpdate)
		}
	})

	return nil
}

// GetSequencers returns the current sequencers
func (m *AppModel) GetSequencers() []*sequencer.Sequencer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sequencers
}

// GetSelectedSequencer returns the currently selected sequencer
func (m *AppModel) GetSelectedSequencer() *sequencer.Sequencer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.selectedIndex < 0 || m.selectedIndex >= len(m.sequencers) {
		return nil
	}
	return m.sequencers[m.selectedIndex]
}

// SetSelectedIndex sets the selected sequencer index
func (m *AppModel) SetSelectedIndex(index int) {
	var seq *sequencer.Sequencer

	m.mu.Lock()
	if index < 0 || index >= len(m.sequencers) {
		m.mu.Unlock()
		return
	}

	if m.selectedIndex != index {
		m.selectedIndex = index
		seq = m.sequencers[index]
	}
	m.mu.Unlock()

	// Notify outside the lock to prevent deadlock
	if seq != nil {
		m.notifyListeners(func(l AppListener) {
			l.OnSelectionChanged(seq)
		})
	}
}

// GetSelectedIndex returns the current selected index
func (m *AppModel) GetSelectedIndex() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.selectedIndex
}

// GetNetwork returns the network
func (m *AppModel) GetNetwork() *network.Network {
	return m.network
}

// GetLastUpdate returns the last update time
func (m *AppModel) GetLastUpdate() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastUpdate
}

// SetAutoRefresh enables or disables auto-refresh
func (m *AppModel) SetAutoRefresh(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.autoRefresh = enabled
}

// IsAutoRefresh returns the auto-refresh state
func (m *AppModel) IsAutoRefresh() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.autoRefresh
}

// SetRefreshInterval sets the refresh interval
func (m *AppModel) SetRefreshInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshInterval = interval
}

// GetRefreshInterval returns the refresh interval
func (m *AppModel) GetRefreshInterval() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.refreshInterval
}

// notifyListeners is a generic helper to notify all listeners
func (m *AppModel) notifyListeners(notify func(AppListener)) {
	m.mu.RLock()
	listeners := make([]AppListener, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, listener := range listeners {
		notify(listener)
	}
}

// notifyError notifies listeners of an error
func (m *AppModel) notifyError(err error) {
	m.notifyListeners(func(l AppListener) {
		l.OnError(err)
	})
}
