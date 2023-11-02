package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProviderInspectionMetricNames(t *testing.T) {
	ec2, _ := GetProviderInspectionMetricName("ec2")
	assert.Equal(t, "custom.ec2.status_check_failed.instance", ec2, "matching metric name corresponding to the provider.")
	rds, _ := GetProviderInspectionMetricName("rds")
	assert.Equal(t, "custom.rds.cpu.used", rds, "matching metric name corresponding to the provider.")
	unknown, ok := GetProviderInspectionMetricName("unknown-provider")
	assert.Equal(t, false, ok, "must be false because it is not a supported provider.")
	assert.Equal(t, "", unknown, "the metric name must be an empty string.")
}

func TestGetProviders(t *testing.T) {
	providers := GetProviders()
	assert.Equal(t, 46, len(providers), "The number of providers supported should be returned.")
}
