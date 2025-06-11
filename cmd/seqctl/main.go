package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	cli "github.com/urfave/cli/v2"

	gbapp "github.com/golem-base/seqctl/pkg/app"
	"github.com/golem-base/seqctl/pkg/config"
	"github.com/golem-base/seqctl/pkg/flags"
	"github.com/golem-base/seqctl/pkg/log"
	"github.com/golem-base/seqctl/pkg/provider"
	"github.com/golem-base/seqctl/pkg/repository"
	"github.com/golem-base/seqctl/pkg/server"
	"github.com/golem-base/seqctl/pkg/version"

	_ "github.com/golem-base/seqctl/pkg/server/swagger"
)

func main() {
	// Initialize basic logging to stderr for startup
	if err := log.Init("info", "text", false, ""); err != nil {
		panic(fmt.Errorf("failed to initialize logging: %w", err))
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle signals
	go func() {
		sig := <-sigChan
		slog.Debug("Received signal", "signal", sig)
		cancel()
	}()

	cliapp := cli.NewApp()
	cliapp.Name = "seqctl"
	cliapp.Usage = "Control panel for managing op-conductor sequencer clusters"
	cliapp.Version = version.Info()
	cliapp.Commands = []*cli.Command{
		{
			Name:   "serve",
			Usage:  "Launch Server",
			Flags:  flags.ServeCommandFlags(),
			Action: runServe,
		},
	}

	// Run the application with the context
	if err := cliapp.RunContext(ctx, os.Args); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}

func runServe(c *cli.Context) error {
	// Load configuration
	cfg, err := config.LoadConfig(c)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logging
	if err := log.Init(
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.NoColor,
		cfg.Log.FilePath,
	); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	// Create provider using factory
	appProvider, err := provider.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Parse cache TTL durations
	discoveryTTL, err := time.ParseDuration(cfg.Cache.DiscoveryTTL)
	if err != nil {
		return fmt.Errorf("invalid cache discovery TTL '%s': %w", cfg.Cache.DiscoveryTTL, err)
	}

	statusTTL, err := time.ParseDuration(cfg.Cache.StatusTTL)
	if err != nil {
		return fmt.Errorf("invalid cache status TTL '%s': %w", cfg.Cache.StatusTTL, err)
	}

	// Create repository with caching
	repo := repository.NewCachedNetworkRepository(appProvider, discoveryTTL, statusTTL)

	// Initialize app with repository
	app := gbapp.New(cfg, repo)

	// Create server
	serverCfg := server.DefaultConfig()
	serverCfg.Address = cfg.Server.Address
	serverCfg.Port = cfg.Server.Port
	server := server.NewServer(serverCfg, app)

	// Run server
	return server.Start(c.Context)
}
