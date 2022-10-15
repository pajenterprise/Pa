// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package logs

import (
	"time"

	"github.com/DataDog/datadog-agent/pkg/logs/auditor"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/input/container"
	"github.com/DataDog/datadog-agent/pkg/logs/input/file"
	"github.com/DataDog/datadog-agent/pkg/logs/input/journald"
	"github.com/DataDog/datadog-agent/pkg/logs/input/listener"
	"github.com/DataDog/datadog-agent/pkg/logs/input/windowsevent"
	"github.com/DataDog/datadog-agent/pkg/logs/pipeline"
	"github.com/DataDog/datadog-agent/pkg/logs/restart"
	"github.com/DataDog/datadog-agent/pkg/logs/sender"
	"github.com/DataDog/datadog-agent/pkg/logs/service"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// Agent represents the data pipeline that collects, decodes,
// processes and sends logs to the backend
// + ------------------------------------------------------ +
// |                                                        |
// | Collector -> Decoder -> Processor -> Sender -> Auditor |
// |                                                        |
// + ------------------------------------------------------ +
type Agent struct {
	auditor          *auditor.Auditor
	destinationsCtx  *sender.DestinationsContext
	pipelineProvider pipeline.Provider
	inputs           []restart.Restartable
}

// NewAgent returns a new Agent
func NewAgent(sources *config.LogSources, services *service.Services, endpoints *config.Endpoints) *Agent {
	// setup the auditor
	auditor := auditor.New(config.LogsAgent.GetString("logs_config.run_path"))
	destinationsCtx := sender.NewDestinationsContext()

	// setup the pipeline provider that provides pairs of processor and sender
	pipelineProvider := pipeline.NewProvider(config.NumberOfPipelines, auditor, endpoints, destinationsCtx)

	// setup the inputs
	inputs := []restart.Restartable{
		file.NewScanner(sources, config.LogsAgent.GetInt("logs_config.open_files_limit"), pipelineProvider, auditor, file.DefaultSleepDuration),
		container.NewLauncher(sources, services, pipelineProvider, auditor),
		listener.NewLauncher(sources, config.LogsAgent.GetInt("logs_config.frame_size"), pipelineProvider),
		journald.NewLauncher(sources, pipelineProvider, auditor),
		windowsevent.NewLauncher(sources, pipelineProvider),
	}

	return &Agent{
		auditor:          auditor,
		destinationsCtx:  destinationsCtx,
		pipelineProvider: pipelineProvider,
		inputs:           inputs,
	}
}

// Start starts all the elements of the data pipeline
// in the right order to prevent data loss
func (a *Agent) Start() {
	starter := restart.NewStarter(a.destinationsCtx, a.auditor, a.pipelineProvider)
	for _, input := range a.inputs {
		starter.Add(input)
	}
	starter.Start()
}

// Stop stops all the elements of the data pipeline
// in the right order to prevent data loss
func (a *Agent) Stop() {
	inputs := restart.NewParallelStopper()
	for _, input := range a.inputs {
		inputs.Add(input)
	}
	stopper := restart.NewSerialStopper(
		inputs,
		a.pipelineProvider,
		a.auditor,
		a.destinationsCtx,
	)

	// This will try to stop everything in order, including the potentially blocking
	// parts like the sender. After StopTimeout it will just stop the last part of the
	// pipeline, disconnecting it from the auditor, to make sure that the pipeline is
	// flushed before stopping.
	// TODO: Add this feature in the stopper.
	c := make(chan struct{})
	go func() {
		stopper.Stop()
		close(c)
	}()
	timeout := time.Duration(config.LogsAgent.GetInt("logs_config.stop_grace_period")) * time.Second
	select {
	case <-c:
	case <-time.After(timeout):
		log.Info("Timed out when stopping logs-agent, forcing it to stop now")
		// We force all destinations to read/flush all the messages they get without
		// trying to write to the network.
		a.destinationsCtx.Stop()
		// Wait again for the stopper to complete.
		<-c
	}
}
