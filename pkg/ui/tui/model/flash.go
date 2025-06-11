package model

import (
	"slices"
	"sync"
	"time"
)

// FlashLevel represents the severity level of a flash message
type FlashLevel int

const (
	FlashInfo FlashLevel = iota
	FlashSuccess
	FlashWarning
	FlashError
)

// FlashMessage represents a temporary status message
type FlashMessage struct {
	Level     FlashLevel
	Message   string
	Timestamp time.Time
	Duration  time.Duration
}

// FlashModel manages flash messages
type FlashModel struct {
	messages  []FlashMessage
	listeners []FlashListener
	mu        sync.RWMutex
}

// FlashListener defines the interface for listening to flash message changes
type FlashListener interface {
	OnFlashMessage(msg FlashMessage)
	OnFlashCleared()
}

// NewFlashModel creates a new flash message model
func NewFlashModel() *FlashModel {
	return &FlashModel{
		messages:  make([]FlashMessage, 0),
		listeners: make([]FlashListener, 0),
	}
}

// AddListener adds a listener for flash messages
func (f *FlashModel) AddListener(listener FlashListener) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check for duplicates
	if slices.Contains(f.listeners, listener) {
		return
	}
	f.listeners = append(f.listeners, listener)
}

// RemoveListener removes a listener
func (f *FlashModel) RemoveListener(listener FlashListener) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i, l := range f.listeners {
		if l == listener {
			f.listeners = slices.Delete(f.listeners, i, i+1)
			break
		}
	}
}

// AddMessage adds a flash message with the specified level
func (f *FlashModel) AddMessage(level FlashLevel, message string) {
	var duration time.Duration
	switch level {
	case FlashInfo, FlashSuccess:
		duration = 3 * time.Second
	case FlashWarning, FlashError:
		duration = 5 * time.Second
	default:
		duration = 3 * time.Second
	}
	f.addMessage(level, message, duration)
}

// Convenience methods for backward compatibility
func (f *FlashModel) Info(message string)    { f.AddMessage(FlashInfo, message) }
func (f *FlashModel) Success(message string) { f.AddMessage(FlashSuccess, message) }
func (f *FlashModel) Warning(message string) { f.AddMessage(FlashWarning, message) }
func (f *FlashModel) Error(message string)   { f.AddMessage(FlashError, message) }

// addMessage adds a message with specified level and duration
func (f *FlashModel) addMessage(level FlashLevel, message string, duration time.Duration) {
	msg := FlashMessage{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Duration:  duration,
	}

	f.mu.Lock()
	f.messages = append(f.messages, msg)
	f.mu.Unlock()

	// Notify listeners
	f.notifyListeners(func(l FlashListener) {
		l.OnFlashMessage(msg)
	})

	// Auto-clear after duration
	go func() {
		time.Sleep(duration)
		f.clearOldMessages()
	}()
}

// GetCurrentMessage returns the most recent active message
func (f *FlashModel) GetCurrentMessage() *FlashMessage {
	f.mu.RLock()
	defer f.mu.RUnlock()

	now := time.Now()
	for i := len(f.messages) - 1; i >= 0; i-- {
		msg := &f.messages[i]
		if now.Sub(msg.Timestamp) < msg.Duration {
			return msg
		}
	}
	return nil
}

// clearOldMessages removes expired messages
func (f *FlashModel) clearOldMessages() {
	f.mu.Lock()

	now := time.Now()
	activeMessages := make([]FlashMessage, 0)

	for _, msg := range f.messages {
		if now.Sub(msg.Timestamp) < msg.Duration {
			activeMessages = append(activeMessages, msg)
		}
	}

	hadMessages := len(f.messages) > 0
	f.messages = activeMessages
	hasMessages := len(f.messages) > 0

	f.mu.Unlock()

	// Notify if all messages were cleared
	if hadMessages && !hasMessages {
		f.notifyListeners(func(l FlashListener) {
			l.OnFlashCleared()
		})
	}
}

// Clear removes all messages
func (f *FlashModel) Clear() {
	f.mu.Lock()
	f.messages = make([]FlashMessage, 0)
	f.mu.Unlock()

	f.notifyListeners(func(l FlashListener) {
		l.OnFlashCleared()
	})
}

// notifyListeners is a generic helper to notify all listeners
func (f *FlashModel) notifyListeners(notify func(FlashListener)) {
	f.mu.RLock()
	listeners := make([]FlashListener, len(f.listeners))
	copy(listeners, f.listeners)
	f.mu.RUnlock()

	// Always notify synchronously for UI operations
	for _, listener := range listeners {
		notify(listener)
	}
}
