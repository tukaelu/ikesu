package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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
		Usage: "We monitor the health of the fish in the \"Ikesu\".",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mackerel-apikey",
				DefaultText: "**********",
				EnvVars:     []string{"MACKEREL_APIKEY", "IKESU_MACKEREL_APIKEY"},
				Required:    true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "checker",
				Usage:     "Detects disruptions in posted metrics and notifies the host as a CRITICAL alert.",
				UsageText: "ikesu checker -config <config file> [-dry-run]",
				Action: func(ctx *cli.Context) error {

					// Show the provider name and metric name, then terminate.
					if ctx.Bool("show-providers") {
						showProvidersInspectionMetricMap()
						return nil
					}

					var l *logger.Logger
					var err error
					if l, err = logger.NewDefaultLogger("info", ctx.Bool("dry-run")); err != nil {
						return err
					}

					config, err := config.NewCheckerConfig(ctx.Context, ctx.String("config"))
					if err != nil {
						return err
					}
					if err := config.Validate(); err != nil {
						return err
					}
					checker := &subcommand.Checker{
						Config: config,
						Client: mackerel.NewClient(ctx.String("mackerel-apikey")),
						DryRun: ctx.Bool("dry-run"),
						Logger: l,
					}

					// wrap function
					handler := func(ctx context.Context) error {
						return checker.Run(ctx)
					}
					l.Log.Info("Run command", "version", ctx.App.Version)
					l.Log.V(1).Info(fmt.Sprintf("config: %+v", config))

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
						EnvVars: []string{"IKESU_CHECKER_CONFIG"},
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
