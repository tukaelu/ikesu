package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mackerelio/mackerel-client-go"
	"github.com/urfave/cli/v2"

	"github.com/tukaelu/ikesu/cmd/ikesu/subcommand"
	"github.com/tukaelu/ikesu/internal/config"
	"github.com/tukaelu/ikesu/internal/constants"
	"github.com/tukaelu/ikesu/internal/logger"
)

var version string
var revision string

func main() {
	cli := &cli.App{
		Name:  "ikesu",
		Usage: "Manage the health condition of the fish in the \"Ikesu\".",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "apikey",
				DefaultText: "**********",
				EnvVars:     []string{"MACKEREL_APIKEY", "IKESU_MACKEREL_APIKEY"},
				Required:    true,
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
			{
				Name:      "check",
				Usage:     "Detects disruptions in posted metrics and notifies the host as a CRITICAL alert.",
				UsageText: "ikesu check -config <config file> [-dry-run]",
				Action: func(ctx *cli.Context) error {

					// Show the provider name and metric name, then terminate.
					if ctx.Bool("show-providers") {
						showProvidersInspectionMetricMap()
						return nil
					}

					var l *logger.Logger
					var err error
					if l, err = logger.NewLogger(ctx.String("log"), ctx.String("log-level"), ctx.Bool("dry-run")); err != nil {
						return err
					}

					config, err := config.NewCheckConfig(ctx.Context, ctx.String("config"))
					if err != nil {
						return err
					}
					if err := config.Validate(); err != nil {
						return err
					}
					client, err := mackerel.NewClientWithOptions(
						ctx.String("apikey"),
						ctx.String("apibase"),
						false,
					)
					if err != nil {
						return err
					}
					check := &subcommand.Check{
						Config: config,
						Client: client,
						DryRun: ctx.Bool("dry-run"),
						Logger: l,
					}

					// wrap function
					handler := func(ctx context.Context) error {
						return check.Run(ctx)
					}
					l.Log.Info("Run command", "version", ctx.App.Version)
					l.Log.Debug("Config", "dump", fmt.Sprintf("%+v", config))

					if isLambda() {
						lambda.StartWithOptions(handler, lambda.WithContext(ctx.Context))
						return nil
					}
					return handler(ctx.Context)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Usage:   "Specify the path to the configuration file.",
						Aliases: []string{"c"},
						EnvVars: []string{"IKESU_CHECK_CONFIG"},
					},
					&cli.BoolFlag{
						Name:  "show-providers",
						Usage: "List the inspection metric names corresponding to the provider for each integration.",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Only a simplified display of the check results is performed, and no alerts are issued.",
					},
				},
			},
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

func isLambda() bool {
	return os.Getenv("AWS_EXECUTION_ENV") != "" || os.Getenv("AWS_LAMBDA_RUNTIME_API") != ""
}

func showProvidersInspectionMetricMap() {
	integrations := constants.GetIntegrations()
	providers := constants.GetProviders()
	immap := constants.GetProvidersInspectionMetricMap()
	for i := 0; i < len(integrations); i++ {
		fmt.Printf("Integration: %s\n", integrations[i])
		fmt.Println(strings.Repeat("-", 30))
		for _, key := range providers {
			if metric, ok := immap[integrations[i]][key]; ok {
				fmt.Printf("provider: %-20s, metric: %s\n", key, metric)
			}
		}
		fmt.Println("")
	}
}
