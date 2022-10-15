// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017 Datadog, Inc.

// +build linux freebsd netbsd openbsd solaris dragonfly

package config

const (
	defaultConfdPath            = "/etc/dd-agent/conf.d"
	defaultAdditionalChecksPath = "/etc/dd-agent/checks.d"
	defaultLogPath              = "/var/log/datadog/agent.log"
	defaultJMXPipePath          = "/opt/datadog-agent/run"
	defaultSyslogURI            = "unixgram:///dev/log"
)
