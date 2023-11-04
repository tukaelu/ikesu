package subcommand

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mackerelio/mackerel-client-go"

	"github.com/tukaelu/ikesu/internal/config"
	"github.com/tukaelu/ikesu/internal/constants"
	"github.com/tukaelu/ikesu/internal/logger"
)

type Checker struct {
	Config *config.CheckerConfig
	Client *mackerel.Client
	DryRun bool

	*logger.Logger
}

// FIXME: If 'service' is ever added to CheckSource.Type, it may be necessary to consider separating the logic.
// see. https://github.com/mackerelio/mackerel-client-go/blob/264b7b7a402a9638b8137ec0a8ab9b8e950eef5a/checks.go#L16
func (c *Checker) Run(ctx context.Context) error {
	var reports []*mackerel.CheckReport

	checkedAt := time.Now().Unix()
	for _, rule := range c.Config.Rules {
		c.Log.Info("CheckerRule", "name", rule.Name)
		p := &mackerel.FindHostsParam{
			Service: rule.Service,
		}
		if rule.Roles != nil {
			p.Roles = append(p.Roles, rule.Roles...)
		}

		hosts, err := c.Client.FindHosts(p)
		if err != nil {
			c.Log.Error(err, "Failed to retrieve the hosts.")
			return err
		}
		c.Log.Info("Retrieved target hosts", "count", len(hosts))

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
					c.Log.Error(err, fmt.Sprintf("Due to a failure in retrieving the metric '%s' for host '%s', it will be counted as 0 and the process will continue. ", metricName, host.ID))
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
			}

			report := &mackerel.CheckReport{
				Source:     mackerel.NewCheckSourceHost(host.ID),
				Name:       fmt.Sprint("IkesuChecker(", rule.Name, ")"),
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
	c.Log.Info("Starting to report the check monitoring results.", "reports", reportCount)
	for i := 0; i < reportCount; i += 100 {
		end := i + 100
		if reportCount < end {
			end = reportCount
		}
		if err := c.Client.PostCheckReports(&mackerel.CheckReports{Reports: reports[i:end]}); err != nil {
			c.Log.Error(err, "Failed to post the check monitoring reports.", "progress", fmt.Sprintf("%d/%d", end, reportCount))
			return err
		}
	}
	return nil
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

func (c *Checker) retrieveMetricsCount(ctx *context.Context, hostId, metricName string, interval int32) (int, error) {
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
			c.Log.Error(err, "FetchHostMetricValues returns error", "hostId", hostId, "metricName", metricName, "from", from, "to", to)
			return 0, err
		} else {
			values = append(values, mv...)
		}
		from = to
		time.Sleep(200)
	}
	return len(values), nil
}
