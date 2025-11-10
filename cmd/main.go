package main

import (
	"fmt"
	"os"

	"github.com/free5gc/nfpcf/pkg/app"
	"github.com/free5gc/nfpcf/pkg/factory"
	"github.com/urfave/cli/v2"
)

func main() {
	cliApp := &cli.App{
		Name:  "nfpcf",
		Usage: "NF Profile Cache Function",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
				Value:   "./config/nfpcfcfg.yaml",
			},
		},
		Action: action,
	}

	if err := cliApp.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func action(cliCtx *cli.Context) error {
	configPath := cliCtx.String("config")

	config, err := factory.ReadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	nfpcfApp, err := app.NewApp(config)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	if err := nfpcfApp.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return nil
}
