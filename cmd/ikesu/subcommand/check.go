package subcommand

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mackerelio/mackerel-client-go"
	"github.com/urfave/cli/v2"

	"github.com/tukaelu/ikesu/internal/config"
	"github.com/tukaelu/ikesu/internal/constants"
	"github.com/tukaelu/ikesu/internal/logger"
)

// NewCheckCommand returns a command that detects disruptions in posted metrics and notifies the host as a CRITICAL alert.
func NewCheckCommand() *cli.Command {
	return &cli.Command{
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
			check := &Check{
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
	}
}

type Check struct {
	Config *config.CheckConfig
	Client *mackerel.Client
	DryRun bool

	*logger.Logger
}

// FIXME: If 'service' is ever added to CheckSource.Type, it may be necessary to consider separating the logic.
// see. https://github.com/mackerelio/mackerel-client-go/blob/264b7b7a402a9638b8137ec0a8ab9b8e950eef5a/checks.go#L16
func (c *Check) Run(ctx context.Context) error {
	var reports []*mackerel.CheckReport

	checkedAt := time.Now().Unix()
	for _, rule := range c.Config.Rules {
		c.Log.Info("CheckRule", "name", rule.Name)
		p := &mackerel.FindHostsParam{
			Service: rule.Service,
		}
		if rule.Roles != nil {
			p.Roles = append(p.Roles, rule.Roles...)
		}

		hosts, err := c.Client.FindHosts(p)
		if err != nil {
			c.Log.Error("Failed to retrieve the hosts.", "reason", err.Error())
			return err
		}
		c.Log.Info("Retrieved target hosts.", "service", rule.Service, "roles", rule.Roles, "count", len(hosts))

		for _, host := range hosts {
			provider := getHostProviderType(host)
			c.Log.Info("Determine the provider of the host.", "host", host.ID, "provider", provider)

			// If the provider is explicitly stated in YAML, validation will only be performed on matching hosts.
			if len(rule.Providers) > 0 {
				if !slices.Contains[[]config.Provider, config.Provider](rule.Providers, config.Provider(provider)) {
					c.Log.Info("Skipping because it is not the target provider.", "host", host.ID, "provider", provider)
					continue
				}
			}

			metricNames := make([]string, 0)
			if suggested, ok := constants.GetProviderInspectionMetricName(provider); ok {
				metricNames = append(metricNames, suggested)
			}
			if specified, ok := rule.InspectionMetrics[provider]; ok {
				metricNames = append(metricNames, specified...)
			}

			if len(metricNames) == 0 {
				c.Log.Info("Skipping as there are no metrics to inspect.", "host", host.ID, "provider", provider)
				continue
			}

			status := mackerel.CheckStatusOK
			sum := 0
			message := ""
			for _, metricName := range metricNames {
				cnt := 0
				if cnt, err = c.retrieveMetricsCount(&ctx, host.ID, metricName, rule.InterruptedInterval.ToValue()); err != nil {
					c.Log.Error(fmt.Sprintf("Due to a failure in retrieving the metric '%s' for host '%s', it will be counted as 0 and the process will continue. ", metricName, host.ID), "reason", err.Error())
				}
				sum += cnt
			}
			if sum == 0 {
				status = mackerel.CheckStatusCritical
				message = fmt.Sprintf(
					"Metrics have been detected as disrupted for over %s on host '%s' with the provider '%s'. The inspected metric(s) is/are [%s]."+
						"To verify the exact situation, please check the posting status of the host's metrics.",
					rule.InterruptedInterval,
					host.ID,
					provider,
					strings.Join(metricNames, ", "),
				)
			} else {
				message = "No disruptions were detected in the metrics."
			}

			report := &mackerel.CheckReport{
				Source:     mackerel.NewCheckSourceHost(host.ID),
				Name:       fmt.Sprint("Ikesu Check(rule=", rule.Name, ")"),
				Status:     status,
				Message:    message,
				OccurredAt: checkedAt,
			}
			reports = append(reports, report)
		}
	}

	if c.DryRun {
		fmt.Println("--- The report will be displayed and then the process will end, because DryRun mode is specified.")
		for _, report := range reports {
			fmt.Printf("%+v\n", report)
		}
		return nil
	}

	// The Post Monitoring Check Reports API requires requests in batches of 100, so it is processed in segments.
	// see. https://mackerel.io/api-docs/entry/check-monitoring#post
	reportCount := len(reports)
	if reportCount == 0 {
		c.Log.Info("There were no results to report.")
	} else {
		c.Log.Info("Starting to report the check monitoring results.", "reports", reportCount)
	}
	for i := 0; i < reportCount; i += 100 {
		end := i + 100
		if reportCount < end {
			end = reportCount
		}
		if err := c.Client.PostCheckReports(&mackerel.CheckReports{Reports: reports[i:end]}); err != nil {
			c.Log.Error("Failed to post the check monitoring reports.", "progress", fmt.Sprintf("%d/%d", end, reportCount), "reason", err.Error())
			return err
		}
		c.Log.Debug("Posted the check monitoring reports.", "progress", fmt.Sprintf("%d/%d", end, reportCount))
	}
	return nil
}

// judge whether it is running on AWS Lambda.
func isLambda() bool {
	return os.Getenv("AWS_EXECUTION_ENV") != "" || os.Getenv("AWS_LAMBDA_RUNTIME_API") != ""
}

// Show the provider name and metric name, then terminate.
func showProvidersInspectionMetricMap() {
	integrations := constants.GetIntegrations()
	providers := constants.GetProviders()
	immap := constants.GetProvidersInspectionMetricMap()
	for i := 0; i < len(integrations); i++ {
		fmt.Printf("Integration: %s\n", integrations[i])
		fmt.Println(strings.Repeat("-", 35))
		for _, key := range providers {
			if metric, ok := immap[integrations[i]][key]; ok {
				fmt.Printf("provider: %-25s, metric: %s\n", key, metric)
			}
		}
		fmt.Println("")
	}
}

// see constants.providersInspectionMetricMap
func getHostProviderType(h *mackerel.Host) string {
	pType := make([]string, 0)
	skipProvider := false
	if h.Meta.AgentName != "" {
		if strings.Contains(h.Meta.AgentName, "mackerel-agent") {
			pType = append(pType, "agent")
		} else if strings.Contains(h.Meta.AgentName, "mackerel-container-agent") {
			pType = append(pType, "container-agent")
			skipProvider = true
		} else {
			// Maybe this host created by the API.
			skipProvider = true
		}
	}
	if !skipProvider && h.Meta.Cloud != nil && h.Meta.Cloud.Provider != "" {
		pType = append(pType, h.Meta.Cloud.Provider)
	}
	return strings.Join(pType, "-")
}

func (c *Check) retrieveMetricsCount(ctx *context.Context, hostId, metricName string, interval int32) (int, error) {
	var values []mackerel.MetricValue
	now := time.Now().Unix()
	from := now - int64(interval)
	to := int64(0)
	attempts := interval/constants.METRIC_INTERVAL_1MIN + 1
	for i := int64(0); i < int64(attempts); i++ {
		to = from + constants.METRIC_INTERVAL_1MIN
		if to > now {
			to = now
		}
		mv, err := c.Client.FetchHostMetricValues(hostId, metricName, from, to)
		// TODO: Isn't there a better way than checking with comparisons or Contains?
		if err != nil && strings.Contains(err.Error(), "metric not found") {
			// If 'metric not found' error is returned from the API, it will be skipped.
			c.Log.Info("FetchHostMetricValues returns metric not found", "hostId", hostId, "metricName", metricName, "from", from, "to", to)
		} else if err != nil {
			c.Log.Error("FetchHostMetricValues returns error", "hostId", hostId, "metricName", metricName, "from", from, "to", to, "reason", err.Error())
			return 0, err
		} else {
			values = append(values, mv...)
		}
		from = to
		time.Sleep(200)
	}
	return len(values), nil
}
