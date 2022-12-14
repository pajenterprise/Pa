// +build linux windows

package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/cmd/agent/api/pb"
	"github.com/DataDog/datadog-agent/cmd/agent/api/pb/mocks"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/process/procutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var originalConfig = config.Datadog

func restoreGlobalConfig() {
	config.Datadog = originalConfig
}

func TestBlacklist(t *testing.T) {
	testBlacklist := []string{
		"^getty",
		"^acpid",
		"^atd",
		"^upstart-udev-bridge",
		"^upstart-socket-bridge",
		"^upstart-file-bridge",
		"^dhclient",
		"^dhclient3",
		"^rpc",
		"^dbus-daemon",
		"udevd",
		"^/sbin/",
		"^/usr/sbin/",
		"^/var/ossec/bin/ossec",
		"^rsyslogd",
		"^whoopsie$",
		"^cron$",
		"^CRON$",
		"^/usr/lib/postfix/master$",
		"^qmgr",
		"^pickup",
		"^sleep",
		"^/lib/systemd/systemd-logind$",
		"^/usr/local/bin/goshe dnsmasq$",
	}
	blacklist := make([]*regexp.Regexp, 0, len(testBlacklist))
	for _, b := range testBlacklist {
		r, err := regexp.Compile(b)
		if err == nil {
			blacklist = append(blacklist, r)
		}
	}
	cases := []struct {
		cmdline     []string
		blacklisted bool
	}{
		{[]string{"getty", "-foo", "-bar"}, true},
		{[]string{"rpcbind", "-x"}, true},
		{[]string{"my-rpc-app", "-config foo.ini"}, false},
		{[]string{"rpc.statd", "-L"}, true},
		{[]string{"/usr/sbin/irqbalance"}, true},
	}

	for _, c := range cases {
		assert.Equal(t, c.blacklisted, IsBlacklisted(c.cmdline, blacklist),
			fmt.Sprintf("Case %v failed", c))
	}
}

func TestOnlyEnvConfig(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	// setting an API Key should be enough to generate valid config
	os.Setenv("DD_API_KEY", "apikey_from_env")
	defer os.Unsetenv("DD_API_KEY")
	os.Setenv("DD_PROCESS_AGENT_ENABLED", "true")
	defer os.Unsetenv("DD_PROCESS_AGENT_ENABLED")

	agentConfig, _ := NewAgentConfig("test", "", "")
	assert.Equal(t, "apikey_from_env", agentConfig.APIEndpoints[0].APIKey)
	assert.True(t, agentConfig.Enabled)

	os.Setenv("DD_PROCESS_AGENT_ENABLED", "false")
	agentConfig, _ = NewAgentConfig("test", "", "")
	assert.Equal(t, "apikey_from_env", agentConfig.APIEndpoints[0].APIKey)
	assert.False(t, agentConfig.Enabled)
}

func TestOnlyEnvConfigArgsScrubbingEnabled(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	os.Setenv("DD_CUSTOM_SENSITIVE_WORDS", "*password*,consul_token,*api_key")
	defer os.Unsetenv("DD_CUSTOM_SENSITIVE_WORDS")

	agentConfig, _ := NewAgentConfig("test", "", "")
	assert.Equal(t, true, agentConfig.Scrubber.Enabled)

	cases := []struct {
		cmdline       []string
		parsedCmdline []string
	}{
		{
			[]string{"spidly", "--mypasswords=123,456", "consul_token", "1234", "--dd_api_key=1234"},
			[]string{"spidly", "--mypasswords=********", "consul_token", "********", "--dd_api_key=********"},
		},
	}

	for i := range cases {
		cases[i].cmdline, _ = agentConfig.Scrubber.ScrubCommand(cases[i].cmdline)
		assert.Equal(t, cases[i].parsedCmdline, cases[i].cmdline)
	}
}

