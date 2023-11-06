package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/tukaelu/ikesu/internal/config/loader/file"
)

func TestConfigLoad(t *testing.T) {
	conf, _ := NewCheckConfig(context.TODO(), "testdata/check.yml")

	cases := &CheckConfig{
		Rules: []MetricCheckRule{
			{
				Name:                "hoge",
				Service:             "hoge_service",
				InterruptedInterval: "24h",
				Providers:           []Provider{"ec2", "rds"},
				InspectionMetrics: map[string][]string{
					"ec2": {
						"custom.foo.bar",
					},
				},
			},
			{
				Name:                "foo",
				Service:             "foo_service",
				Roles:               []string{"role1", "role2"},
				InterruptedInterval: "12h",
				Providers:           []Provider{"lambda"},
			},
		},
	}

	assert.EqualValues(t, cases, conf)
}

func TestConfigFileLoading(t *testing.T) {
	var cc *CheckConfig
	var err error
	_, err = NewCheckConfig(context.TODO(), "testdata")
	assert.EqualError(t, ErrNoSuchConfigFile, err.Error())
	_, err = NewCheckConfig(context.TODO(), "testdata/empty.yml")
	assert.EqualError(t, ErrEmptyConfigFile, err.Error())
	cc, err = NewCheckConfig(context.TODO(), "testdata/check_none.yml")
	assert.EqualError(t, ErrNoCheckRules, cc.Validate().Error())
	assert.Equal(t, nil, err)
}

func TestInterruptedInterval(t *testing.T) {
	cases := []struct {
		interval InterruptedInterval
		expected int32
		err      string
	}{
		{
			interval: "12h",
			expected: (60 * 60 * 12),
		},
		{
			interval: "24h",
			expected: (60 * 60 * 24),
		},
		{
			interval: "48h",
			expected: (60 * 60 * 48),
		},
		{ // It checks that it is valid within the period defined by constants.MAX_INTERRUPTED_INTERVAL.
			interval: "720h",
			expected: (60 * 60 * 24 * 30),
		},
		{ // This case would not normally occur as it should be prevented by validation, but the conversion can be done correctly.
			interval: "721h",
			expected: (60*60*24*30 + 60*60),
			err:      "interrupted_interval out of range: 2595600",
		},
	}
	for _, c := range cases {
		t.Run(string(c.interval), func(t *testing.T) {
			if c.err == "" {
				assert.Equal(t, nil, c.interval.validate())
			} else {
				assert.Equal(t, c.err, (c.interval.validate()).Error())
			}
			assert.Equal(t, c.expected, c.interval.ToValue())
		})
	}
}

func TestProviderValidation(t *testing.T) {
	cases := []struct {
		provider Provider
		expected error
	}{
		{
			provider: "ec2",
			expected: nil,
		},
		{
			provider: "rds",
			expected: nil,
		},
		{
			provider: "agent-ec2",
			expected: nil,
		},
		{ // Verify that a validation error occurs when an unsupported provider name is specified.
			provider: "unknown",
			expected: fmt.Errorf("unsupported provider, unknown has been set"),
		},
	}
	for _, c := range cases {
		t.Run(string(c.provider), func(t *testing.T) {
			assert.Equal(t, c.expected, c.provider.validate())
		})
	}
}
