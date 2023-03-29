package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func createSigChannel() (chan os.Signal, func()) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return signalCh, func() {
		close(signalCh)
	}
}

func startGracefulShutdown(ctx context.Context, internalServices *internalServicesContainer) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	internalServices.jetstreamSub.Close()

	if err := internalServices.apiService.Shutdown(ctx); err != nil {
		lo.Fatal("Could not gracefully shutdown api server", "err", err)
	}

	internalServices.taskerService.Stop()
}