func TestOnlyEnvConfigArgsScrubbingDisabled(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	os.Setenv("DD_SCRUB_ARGS", "false")
	os.Setenv("DD_CUSTOM_SENSITIVE_WORDS", "*password*,consul_token,*api_key")
	defer os.Unsetenv("DD_SCRUB_ARGS")
	defer os.Unsetenv("DD_CUSTOM_SENSITIVE_WORDS")

	agentConfig, _ := NewAgentConfig("test", "", "")
	assert.Equal(t, false, agentConfig.Scrubber.Enabled)

	cases := []struct {
		cmdline       []string
		parsedCmdline []string
	}{
		{
			[]string{"spidly", "--mypasswords=123,456", "consul_token", "1234", "--dd_api_key=1234"},
			[]string{"spidly", "--mypasswords=123,456", "consul_token", "1234", "--dd_api_key=1234"},
		},
	}

	for i := range cases {
		fp := &procutil.Process{Cmdline: cases[i].cmdline}
		cases[i].cmdline = agentConfig.Scrubber.ScrubProcessCommand(fp)
		assert.Equal(t, cases[i].parsedCmdline, cases[i].cmdline)
	}
}

func TestOnlyEnvConfigLogLevelOverride(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	os.Setenv("DD_LOG_LEVEL", "error")
	defer os.Unsetenv("DD_LOG_LEVEL")
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")

	agentConfig, _ := NewAgentConfig("test", "", "")
	assert.Equal(t, "error", agentConfig.LogLevel)
}

func TestDisablingDNSInspection(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	t.Run("via YAML", func(t *testing.T) {
		cfg, err := NewAgentConfig(
			"test",
			"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-DisableDNS.yaml",
			"",
		)

		assert.Nil(t, err)
		assert.True(t, cfg.DisableDNSInspection)
	})

	t.Run("via ENV variable", func(t *testing.T) {
		os.Setenv("DD_DISABLE_DNS_INSPECTION", "true")
		defer os.Unsetenv("DD_DISABLE_DNS_INSPECTION")
		cfg, err := NewAgentConfig("test", "", "")

		assert.Nil(t, err)
		assert.True(t, cfg.DisableDNSInspection)
	})
}

func TestEnableHTTPMonitoring(t *testing.T) {
	t.Run("via YAML", func(t *testing.T) {
		config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
		defer restoreGlobalConfig()

		cfg, err := NewAgentConfig(
			"test",
			"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-EnableHTTP.yaml",
			"",
		)

		assert.Nil(t, err)
		assert.True(t, cfg.EnableHTTPMonitoring)
	})

	t.Run("via ENV variable", func(t *testing.T) {
		config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
		defer restoreGlobalConfig()

		os.Setenv("DD_SYSTEM_PROBE_NETWORK_ENABLE_HTTP_MONITORING", "true")
		defer os.Unsetenv("DD_SYSTEM_PROBE_NETWORK_ENABLE_HTTP_MONITORING")
		cfg, err := NewAgentConfig("test", "", "")

		assert.Nil(t, err)
		assert.True(t, cfg.EnableHTTPMonitoring)
	})
}

func TestEnableGatewayLookup(t *testing.T) {
	t.Run("via YAML", func(t *testing.T) {
		config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
		defer restoreGlobalConfig()

		// default config
		cfg, err := NewAgentConfig("test", "", "")
		assert.NoError(t, err)
		assert.False(t, cfg.EnableGatewayLookup)

		cfg, err = NewAgentConfig(
			"test",
			"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-EnableGwLookup.yaml",
			"",
		)

		assert.NoError(t, err)
		assert.True(t, cfg.EnableGatewayLookup)
	})

	t.Run("via ENV variable", func(t *testing.T) {
		config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
		defer restoreGlobalConfig()

		os.Setenv("DD_SYSTEM_PROBE_NETWORK_ENABLE_GATEWAY_LOOKUP", "true")
		defer os.Unsetenv("DD_SYSTEM_PROBE_NETWORK_ENABLE_GATEWAY_LOOKUP")
		cfg, err := NewAgentConfig("test", "", "")

		assert.NoError(t, err)
		assert.True(t, cfg.EnableGatewayLookup)
	})
}

