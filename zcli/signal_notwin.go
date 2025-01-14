//go:build !windows
// +build !windows

package zcli

import (
	"os"
	"os/signal"
	"syscall"
)

func KillSignal() bool {
	sig, stop := SignalChan()
	s := <-sig
	stop()
	return s != syscall.SIGUSR2
}

func SignalChan() (<-chan os.Signal, func()) {
	quit := make(chan os.Signal, 1)
	//nolint
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGUSR2)
	return quit, func() {
		signal.Stop(quit)
	}
}
