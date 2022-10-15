// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package scheduler

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/service"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreConfigIfLogsExcluded(t *testing.T) {
	logSources := config.NewLogSources()
	services := service.NewServices()
	scheduler := NewScheduler(logSources, services)
	servicesStreamIn := services.GetAddedServicesForType(config.DockerType)
	servicesStreamOut := services.GetRemovedServicesForType(config.DockerType)

	configService := integration.Config{
		LogsConfig:   []byte(""),
		TaggerEntity: "container_id://a1887023ed72a2b0d083ef465e8edfe4932a25731d4bda2f39f288f70af3405b",
		Entity:       "docker://a1887023ed72a2b0d083ef465e8edfe4932a25731d4bda2f39f288f70af3405b",
		ClusterCheck: false,
		CreationTime: 0,
		LogsExcluded: true,
	}

	go scheduler.Schedule([]integration.Config{configService})
	select {
	case <-servicesStreamIn:
		assert.Fail(t, "config must be ignored")
	case <-time.After(100 * time.Millisecond):
		break
	}

	go scheduler.Unschedule([]integration.Config{configService})
	select {
	case <-servicesStreamOut:
		assert.Fail(t, "config must be ignored")
	case <-time.After(100 * time.Millisecond):
		break
	}
}
