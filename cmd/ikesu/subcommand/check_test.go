package subcommand

import (
	"testing"

	"github.com/mackerelio/mackerel-client-go"
	"github.com/stretchr/testify/assert"
)

func TestGetHostProviderType(t *testing.T) {
	cases := []struct {
		host     *mackerel.Host
		expected string
	}{
		{
			host:     newHost("ec2", ""),
			expected: "ec2",
		},
		{
			host:     newHost("rds", ""),
			expected: "rds",
		},
		{
			host:     newHost("", "mackerel-agent"),
			expected: "agent",
		},
		{
			host:     newHost("ec2", "mackerel-agent"),
			expected: "agent-ec2",
		},
		{
			host:     newHost("", "mackerel-container-agent"),
			expected: "container-agent",
		},
		{
			host:     newHost("", "unknown"),
			expected: "",
		},
	}
	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			assert.Equal(t, c.expected, getHostProviderType(c.host))
		})
	}
}

func newHost(provider string, agent string) *mackerel.Host {
	return &mackerel.Host{
		Meta: mackerel.HostMeta{
			AgentName: agent,
			Cloud: &mackerel.Cloud{
				Provider: provider,
			},
		},
	}
}
