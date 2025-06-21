package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	cli "github.com/urfave/cli/v2"

	gbapp "github.com/golem-base/seqctl/pkg/app"
	"github.com/golem-base/seqctl/pkg/config"
	"github.com/golem-base/seqctl/pkg/flags"
	"github.com/golem-base/seqctl/pkg/log"
	"github.com/golem-base/seqctl/pkg/provider"
	"github.com/golem-base/seqctl/pkg/repository"
	"github.com/golem-base/seqctl/pkg/ui/web"
	"github.com/golem-base/seqctl/pkg/version"

	_ "github.com/golem-base/seqctl/pkg/swagger"
)

func runWeb(c *cli.Context) error {
	// Load configuration
	cfg, err := config.LoadConfig(c)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logging for web mode
	if err := log.Init(
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.NoColor,
		cfg.Log.FilePath,
	); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	slog.Info("Starting web server", "address", c.String("address"), "port", c.Int("port"))

	// Debug: Check if context works
	go func() {
		<-c.Context.Done()
		slog.Info("Context cancelled in runWeb")
	}()

	// Create provider using factory
	appProvider, err := provider.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Create repository with caching
	repo := repository.NewCachedNetworkRepository(appProvider, 0, 0)

	// Initialize app with repository
	app := gbapp.New(cfg, repo)

	// Create web server
	serverConfig := web.DefaultServerConfig()
	serverConfig.Address = c.String("address")
	serverConfig.Port = c.Int("port")
	serverConfig.RefreshInterval = c.Int("refresh-interval")
	server := web.NewServer(serverConfig, app)

	// Run web server
	return server.Start(c.Context)
}

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
	cliapp.Version = version.VersionInfo()
	cliapp.Commands = []*cli.Command{
		{
			Name:   "web",
			Usage:  "Launch Web Interface",
			Flags:  append(flags.CommonFlags, flags.WebFlags...),
			Action: runWeb,
		},
	}

	// Run the application with the context
	if err := cliapp.RunContext(ctx, os.Args); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}