func TestIgnoreConntrackInitFailure(t *testing.T) {
	t.Run("via YAML", func(t *testing.T) {
		config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
		defer restoreGlobalConfig()

		cfg, err := NewAgentConfig(
			"test",
			"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-IgnoreCTInitFailure.yaml",
			"",
		)

		assert.Nil(t, err)
		assert.True(t, cfg.IgnoreConntrackInitFailure)
	})

	t.Run("via ENV variable", func(t *testing.T) {
		config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
		defer restoreGlobalConfig()

		os.Setenv("DD_SYSTEM_PROBE_NETWORK_IGNORE_CONNTRACK_INIT_FAILURE", "true")
		defer os.Unsetenv("DD_SYSTEM_PROBE_NETWORK_IGNORE_CONNTRACK_INIT_FAILURE")
		cfg, err := NewAgentConfig("test", "", "")

		assert.Nil(t, err)
		assert.True(t, cfg.IgnoreConntrackInitFailure)
	})
}

func TestGetHostname(t *testing.T) {
	cfg := NewDefaultAgentConfig(false)
	h, err := getHostname(cfg.DDAgentBin, 0)
	assert.Nil(t, err)
	// verify we fall back to getting os hostname
	expectedHostname, _ := os.Hostname()
	assert.Equal(t, expectedHostname, h)
}

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	agentConfig := NewDefaultAgentConfig(false)

	// assert that some sane defaults are set
	assert.Equal("info", agentConfig.LogLevel)
	assert.Equal(true, agentConfig.AllowRealTime)
	assert.Equal(true, agentConfig.Scrubber.Enabled)

	os.Setenv("DOCKER_DD_AGENT", "yes")
	agentConfig = NewDefaultAgentConfig(false)
	assert.Equal(os.Getenv("HOST_PROC"), "")
	assert.Equal(os.Getenv("HOST_SYS"), "")
	os.Setenv("DOCKER_DD_AGENT", "no")
	assert.Equal(6062, agentConfig.ProcessExpVarPort)

	os.Unsetenv("DOCKER_DD_AGENT")
}

