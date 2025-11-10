package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/free5gc/nfpcf/internal/cache"
	"github.com/free5gc/nfpcf/internal/sbi"
	"github.com/free5gc/nfpcf/internal/sbi/consumer"
	"github.com/free5gc/nfpcf/internal/sbi/processor"
	"github.com/free5gc/nfpcf/pkg/factory"
)

type App struct {
	config    *factory.Config
	cache     *cache.NFProfileCache
	processor *processor.Processor
	server    *sbi.Server
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewApp(config *factory.Config) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	app.cache = cache.NewNFProfileCache(config.Cache.TTL)

	nrfClient := consumer.NewNRFClient(config.NRF.URL)

	app.processor = processor.NewProcessor(app.cache, nrfClient)

	app.server = sbi.NewServer(app.processor, config.Server.BindAddr)

	return app, nil
}

func (a *App) Start() error {
	fmt.Println("Starting NFPCF (NF Profile Cache Function)...")
	fmt.Printf("  Version: %s\n", a.config.Info.Version)
	fmt.Printf("  Backend NRF: %s\n", a.config.NRF.URL)
	fmt.Printf("  Cache TTL: %s\n", a.config.Cache.TTL)

	go a.handleSignals()

	if err := a.server.Run(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func (a *App) Stop() {
	fmt.Println("Stopping NFPCF...")

	if a.cache != nil {
		a.cache.Stop()
	}

	if a.server != nil {
		a.server.Shutdown()
	}

	a.cancel()
}

func (a *App) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		fmt.Println("\nReceived shutdown signal")
		a.Stop()
	case <-a.ctx.Done():
		return
	}
}
