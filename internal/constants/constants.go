package constants

import (
	"sort"
	"strings"
)

const (
	METRIC_INTERVAL_1MIN     = 60 * 60 * 20      // 20h (72,000sec)
	MAX_INTERRUPTED_INTERVAL = 60 * 60 * 24 * 30 // 30d (2,592,000sec)
)

// Determine the name of the metric that must be retained by the integration.
// If a combination exists for a provider, the metric that exists in that case is also enumerated. (e.g. agent-ec2)
// Metric names containing wildcards are not supported. And additionally, this definition does not ensure a guaranteed certainty.
// FIXME: The definition of this Map is incomplete.
var providersInspectionMetricMap = map[string]map[string]string{
	// AWS Integrations
	// https://mackerel.io/ja/docs/entry/integrations/aws
	"aws": {
		"ec2":           "custom.ec2.status_check_failed.instance", // There might not be a guarantee of reliable acquisition.
		"elb":           "custom.elb.count.request_count",
		"alb":           "custom.alb.request.count",
		"nlb":           "custom.nlb.tcp_reset.client_count",
		"rds":           "custom.rds.cpu.used", // There might not be a guarantee of reliable acquisition.
		"elasticache":   "",                    // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"redshift":      "",                    // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"lambda":        "",                    // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"sqs":           "",                    // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"dynamodb":      "custom.dynamodb.requests.success_requests",
		"cloudfront":    "custom.cloudfront.error_rate.total_error_rate",
		"apigateway":    "custom.apigateway.requests.count",
		"kinesis":       "", // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"s3":            "custom.s3.errors.4xx",
		"es":            "custom.es.automated_snapshot_failure.failure",
		"ecscluster":    "", // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"ses":           "custom.ses.email_sending_events.send",
		"states":        "custom.states.executions.succeeded", // Step Functions
		"efs":           "custom.efs.client_connections.count",
		"firehose":      "custom.firehose.throttled_records.records",
		"batch":         "", // It is not supported because there are only metric names that include wildcards.
		"waf":           "", // It is not supported because there are only metric names that include wildcards.
		"aws/billing":   "", // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"aws/route53":   "", // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"aws/connect":   "custom.connect.voice_calls.concurrent_calls",
		"aws/docdb":     "", // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
		"aws/codebuild": "custom.codebuild.builds.count",
	},

	// Azure Integrations
	// https://mackerel.io/ja/docs/entry/integrations/azure
	"azure": {
		"sqldatabase":              "custom.azure.sql_database.connection.successful",      // There might not be a guarantee of reliable acquisition.
		"rediscache":               "custom.azure.redis_cache.total_keys.count",            // There might not be a guarantee of reliable acquisition.
		"azurevm":                  "custom.azure.virtual_machine.cpu.percent",             // There might not be a guarantee of reliable acquisition.
		"appservice":               "custom.azure.app_service.requests.requests",           // There might not be a guarantee of reliable acquisition.
		"functions":                "custom.azure.functions.requests.requests",             // There might not be a guarantee of reliable acquisition.
		"azure/loadbalancer":       "",                                                     // It is not supported because there are only metric names that include wildcards.
		"azure/dbformysql":         "custom.azure.db_for_mysql.connections.active",         // There might not be a guarantee of reliable acquisition.
		"azure/dbforpostgresql":    "custom.azure.db_for_postgresql.connections.active",    // There might not be a guarantee of reliable acquisition.
		"azure/applicationgateway": "custom.azure.application_gateway.response_status.2xx", // There might not be a guarantee of reliable acquisition.
		"azure/blobstorage":        "",                                                     // It is not supported because there are only metric names that include wildcards.
		"azure/files":              "custom.azure.files.file_share_count.count",            // There might not be a guarantee of reliable acquisition.
	},

	// Google Cloud Integrations
	// https://mackerel.io/ja/docs/entry/integrations/gcp
	"gcp": {
		"gce":           "custom.gce.instance.cpu.used",              // There might not be a guarantee of reliable acquisition.
		"gcp/cloudsql":  "custom.cloudsql.network.connections.count", // There might not be a guarantee of reliable acquisition.
		"gcp/appengine": "",                                          // It is not supported because i couldn't find any metrics that are guaranteed to be reliably obtained.
	},

	// Combination of cloud integration and installed agents.
	"with-agent": {
		"agent-ec2":     "custom.ec2.status_check_failed.instance",  // There might not be a guarantee of reliable acquisition.
		"agent-azurevm": "custom.azure.virtual_machine.cpu.percent", // There might not be a guarantee of reliable acquisition
		"agent-gce":     "custom.gce.instance.cpu.used",             // There might not be a guarantee of reliable acquisition.
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
	lcp := strings.ToLower(provider)
	for _, integ := range providersInspectionMetricMap {
		if metric, ok := integ[lcp]; ok {
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
