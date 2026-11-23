package core

import (
	"context"
	"sync"

	"github.com/alireza0/s-ui/util/common"

	sb "github.com/sagernet/sing-box"
	_ "github.com/sagernet/sing-box/experimental/clashapi"
	_ "github.com/sagernet/sing-box/experimental/v2rayapi"
	"github.com/sagernet/sing-box/option"
	_ "github.com/sagernet/sing-box/transport/v2rayquic"
)

// Core owns the running sing-box instance. Everything mutable lives behind mu,
// and the managers are read off the Box rather than cached in package vars, so
// a caller cannot mix managers from one box with an instance from another.
//
// RWMutex because the panel's status poll calls IsRunning on every request and
// must not queue behind a start.
//
// Lock ordering: service.lifecycleMu (outer) before Core.mu (inner).
type Core struct {
	mu         sync.RWMutex
	isRunning  bool
	instance   *Box
	userLimits map[string]int
	userRates  map[string]UserRate

	// Carries the protocol registries. Built once in NewCore and never
	// reassigned, so it needs no lock.
	ctx context.Context
}

func NewCore() *Core {
	ctx := sb.Context(context.Background(), InboundRegistry(), OutboundRegistry(),
		EndpointRegistry(), DNSTransportRegistry(), ServiceRegistry(), CertificateProviderRegistry())
	return &Core{ctx: ctx}
}

func (c *Core) GetCtx() context.Context {
	return c.ctx
}

func (c *Core) GetInstance() *Box {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.instance
}

func (c *Core) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isRunning
}

// running returns the live box, or an error when the core is stopped. One
// RLock, so a caller cannot see isRunning true and then find instance nil.
func (c *Core) running() (*Box, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.isRunning || c.instance == nil {
		return nil, common.NewError("sing-box is not running")
	}
	return c.instance, nil
}

// Start builds a box from the config and publishes it. The build and the start
// run outside the lock -- they bind listeners and take seconds -- and only the
// publish is locked, so a half-built box is never visible.
func (c *Core) Start(sbConfig []byte) error {
	var opt option.Options
	// Returned, not just logged: an empty option set starts a box with zero
	// inbounds that reports itself healthy, so the watchdog never retries.
	if err := opt.UnmarshalJSONContext(c.ctx, sbConfig); err != nil {
		return common.NewErrorf("unmarshal config: %v", err)
	}

	instance, err := NewBox(Options{
		Context: c.ctx,
		Options: opt,
	})
	if err != nil {
		return err
	}

	if err = instance.Start(); err != nil {
		_ = instance.Close()
		return err
	}

	c.mu.Lock()
	if tracker := instance.SessionTracker(); tracker != nil {
		if len(c.userLimits) > 0 {
			tracker.SetUserLimits(c.userLimits)
		}
		if len(c.userRates) > 0 {
			tracker.SetUserRates(c.userRates)
		}
	}
	c.instance = instance
	c.isRunning = true
	c.mu.Unlock()
	return nil
}

// SetUserRates updates per-user bandwidth caps both in Core state (so restarts
// keep them) and on the live SessionTracker immediately. A user missing from
// rates becomes unlimited.
func (c *Core) SetUserRates(rates map[string]UserRate) {
	copied := make(map[string]UserRate, len(rates))
	for user, caps := range rates {
		if user != "" && (caps.UpMbps > 0 || caps.DownMbps > 0) {
			copied[user] = caps
		}
	}
	c.mu.Lock()
	c.userRates = copied
	instance := c.instance
	running := c.isRunning
	c.mu.Unlock()
	if running && instance != nil {
		if tracker := instance.SessionTracker(); tracker != nil {
			tracker.SetUserRates(copied)
		}
	}
}

// SetUserLimits updates per-user concurrent connection limits both in Core
// state (so restarts keep them) and on the live SessionTracker immediately.
func (c *Core) SetUserLimits(limits map[string]int) {
	copied := make(map[string]int, len(limits))
	for user, limit := range limits {
		if user != "" && limit > 0 {
			copied[user] = limit
		}
	}
	c.mu.Lock()
	c.userLimits = copied
	instance := c.instance
	running := c.isRunning
	c.mu.Unlock()

	if running && instance != nil && instance.SessionTracker() != nil {
		instance.SessionTracker().SetUserLimits(copied)
	}
}

// Stop detaches the instance under the lock and closes it outside, since
// Box.Close shuts down ten manager subsystems in sequence.
func (c *Core) Stop() error {
	c.mu.Lock()
	instance := c.instance
	c.instance = nil
	c.isRunning = false
	c.mu.Unlock()

	if instance == nil {
		return nil
	}
	return instance.Close()
}
