package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ethereum-optimism/optimism/op-service/ctxinterrupt"
	cli "github.com/urfave/cli/v2"

	gbapp "github.com/golem-base/seqctl/pkg/app"
	"github.com/golem-base/seqctl/pkg/config"
	"github.com/golem-base/seqctl/pkg/flags"
	"github.com/golem-base/seqctl/pkg/log"
	"github.com/golem-base/seqctl/pkg/provider"
	"github.com/golem-base/seqctl/pkg/provider/k8s"
	tui "github.com/golem-base/seqctl/pkg/ui/tui"
	"github.com/golem-base/seqctl/pkg/version"
)

func main() {
	ctx, stopWaiting := ctxinterrupt.WithSignalWaiter(context.Background())
	defer stopWaiting()

	cliapp := cli.NewApp()
	cliapp.Name = "seqctl"
	cliapp.Usage = "Terminal UI for managing op-conductor sequencer clusters"
	cliapp.Version = version.VersionInfo()
	cliapp.Flags = flags.Flags
	cliapp.ArgsUsage = "<network-name>"

	cliapp.Action = func(c *cli.Context) error {
		// Require network name as argument
		if c.NArg() < 1 {
			return fmt.Errorf("network name is required")
		}
		networkName := c.Args().Get(0)

		// Load configuration
		cfg, err := config.LoadConfig(c)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize logging for TUI mode (redirect to file or disable)
		if err := log.InitForTUI(
			cfg.LogLevel,
			cfg.LogFormat,
			cfg.LogNoColor,
			cfg.LogFile,
		); err != nil {
			return fmt.Errorf("failed to initialize logging: %w", err)
		}

		slog.Debug("Loading configuration from Kubernetes", "selector", cfg.K8sSelector)

		// Initialize K8s client
		k8sClient, err := k8s.NewClient(cfg.K8sConfig)
		if err != nil {
			return fmt.Errorf("failed to create K8s client: %w", err)
		}

		// Create provider (use networkName as namespace, empty string means all namespaces)
		k8sProvider := provider.NewK8sProvider(k8sClient, networkName, cfg.K8sSelector)

		// Initialize app with provider
		app := gbapp.New(cfg, k8sProvider)

		// Get network
		network, err := app.GetNetwork(ctx, networkName)
		if err != nil {
			return fmt.Errorf("failed to get network %s: %w", networkName, err)
		}

		// Create and configure TUI
		tui := tui.NewTUI(network)

		// Apply TUI configuration from flags
		if c.IsSet("refresh-interval") {
			tui.SetRefreshInterval(c.Duration("refresh-interval"))
		}

		if c.IsSet("auto-refresh") {
			tui.SetAutoRefresh(c.Bool("auto-refresh"))
		}

		// Run TUI
		return tui.Run()
	}

	// Run the application
	if err := cliapp.Run(os.Args); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}
