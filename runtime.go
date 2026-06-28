package main

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sagernet/sing-box/daemon"
	"google.golang.org/protobuf/types/known/emptypb"
)

type runtimeHandle struct {
	service   *daemon.StartedService
	logs      *logBuffer
	logCtx    context.Context
	logCancel context.CancelFunc
	stateMu   sync.RWMutex
	nowState  serviceState
	lastError string
}

type serviceState string

const (
	serviceStateRunning serviceState = "running"
	serviceStateStopped serviceState = "stopped"
	serviceStateClosed  serviceState = "closed"
)

type serviceSnapshot struct {
	State     serviceState `json:"state"`
	Running   bool         `json:"running"`
	Closed    bool         `json:"closed"`
	LastError string       `json:"lastError,omitempty"`
}

var (
	nextHandle atomic.Uint64
	handlesMu  sync.Mutex
	handles    = map[uint64]*runtimeHandle{}
)

func newRuntimeHandle(service *daemon.StartedService) *runtimeHandle {
	ctx, cancel := context.WithCancel(context.Background())
	return &runtimeHandle{
		service:   service,
		logs:      newLogBuffer(),
		logCtx:    ctx,
		logCancel: cancel,
		nowState:  serviceStateRunning,
	}
}

func getHandle(handle uint64) (*runtimeHandle, bool) {
	handlesMu.Lock()
	defer handlesMu.Unlock()
	runtime, ok := handles[handle]
	return runtime, ok
}

func (h *runtimeHandle) close() {
	if h.logCancel != nil {
		h.logCancel()
	}
	h.setState(serviceStateClosed)
	h.service.Close()
}

func (h *runtimeHandle) state() serviceState {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return h.nowState
}

func (h *runtimeHandle) setState(state serviceState) {
	h.stateMu.Lock()
	h.nowState = state
	h.stateMu.Unlock()
}

func (h *runtimeHandle) setLastError(message string) {
	h.stateMu.Lock()
	h.lastError = message
	h.stateMu.Unlock()
}

func (h *runtimeHandle) snapshot() serviceSnapshot {
	h.stateMu.RLock()
	defer h.stateMu.RUnlock()
	return serviceSnapshot{
		State:     h.nowState,
		Running:   h.nowState == serviceStateRunning,
		Closed:    h.nowState == serviceStateClosed,
		LastError: h.lastError,
	}
}

func (h *runtimeHandle) startLogSubscription() {
	go func() {
		_ = h.service.SubscribeLog(&emptypb.Empty{}, &logStreamServer{
			ctx:    h.logCtx,
			buffer: h.logs,
		})
	}()
}
