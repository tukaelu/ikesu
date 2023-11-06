package config

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/tukaelu/ikesu/internal/config/loader"
	"github.com/tukaelu/ikesu/internal/constants"
)

var (
	ErrNoCheckRules     = fmt.Errorf("No check rules defined.")
	ErrNoSuchConfigFile = fmt.Errorf("No such config file.")
	ErrEmptyConfigFile  = fmt.Errorf("The specified config file is empty.")
)

type CheckConfig struct {
	Rules []MetricCheckRule `yaml:"check"`
}

type MetricCheckRule struct {
	Name                string              `yaml:"name"`
	Service             string              `yaml:"service"`
	Roles               []string            `yaml:"roles"`
	InterruptedInterval InterruptedInterval `yaml:"interrupted_interval"`
	Providers           []Provider          `yaml:"providers"`
	InspectionMetrics   map[string][]string `yaml:"inspection_metrics"`
}

type InterruptedInterval string
type Provider string

// Validate returns the result of the validation.
func (c *CheckConfig) Validate() error {
	if c == nil || len(c.Rules) == 0 {
		return ErrNoCheckRules
	}

	var err error
	for _, rule := range c.Rules {
		if e := rule.validate(); e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func (r *MetricCheckRule) validate() error {
	var err error
	if r.Name == "" {
		err = errors.Join(err, fmt.Errorf("No name has been specified for the check."))
	}
	if r.Service == "" {
		err = errors.Join(err, fmt.Errorf("Service not specified for check '%s'.", r.Name))
	}
	err = errors.Join(err, r.InterruptedInterval.validate())
	for _, provider := range r.Providers {
		err = errors.Join(err, provider.validate())
	}
	return err
}

func (p InterruptedInterval) validate() error {
	d, err := time.ParseDuration(string(p))
	if err == nil {
		sec := d.Seconds()
		if sec < 0 || float64(constants.MAX_INTERRUPTED_INTERVAL) < sec {
			return fmt.Errorf("interrupted_interval out of range: %d", int32(sec))
		}
	}
	return nil
}

// ToValue returns an int32 type value of the time interval from the string.
func (p InterruptedInterval) ToValue() int32 {
	d, _ := time.ParseDuration(string(p))
	return int32(d.Seconds())
}

func (p Provider) validate() error {
	providers := constants.GetProviders()
	if !slices.Contains(providers, string(p)) {
		return fmt.Errorf("unsupported provider, %s has been set", p)
	}
	return nil
}

// NewCheckConfig returns the configuration content loaded from YAML.
func NewCheckConfig(ctx context.Context, confPath string) (*CheckConfig, error) {
	u, err := url.Parse(confPath)
	if err != nil {
		return nil, err
	}

	buf, err := loader.LoadWithContext(ctx, u)
	if err != nil {
		return nil, err
	}

	conf := &CheckConfig{}
	if err := yaml.Unmarshal(buf, conf); err != nil {
		return nil, err
	}
	for i := 0; i < len(conf.Rules); i++ {
		// If InterruptedInterval is unspecified, set it to a default value "24h".
		if conf.Rules[i].InterruptedInterval == "" {
			conf.Rules[i].InterruptedInterval = InterruptedInterval("24h")
		}
	}
	return conf, nil
}