func TestAgentConfigYamlAndSystemProbeConfig(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	assert := assert.New(t)

	agentConfig, err := NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig.yaml",
		"",
	)
	assert.NoError(err)

	ep := agentConfig.APIEndpoints[0]
	assert.Equal("apikey_20", ep.APIKey)
	assert.Equal("my-process-app.datadoghq.com", ep.Endpoint.Hostname())
	assert.Equal(10, agentConfig.QueueSize)
	assert.Equal(true, agentConfig.AllowRealTime)
	assert.Equal(true, agentConfig.Enabled)
	assert.Equal(append(processChecks), agentConfig.EnabledChecks)
	assert.Equal(8*time.Second, agentConfig.CheckIntervals[ContainerCheckName])
	assert.Equal(30*time.Second, agentConfig.CheckIntervals[ProcessCheckName])
	assert.Equal(100, agentConfig.Windows.ArgsRefreshInterval)
	assert.Equal(false, agentConfig.Windows.AddNewArgs)
	assert.Equal(false, agentConfig.Scrubber.Enabled)
	assert.Equal(5065, agentConfig.ProcessExpVarPort)
	assert.False(agentConfig.DisableDNSInspection)

	agentConfig, err = NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig.yaml",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-Net.yaml",
	)
	assert.NoError(err)

	assert.Equal("apikey_20", ep.APIKey)
	assert.Equal("my-process-app.datadoghq.com", ep.Endpoint.Hostname())
	assert.Equal("server-01", agentConfig.HostName)
	assert.Equal(10, agentConfig.QueueSize)
	assert.Equal(true, agentConfig.AllowRealTime)
	assert.Equal(true, agentConfig.Enabled)
	assert.Equal(8*time.Second, agentConfig.CheckIntervals[ContainerCheckName])
	assert.Equal(30*time.Second, agentConfig.CheckIntervals[ProcessCheckName])
	assert.Equal(100, agentConfig.Windows.ArgsRefreshInterval)
	assert.Equal(false, agentConfig.Windows.AddNewArgs)
	assert.Equal(false, agentConfig.Scrubber.Enabled)
	if runtime.GOOS != "windows" {
		assert.Equal("/var/my-location/system-probe.log", agentConfig.SystemProbeAddress)
	}
	assert.Equal(append(processChecks, ConnectionsCheckName, NetworkCheckName), agentConfig.EnabledChecks)
	assert.Equal(500, agentConfig.ClosedChannelSize)
	assert.True(agentConfig.SysProbeBPFDebug)
	assert.Empty(agentConfig.ExcludedBPFLinuxVersions)
	assert.False(agentConfig.DisableTCPTracing)
	assert.False(agentConfig.DisableUDPTracing)
	assert.False(agentConfig.DisableIPv6Tracing)
	assert.False(agentConfig.DisableDNSInspection)

	agentConfig, err = NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig.yaml",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-Net-2.yaml",
	)
	assert.NoError(err)

	assert.Equal("apikey_20", ep.APIKey)
	assert.Equal("my-process-app.datadoghq.com", ep.Endpoint.Hostname())
	assert.Equal(10, agentConfig.QueueSize)
	assert.Equal(true, agentConfig.AllowRealTime)
	assert.Equal(true, agentConfig.Enabled)
	assert.Equal(8*time.Second, agentConfig.CheckIntervals[ContainerCheckName])
	assert.Equal(30*time.Second, agentConfig.CheckIntervals[ProcessCheckName])
	assert.Equal(100, agentConfig.Windows.ArgsRefreshInterval)
	assert.Equal(false, agentConfig.Windows.AddNewArgs)
	assert.Equal(false, agentConfig.Scrubber.Enabled)
	assert.False(agentConfig.SysProbeBPFDebug)
	assert.Equal(1000, agentConfig.ClosedChannelSize)
	assert.Equal(agentConfig.ExcludedBPFLinuxVersions, []string{"5.5.0", "4.2.1"})
	if runtime.GOOS != "windows" {
		assert.Equal("/var/my-location/system-probe.log", agentConfig.SystemProbeAddress)
	}
	assert.Equal(append(processChecks), agentConfig.EnabledChecks)
	assert.True(agentConfig.DisableTCPTracing)
	assert.True(agentConfig.DisableUDPTracing)
	assert.True(agentConfig.DisableIPv6Tracing)
	assert.False(agentConfig.DisableDNSInspection)
	assert.Equal(map[string][]string{"172.0.0.1/20": {"*"}, "*": {"443"}, "127.0.0.1": {"5005"}}, agentConfig.ExcludedSourceConnections)
	assert.Equal(map[string][]string{"172.0.0.1/20": {"*"}, "*": {"*"}, "2001:db8::2:1": {"5005"}}, agentConfig.ExcludedDestinationConnections)

	agentConfig, err = NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig.yaml",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-Net-Windows.yaml",
	)
	assert.NoError(err)

	if runtime.GOOS == "windows" {
		assert.Equal("localhost:4444", agentConfig.SystemProbeAddress)
	}
}

