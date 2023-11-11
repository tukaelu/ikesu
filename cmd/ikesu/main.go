package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/tukaelu/ikesu/cmd/ikesu/subcommand"
)

var version string
var revision string

func main() {
	cli := &cli.App{
		Name:  "ikesu",
		Usage: "Manage the health condition of the fish in the \"Ikesu\".",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "apikey",
				EnvVars:  []string{"MACKEREL_APIKEY", "IKESU_MACKEREL_APIKEY"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "apibase",
				EnvVars: []string{"MACKEREL_APIBASE", "IKESU_MACKEREL_APIBASE"},
				Value:   "https://api.mackerelio.com/",
			},
			&cli.StringFlag{
				Name:  "log",
				Usage: "Specify the path to the log file. If not specified, the log will be output to stdout.",
			},
			&cli.StringFlag{
				Name:    "log-level",
				EnvVars: []string{"IKESU_LOG_LEVEL"},
				Value:   "info",
				Action: func(ctx *cli.Context, s string) error {
					valid := []string{"debug", "info", "warn", "error"}
					if !slices.Contains(valid, s) {
						return fmt.Errorf("unsupported log level '%s' has been set. It supports debug, info, warn, and error.", s)
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			subcommand.NewCheckCommand(),
		},
	}

	cli.Version = fmt.Sprintf("%s (rev.%s)", version, revision)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	if err := cli.RunContext(ctx, os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
