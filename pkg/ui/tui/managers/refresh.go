package managers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golem-base/seqctl/pkg/ui/tui/model"
	"github.com/rivo/tview"
)

// RefreshManager handles auto-refresh functionality
type RefreshManager struct {
	appModel   *model.AppModel
	flashModel *model.FlashModel
	app        *tview.Application

	// Protected state
	mu       sync.RWMutex
	enabled  bool
	interval time.Duration

	// Runtime state
	ticker *time.Ticker
	cancel context.CancelFunc
}

// NewRefreshManager creates a new refresh manager
func NewRefreshManager(appModel *model.AppModel, flashModel *model.FlashModel, app *tview.Application) *RefreshManager {
	return &RefreshManager{
		appModel:   appModel,
		flashModel: flashModel,
		app:        app,
		enabled:    true,
		interval:   5 * time.Second,
	}
}

// Start begins auto-refresh with the current settings
func (r *RefreshManager) Start() {
	r.Stop()

	r.mu.RLock()
	enabled := r.enabled
	interval := r.interval
	r.mu.RUnlock()

	if !enabled {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.ticker = time.NewTicker(interval)

	go func() {
		defer r.ticker.Stop()
		for {
			select {
			case <-r.ticker.C:
				r.performRefresh()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops auto-refresh
func (r *RefreshManager) Stop() {
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	if r.ticker != nil {
		r.ticker.Stop()
		r.ticker = nil
	}
}

// SetEnabled enables or disables auto-refresh
func (r *RefreshManager) SetEnabled(enabled bool) {
	r.mu.Lock()
	r.enabled = enabled
	r.mu.Unlock()

	if enabled {
		r.Start()
	} else {
		r.Stop()
	}
}

// SetInterval sets the refresh interval
func (r *RefreshManager) SetInterval(interval time.Duration) {
	r.mu.Lock()
	r.interval = interval
	r.mu.Unlock()

	r.mu.RLock()
	enabled := r.enabled
	r.mu.RUnlock()

	if enabled {
		r.Stop()
		r.Start()
	}
}

// IsEnabled returns whether auto-refresh is enabled
func (r *RefreshManager) IsEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.enabled
}

// GetInterval returns the current refresh interval
func (r *RefreshManager) GetInterval() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.interval
}

// RefreshNow triggers an immediate refresh
func (r *RefreshManager) RefreshNow() {
	r.performRefresh()
}

// InitialLoad performs the initial data load
func (r *RefreshManager) InitialLoad() {
	r.performRefresh()
}

// performRefresh executes a refresh operation
func (r *RefreshManager) performRefresh() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := r.appModel.Refresh(ctx)
		if err != nil {
			r.app.QueueUpdateDraw(func() {
				r.flashModel.Error(fmt.Sprintf("Refresh failed: %s", err.Error()))
			})
		}
	}()
}
