package main

import (
	"context"
	"sync"

	"github.com/sagernet/sing-box/daemon"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type logEvent struct {
	Sequence  uint64 `json:"sequence"`
	Level     int32  `json:"level,omitempty"`
	LevelName string `json:"levelName,omitempty"`
	Message   string `json:"message,omitempty"`
	Reset     bool   `json:"reset,omitempty"`
}

type logBuffer struct {
	mu      sync.Mutex
	nextSeq uint64
	entries []logEvent
}

func newLogBuffer() *logBuffer {
	return &logBuffer{}
}

func (b *logBuffer) appendReset() {
	b.mu.Lock()
	b.nextSeq++
	b.entries = append(b.entries, logEvent{
		Sequence: b.nextSeq,
		Reset:    true,
	})
	b.mu.Unlock()
}

func (b *logBuffer) appendMessage(level int32, message string) {
	b.mu.Lock()
	b.nextSeq++
	b.entries = append(b.entries, logEvent{
		Sequence:  b.nextSeq,
		Level:     level,
		LevelName: logLevelName(level),
		Message:   message,
	})
	b.mu.Unlock()
}

func (b *logBuffer) drain(maxEntries int) []logEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	if maxEntries <= 0 || maxEntries >= len(b.entries) {
		entries := b.entries
		b.entries = nil
		return entries
	}
	entries := append([]logEvent(nil), b.entries[:maxEntries]...)
	b.entries = append([]logEvent(nil), b.entries[maxEntries:]...)
	return entries
}

func (b *logBuffer) clear() {
	b.mu.Lock()
	b.entries = nil
	b.mu.Unlock()
}

func logLevelName(level int32) string {
	switch level {
	case -1:
		return "disabled"
	case 0:
		return "panic"
	case 1:
		return "fatal"
	case 2:
		return "error"
	case 3:
		return "warn"
	case 4:
		return "info"
	case 5:
		return "debug"
	case 6:
		return "trace"
	default:
		return "unknown"
	}
}

type logStreamServer struct {
	grpc.ServerStream
	ctx    context.Context
	buffer *logBuffer
}

func (s *logStreamServer) Send(message *daemon.Log) error {
	if message.Reset_ {
		s.buffer.appendReset()
	}
	for _, entry := range message.Messages {
		s.buffer.appendMessage(int32(entry.Level), entry.Message)
	}
	return nil
}

func (s *logStreamServer) SetHeader(metadata.MD) error  { return nil }
func (s *logStreamServer) SendHeader(metadata.MD) error { return nil }
func (s *logStreamServer) SetTrailer(metadata.MD)       {}
func (s *logStreamServer) Context() context.Context     { return s.ctx }
func (s *logStreamServer) SendMsg(any) error            { return nil }
func (s *logStreamServer) RecvMsg(any) error            { return nil }
