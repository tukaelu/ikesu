package constants

import "sort"

const (
	METRIC_INTERVAL_1MIN     = 60 * 60 * 20      // 20h (72,000sec)
	MAX_INTERRUPTED_INTERVAL = 60 * 60 * 24 * 30 // 30d (2,592,000sec)
)

// Determine the name of the metric that must be retained by the integration.
// If a combination exists for a provider, the metric that exists in that case is also enumerated. (e.g. agent-ec2)
// Metric names containing wildcards are not supported.
// FIXME: The definition of this Map is incomplete.
// TODO: If the metric fluctuates depending on the platform's subscription plan, it may not be covered.
var providersInspectionMetricMap = map[string]map[string]string{
	// AWS Integrations
	// https://mackerel.io/ja/docs/entry/integrations/aws
	"aws": {
		"ec2":           "custom.ec2.status_check_failed.instance",
		"elb":           "custom.elb.host_count.healthy",
		"alb":           "custom.alb.request.count",
		"nlb":           "custom.nlb.bytes.processed",
		"rds":           "custom.rds.cpu.used",
		"elasticache":   "custom.elasticache.cpu.used",
		"redshift":      "custom.redshift.cpu.used",
		"lambda":        "custom.lambda.count.invocations",
		"sqs":           "",
		"dynamodb":      "",
		"cloudfront":    "",
		"apigateway":    "",
		"kinesis":       "",
		"s3":            "",
		"es":            "",
		"ecs":           "",
		"ses":           "",
		"stepfunctions": "",
		"efs":           "",
		"firehose":      "",
		"batch":         "",
		"waf":           "",
		"billing":       "",
		"route53":       "",
		"connect":       "",
		"docdb":         "",
		"codebuild":     "",
	},

	// Azure Integrations
	// https://mackerel.io/ja/docs/entry/integrations/azure
	"azure": {
		"sql_database":        "",
		"redis_cache":         "",
		"virtual_machine":     "",
		"app_service":         "",
		"functions":           "",
		"load_balancer":       "",
		"db_for_mysql":        "",
		"db_for_postgresql":   "",
		"application_gateway": "",
		"blob_storage":        "",
		"files":               "",
	},

	// Google Cloud Integrations
	// https://mackerel.io/ja/docs/entry/integrations/gcp
	"gcp": {
		"computeengine": "",
		"cloudsql":      "",
		"appengine":     "",
	},

	// Combination of cloud integration and installed agents.
	"with-agent": {
		"agent-ec2":           "custom.ec2.status_check_failed.instance",
		"agent-vm":            "",
		"agent-computeengine": "",
		// No differences due to these patterns will occur. It is a consistent container-agent.
		// "container-agent-ecs":        "",
		// "container-agent-fargate":    "",
		// "container-agent-kubernetes": "",
	},

	// No inspection of the metric is performed, as it would normally be subject to connectivity monitoring or automatic retirement.
	"only-agent": {
		"agent":           "",
		"container-agent": "",
	},
}

// GetProviderInspectionMetricName returns the validation metric name corresponding to the provider name.
func GetProviderInspectionMetricName(provider string) (string, bool) {
	for _, integ := range providersInspectionMetricMap {
		if metric, ok := integ[provider]; ok {
			if ok && metric == "" {
				ok = false
			}
			return metric, ok
		}
	}
	return "", false
}

// GetProviders returns a sorted string slice of provider names.
// It is sorted in ascending order of provider names, not in the order defined above.
func GetProviders() []string {
	var providers []string
	for _, mmap := range providersInspectionMetricMap {
		for provider := range mmap {
			providers = append(providers, provider)
		}
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i] < providers[j]
	})
	return providers
}

// GetIntegrations returns integration kind names.
func GetIntegrations() []string {
	return []string{"aws", "azure", "gcp", "with-agent", "only-agent"}
}

// GetProvidersInspectionMetricMap returns a providersInspectionMetricMap.
func GetProvidersInspectionMetricMap() map[string]map[string]string {
	return providersInspectionMetricMap
}