func TestProxyEnv(t *testing.T) {
	assert := assert.New(t)
	for i, tc := range []struct {
		host     string
		port     int
		user     string
		pass     string
		expected string
	}{
		{
			"example.com",
			1234,
			"",
			"",
			"http://example.com:1234",
		},
		{
			"https://example.com",
			4567,
			"foo",
			"bar",
			"https://foo:bar@example.com:4567",
		},
		{
			"example.com",
			0,
			"foo",
			"",
			"http://foo@example.com:3128",
		},
	} {
		os.Setenv("PROXY_HOST", tc.host)
		if tc.port > 0 {
			os.Setenv("PROXY_PORT", strconv.Itoa(tc.port))
		} else {
			os.Setenv("PROXY_PORT", "")
		}
		os.Setenv("PROXY_USER", tc.user)
		os.Setenv("PROXY_PASSWORD", tc.pass)
		pf, err := proxyFromEnv(nil)
		assert.NoError(err, "proxy case %d had error", i)
		u, err := pf(&http.Request{})
		assert.NoError(err)
		assert.Equal(tc.expected, u.String())
	}

	os.Unsetenv("PROXY_HOST")
	os.Unsetenv("PROXY_PORT")
	os.Unsetenv("PROXY_USER")
	os.Unsetenv("PROXY_PASSWORD")
}

func TestEnvSiteConfig(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	assert := assert.New(t)

	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	agentConfig, err := NewAgentConfig("test", "./testdata/TestEnvSiteConfig.yaml", "")
	assert.NoError(err)
	assert.Equal("process.datadoghq.io", agentConfig.APIEndpoints[0].Endpoint.Hostname())

	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	agentConfig, err = NewAgentConfig("test", "./testdata/TestEnvSiteConfig-2.yaml", "")
	assert.NoError(err)
	assert.Equal("process.datadoghq.eu", agentConfig.APIEndpoints[0].Endpoint.Hostname())

	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	agentConfig, err = NewAgentConfig("test", "./testdata/TestEnvSiteConfig-3.yaml", "")
	assert.NoError(err)
	assert.Equal("burrito.com", agentConfig.APIEndpoints[0].Endpoint.Hostname())

	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	os.Setenv("DD_PROCESS_AGENT_URL", "https://test.com")
	agentConfig, err = NewAgentConfig("test", "./testdata/TestEnvSiteConfig-3.yaml", "")
	assert.NoError(err)
	assert.Equal("test.com", agentConfig.APIEndpoints[0].Endpoint.Hostname())

	os.Unsetenv("DD_PROCESS_AGENT_URL")
}

func TestEnvProcessAdditionalEndpoints(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	assert := assert.New(t)

	expected := make(map[string]string)
	expected["key1"] = "url1.com"
	expected["key2"] = "url2.com"
	expected["key3"] = "url2.com"
	expected["apikey_20"] = "my-process-app.datadoghq.com" // from config file

	os.Setenv("DD_PROCESS_ADDITIONAL_ENDPOINTS", `{"https://url1.com": ["key1"], "https://url2.com": ["key2", "key3"]}`)
	defer os.Unsetenv("DD_PROCESS_ADDITIONAL_ENDPOINTS")

	agentConfig, err := NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig.yaml",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-Net.yaml",
	)
	assert.NoError(err)

	for _, actual := range agentConfig.APIEndpoints {
		assert.Equal(expected[actual.APIKey], actual.Endpoint.Hostname(), actual)
	}
}

func TestEnvOrchestratorAdditionalEndpoints(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	assert := assert.New(t)

	expected := make(map[string]string)
	expected["key1"] = "url1.com"
	expected["key2"] = "url2.com"
	expected["key3"] = "url2.com"
	expected["apikey_20"] = "orchestrator.datadoghq.com" // from config file

	os.Setenv("DD_ORCHESTRATOR_ADDITIONAL_ENDPOINTS", `{"https://url1.com": ["key1"], "https://url2.com": ["key2", "key3"]}`)
	defer os.Unsetenv("DD_ORCHESTRATOR_ADDITIONAL_ENDPOINTS")

	agentConfig, err := NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig.yaml",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-Net.yaml",
	)
	assert.NoError(err)

	for _, actual := range agentConfig.Orchestrator.OrchestratorEndpoints {
		assert.Equal(expected[actual.APIKey], actual.Endpoint.Hostname(), actual)
	}
}

