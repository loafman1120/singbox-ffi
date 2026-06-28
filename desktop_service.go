package main

import (
	"context"
	"errors"
	"os"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/daemon"
	libbox "github.com/sagernet/sing-box/experimental/libbox"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing/service/filemanager"
)

type desktopCommandHandler struct{}

func (desktopCommandHandler) ServiceStop() error   { return nil }
func (desktopCommandHandler) ServiceReload() error { return nil }
func (desktopCommandHandler) GetSystemProxyStatus() (*libbox.SystemProxyStatus, error) {
	status, err := getSystemProxyStatus()
	if err != nil {
		return nil, err
	}
	return &libbox.SystemProxyStatus{Available: status.Supported, Enabled: status.Enabled}, nil
}
func (desktopCommandHandler) SetSystemProxyEnabled(enabled bool) error {
	return setSystemProxy(enabled, "127.0.0.1", 2080, "<local>")
}
func (desktopCommandHandler) TriggerNativeCrash() error { return errors.New("native crash disabled") }
func (desktopCommandHandler) WriteDebugMessage(string)  {}
func (desktopCommandHandler) ConnectSSHAgent() (int32, error) {
	return 0, errors.New("ssh agent is not implemented by singbox-ffi")
}

func newDesktopStartedService() *daemon.StartedService {
	options := getCurrentInitOptions()
	return daemon.NewStartedService(daemon.ServiceOptions{
		Context:           desktopServiceContext(options),
		Handler:           desktopCommandHandler{},
		Debug:             options.debug,
		LogMaxLines:       int(options.logMaxLines),
		OOMKillerEnabled:  options.oomKillerEnabled,
		OOMKillerDisabled: options.oomKillerDisabled,
		OOMMemoryLimit:    uint64(options.oomMemoryLimit),
	})
}

func desktopServiceContext(options initOptions) context.Context {
	ctx := context.Background()
	ctx = filemanager.WithDefault(ctx, options.workingPath, options.tempPath, os.Getuid(), os.Getgid())
	return box.Context(ctx, include.InboundRegistry(), include.OutboundRegistry(), include.EndpointRegistry(), include.DNSTransportRegistry(), include.ServiceRegistry(), include.CertificateProviderRegistry())
}
