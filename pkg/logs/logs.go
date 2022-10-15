// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package logs

import (
	"encoding/json"
	"errors"
	"fmt"
	stdHttp "net/http"
	"net/url"
	"path"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-agent/pkg/logs/metrics"

	coreConfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util"
	httputils "github.com/DataDog/datadog-agent/pkg/util/http"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/DataDog/datadog-agent/pkg/version"

	"github.com/DataDog/datadog-agent/pkg/logs/client/http"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/scheduler"
	"github.com/DataDog/datadog-agent/pkg/logs/service"
	"github.com/DataDog/datadog-agent/pkg/logs/status"
)

const (
	// key used to display a warning message on the agent status
	invalidProcessingRules = "invalid_global_processing_rules"
	invalidEndpoints       = "invalid_endpoints"
)

// Transport is the transport used by logs-agent, i.e TCP or HTTP
type Transport string

const (
	// TransportHTTP indicates logs-agent is using HTTP transport
	TransportHTTP Transport = "HTTP"
	// TransportTCP indicates logs-agent is using TCP transport
	TransportTCP Transport = "TCP"
	// Metadata Polling interval in Seconds
	MetaPollInterval time.Duration = 2 * time.Second
	// Metadata API endpoint
	MetaEndpoint string = "api/v1/tags/hosts/"
)

var (
	// isRunning indicates whether logs-agent is running or not
	isRunning int32
	// logs-agent
	agent *Agent
	// scheduler is plugged to autodiscovery to collect integration configs
	// and schedule log collection for different kind of inputs
	adScheduler *scheduler.Scheduler
	// CurrentTransport is the current transport used by logs-agent, i.e TCP or HTTP
	CurrentTransport Transport
)

// Start starts logs-agent
func Start() error {
	if IsAgentRunning() {
		return nil
	}

	// setup the sources and the services
	sources := config.NewLogSources()
	services := service.NewServices()

	// setup the config scheduler
	adScheduler = scheduler.NewScheduler(sources, services)

	// setup the server config
	httpConnectivity := config.HTTPConnectivityFailure
	if endpoints, err := config.BuildHTTPEndpoints(); err == nil {
		httpConnectivity = http.CheckConnectivity(endpoints.Main)
	}
	endpoints, err := config.BuildEndpoints(httpConnectivity)
	if err != nil {
		message := fmt.Sprintf("Invalid endpoints: %v", err)
		status.AddGlobalError(invalidEndpoints, message)

		e := errors.New(message)
		log.Error("Could not start logs-agent: ", e)
	}
	CurrentTransport = TransportTCP
	if endpoints.UseHTTP {
		CurrentTransport = TransportHTTP
	}

	// setup the status
	status.Init(&isRunning, endpoints, sources, metrics.LogsExpvars)

	// setup global processing rules
	processingRules, err := config.GlobalProcessingRules()
	if err != nil {
		message := fmt.Sprintf("Invalid processing rules: %v", err)
		status.AddGlobalError(invalidProcessingRules, message)
		e := errors.New(message)
		log.Error("Could not start logs-agent: ", e)

		return e
	}

	// setup and start the agent
	agent = NewAgent(sources, services, processingRules, endpoints)
	log.Info("Starting logs-agent...")

	metadataTO := coreConfig.Datadog.GetInt("logs_config.logs_meta_timeout")
	if metadataTO > 0 {
		if coreConfig.Datadog.GetString("app_key") != "" {
			// poll for a certain amount of time
			err := metadataReady(endpoints, metadataTO)
			if err != nil {
				log.Infof("There was an issue waiting for the metadata: %v", err)
			}
		} else {
			log.Info("Application key required to wait for host tags on backend, skipping wait.")
		}
	}

	agent.Start()
	atomic.StoreInt32(&isRunning, 1)
	log.Info("logs-agent started")

	// add the default sources
	for _, source := range config.DefaultSources() {
		sources.AddSource(source)
	}

	return nil
}

// Stop stops properly the logs-agent to prevent data loss,
// it only returns when the whole pipeline is flushed.
func Stop() {
	log.Info("Stopping logs-agent")
	if IsAgentRunning() {
		if agent != nil {
			agent.Stop()
			agent = nil
		}
		if adScheduler != nil {
			adScheduler.Stop()
			adScheduler = nil
		}
		status.Clear()
		atomic.StoreInt32(&isRunning, 0)
	}
	log.Info("logs-agent stopped")
}

func pollMeta(endpoint string, tags []string) (bool, error) {
	hostname, err := util.GetHostname()
	if err != nil {
		return false, err
	}

	uri, err := url.Parse(endpoint)
	if err != nil {
		return false, err
	}
	uri.Scheme = "https"

	uri.Path = path.Join(uri.Path, hostname)
	transport := httputils.CreateHTTPTransport()

	// TODO: set a timeout on the client
	client := &stdHttp.Client{
		Transport: transport,
	}

	log.Debugf("Polling for metadata: %v", uri)
	req, err := stdHttp.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("datadog-agent/%s", version.AgentVersion))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", coreConfig.Datadog.GetString("api_key"))
	req.Header.Set("DD-APPLICATION-KEY", coreConfig.Datadog.GetString("app_key"))

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Server will respond 200 if the key is valid or 403 if invalid
	if resp.StatusCode == 200 {
		var jsonResponse map[string]interface{}

		json.NewDecoder(resp.Body).Decode(&jsonResponse)
		log.Debugf("metadata response received: %v", jsonResponse)
		_, found := jsonResponse["tags"]
		if !found {
			return false, nil
		}

		tagSet := make(map[string]struct{})
		for _, tag := range jsonResponse["tags"].([]interface{}) {
			tagSet[tag.(string)] = struct{}{}
		}

		ready := true
		for _, tag := range tags {
			_, ok := tagSet[tag]
			if !ok {
				ready = false
			}
		}

		return ready, nil

	} else if resp.StatusCode == 403 {
		return false, nil
	}

	return false, nil
}

//TODO: stop this if the agents shuts down
func metadataReady(endpoints *config.Endpoints, timeout int) error {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	ticker := time.NewTicker(MetaPollInterval)

	var api string
	re := regexp.MustCompile(`datadoghq.(com|eu){1}$`)
	if re.MatchString(endpoints.Main.Host) {
		api = path.Join(fmt.Sprintf("api.%s", re.FindString(endpoints.Main.Host)), MetaEndpoint)
	} else {
		message := fmt.Sprintf("unsupported target domain: %s", endpoints.Main.Host)
		return errors.New(message)
	}

	tags := coreConfig.Datadog.GetStringSlice("tags")

	for {
		select {
		case <-timer.C:
			log.Info("Timeout waiting for host metadata, some log entries may be missing host tags")
			return errors.New("unable to resolve metadata in time")
		case <-ticker.C:
			found, err := pollMeta(api, tags)
			if err != nil {
				log.Infof("There was an issue grabbing the host tags: %v", err)
			} else if found {
				return nil
			}
		}
	}
}

// IsAgentRunning returns true if the logs-agent is running.
func IsAgentRunning() bool {
	return status.Get().IsRunning
}

// GetStatus returns logs-agent status
func GetStatus() status.Status {
	return status.Get()
}

// GetScheduler returns the logs-config scheduler if set.
func GetScheduler() *scheduler.Scheduler {
	return adScheduler
}