func TestEnvAdditionalEndpointsMalformed(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	assert := assert.New(t)

	expected := make(map[string]string)
	expected["apikey_20"] = "my-process-app.datadoghq.com" // from config file

	os.Setenv("DD_PROCESS_ADDITIONAL_ENDPOINTS", `"https://url1.com","key1"`)
	defer os.Unsetenv("DD_PROCESS_ADDITIONAL_ENDPOINTS")

	agentConfig, err := NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig.yaml",
		"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-Net.yaml",
	)
	assert.NoError(err)

	for _, actual := range agentConfig.APIEndpoints {
		assert.Equal(expected[actual.APIKey], actual.Endpoint.Hostname(), actual)
	}
}

func TestNetworkConfig(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {
		agentConfig, err := NewAgentConfig(
			"test",
			"./testdata/TestDDAgentConfigYamlOnly.yaml",
			"./testdata/TestDDAgentConfig-NetConfig.yaml",
		)
		require.NoError(t, err)

		assert.True(t, agentConfig.EnableSystemProbe)
		assert.True(t, agentConfig.Enabled)
		assert.ElementsMatch(t, []string{ConnectionsCheckName, NetworkCheckName, ProcessCheckName, RTProcessCheckName}, agentConfig.EnabledChecks)
	})

	t.Run("env", func(t *testing.T) {
		os.Setenv("DD_SYSTEM_PROBE_NETWORK_CONFIG_ENABLED", "true")
		defer os.Unsetenv("DD_SYSTEM_PROBE_NETWORK_CONFIG_ENABLED")

		agentConfig, err := NewAgentConfig("test", "", "")
		require.NoError(t, err)

		assert.True(t, agentConfig.EnableSystemProbe)
		assert.True(t, agentConfig.Enabled)
		assert.ElementsMatch(t, []string{ConnectionsCheckName, NetworkCheckName, ProcessCheckName, RTProcessCheckName}, agentConfig.EnabledChecks)
	})
}

func TestSystemProbeNoNetwork(t *testing.T) {
	agentConfig, err := NewAgentConfig(
		"test",
		"./testdata/TestDDAgentConfigYamlOnly.yaml",
		"./testdata/TestDDAgentConfig-OOMKillOnly.yaml",
	)
	require.NoError(t, err)

	assert.True(t, agentConfig.EnableSystemProbe)
	assert.True(t, agentConfig.Enabled)
	assert.ElementsMatch(t, []string{OOMKillCheckName, ProcessCheckName, RTProcessCheckName}, agentConfig.EnabledChecks)

}

func TestIsAffirmative(t *testing.T) {
	value, err := isAffirmative("yes")
	assert.Nil(t, err)
	assert.True(t, value)

	value, err = isAffirmative("True")
	assert.Nil(t, err)
	assert.True(t, value)

	value, err = isAffirmative("1")
	assert.Nil(t, err)
	assert.True(t, value)

	_, err = isAffirmative("")
	assert.NotNil(t, err)

	value, err = isAffirmative("ok")
	assert.Nil(t, err)
	assert.False(t, value)
}

func TestEnablingDNSStatsCollection(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	t.Run("via YAML", func(t *testing.T) {
		cfg, err := NewAgentConfig(
			"test",
			"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-EnableDNSStats.yaml",
			"",
		)

		assert.Nil(t, err)
		assert.True(t, cfg.CollectDNSStats)
	})

	t.Run("via ENV variable", func(t *testing.T) {
		defer os.Unsetenv("DD_COLLECT_DNS_STATS")

		os.Setenv("DD_COLLECT_DNS_STATS", "false")
		cfg, err := NewAgentConfig("test", "", "")
		assert.Nil(t, err)
		assert.False(t, cfg.CollectDNSStats)

		os.Setenv("DD_COLLECT_DNS_STATS", "true")
		cfg, err = NewAgentConfig("test", "", "")
		assert.Nil(t, err)
		assert.True(t, cfg.CollectDNSStats)
	})
}

func TestEnablingDNSDomainCollection(t *testing.T) {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	defer restoreGlobalConfig()

	t.Run("via YAML", func(t *testing.T) {
		cfg, err := NewAgentConfig(
			"test",
			"./testdata/TestDDAgentConfigYamlAndSystemProbeConfig-EnableDNSDomains.yaml",
			"",
		)

		assert.Nil(t, err)
		assert.True(t, cfg.CollectDNSDomains)
	})

	t.Run("via ENV variable", func(t *testing.T) {
		defer os.Unsetenv("DD_COLLECT_DNS_DOMAINS")

		os.Setenv("DD_COLLECT_DNS_DOMAINS", "false")
		cfg, err := NewAgentConfig("test", "", "")
		assert.Nil(t, err)
		assert.False(t, cfg.CollectDNSDomains) // default value should be false

		os.Setenv("DD_COLLECT_DNS_DOMAINS", "true")
		cfg, err = NewAgentConfig("test", "", "")
		assert.Nil(t, err)
		assert.True(t, cfg.CollectDNSDomains)
	})
}

func TestGetHostnameFromGRPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAgentClient(ctrl)

	mockClient.EXPECT().GetHostname(
		gomock.Any(),
		&pb.HostnameRequest{},
	).Return(&pb.HostnameReply{Hostname: "unit-test-hostname"}, nil)

	t.Run("hostname returns from grpc", func(t *testing.T) {
		hostname, err := getHostnameFromGRPC(func(ctx context.Context, opts ...grpc.DialOption) (pb.AgentClient, error) {
			return mockClient, nil
		}, defaultGRPCConnectionTimeout)

		assert.Nil(t, err)
		assert.Equal(t, "unit-test-hostname", hostname)
	})

	t.Run("grpc client is unavailable", func(t *testing.T) {
		grpcErr := errors.New("no grpc client")
		hostname, err := getHostnameFromGRPC(func(ctx context.Context, opts ...grpc.DialOption) (pb.AgentClient, error) {
			return nil, grpcErr
		}, defaultGRPCConnectionTimeout)

		assert.NotNil(t, err)
		assert.Equal(t, grpcErr, errors.Unwrap(err))
		assert.Empty(t, hostname)
	})
}

func TestGetHostnameFromCmd(t *testing.T) {
	t.Run("valid hostname", func(t *testing.T) {
		h, err := getHostnameFromCmd("agent-success", fakeExecCommand)
		assert.Nil(t, err)
		assert.Equal(t, "unit_test_hostname", h)
	})

	t.Run("no hostname returned", func(t *testing.T) {
		h, err := getHostnameFromCmd("agent-empty_hostname", fakeExecCommand)
		assert.NotNil(t, err)
		assert.Equal(t, "", h)
	})
}

// TestGetHostnameShellCmd is a method that is called as a substitute for a dd-agent shell command,
// the GO_TEST_PROCESS flag ensures that if it is called as part of the test suite, it is skipped.
func TestGetHostnameShellCmd(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "agent-success":
		assert.EqualValues(t, []string{"hostname"}, args)
		fmt.Fprintf(os.Stdout, "unit_test_hostname")
	case "agent-empty_hostname":
		assert.EqualValues(t, []string{"hostname"}, args)
		fmt.Fprintf(os.Stdout, "")
	}
}

// fakeExecCommand is a function that initialises a new exec.Cmd, one which will
// simply call TestShellProcessSuccess rather than the command it is provided. It will
// also pass through the command and its arguments as an argument to TestShellProcessSuccess
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestGetHostnameShellCmd", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}
