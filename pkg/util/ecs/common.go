// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017 Datadog, Inc.

// +build docker

package ecs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/docker"
	log "github.com/cihub/seelog"
)

const (
	metadataURL string = "http://169.254.170.2/v2/metadata"
	// StatsURL is the endpoint where task stats are exposed
	StatsURL string = "http://169.254.170.2/v2/stats"
	timeout         = 500 * time.Millisecond
)

// GetTaskMetadata extracts the metadata payload for the task the agent is in.
func GetTaskMetadata() (TaskMetadata, error) {
	var meta TaskMetadata
	log.Infof("Getting task metadata...") // TODO: delete me
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(metadataURL)
	if err != nil {
		return meta, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&meta)
	if err != nil {
		log.Errorf("decoding failed!") // TODO: delete me
	}
	return meta, err
}

// GetECSContainers returns all containers exposed by the ECS API as plain ECSContainers
func GetECSContainers() ([]Container, error) {
	meta, err := GetTaskMetadata()
	if err != nil || len(meta.Containers) == 0 {
		log.Errorf("unable to retrieve task metadata")
		return nil, err
	}
	return meta.Containers, nil
}

// GetContainers returns all containers exposed by the ECS API
// after transforming them into "generic" Docker containers.
// TODO: add a cache
func GetContainers() ([]*docker.Container, error) {
	var containers []*docker.Container
	var stats ContainerStats

	ecsContainers, err := GetECSContainers()
	if err != nil {
		log.Error("unable to get the container list from ecs")
		return containers, err
	}
	for _, c := range ecsContainers {
		entityID := fmt.Sprintf("docker://%s", c.DockerID)
		ctr := &docker.Container{
			Type:     "ECS",
			ID:       c.DockerID,
			EntityID: entityID,
			Name:     c.DockerName,
			Image:    c.Image,
			ImageID:  c.ImageID,
		}

		createdAt, err := time.Parse(time.RFC3339, c.CreatedAt)
		if err != nil {
			log.Errorf("unable to determine creation time for container %s - %s", c.DockerID, err)
		} else {
			ctr.Created = createdAt.Unix()
		}
		startedAt, err := time.Parse(time.RFC3339, c.StartedAt)
		if err != nil {
			log.Errorf("unable to determine creation time for container %s - %s", c.DockerID, err)
		} else {
			ctr.StartedAt = startedAt.Unix()
		}

		if l, found := c.Limits["cpu"]; found && l > 0 {
			ctr.CPULimit = float64(l)
		}
		if l, found := c.Limits["memory"]; found && l > 0 {
			ctr.MemLimit = l
		}
		stats, err = GetContainerStats(c)
		if err != nil {
			log.Errorf("Unable to get stats from ECS for container %s - %s", c.DockerID, err)
		} else {
			// TODO: add metrics - complete for https://github.com/DataDog/datadog-process-agent/blob/970729924e6b2b6fe3a912b62657c297621723cc/checks/container_rt.go#L110-L128
			// start with a hack (translate ecs stats to docker cgroup stuff)
			// then support ecs stats natively
			cpu, mem, io := convertECSStats(stats)
			log.Errorf("converted stats:")
			log.Errorf("CPU: %s", cpu.User)
			log.Errorf("MEM: %s", mem.RSS)
			log.Errorf("IO: %s", io.ReadBytes)
			ctr.CPU = &cpu
			ctr.Memory = &mem
			ctr.IO = &io
		}
		containers = append(containers, ctr)
	}
	return containers, err
}

// GetContainerStats retrives stats about a container from the ECS stats endpoint
func GetContainerStats(c Container) (ContainerStats, error) {
	var stats ContainerStats
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(StatsURL + "/" + c.DockerID)
	if err != nil {
		return stats, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&stats)
	if err != nil {
		log.Errorf("decoding failed!") // TODO: delete me
		return stats, err
	}
	stats.IO.ReadBytes = computeIOStats(stats.IO.BytesPerDeviceAndKind, "Read")
	stats.IO.WriteBytes = computeIOStats(stats.IO.BytesPerDeviceAndKind, "Write")
	log.Errorf("cpu user found: %d", stats.CPU.User)
	log.Errorf("memory rss found: %d", stats.Memory.RSS)
	log.Errorf("io read bytes found: %d", stats.IO.ReadBytes)
	return stats, nil
}

// computeIOStats sums all values across devices for an operation kind.
func computeIOStats(ops []OPStat, kind string) uint64 {
	var res uint64
	for _, op := range ops {
		if op.Kind == kind {
			res += op.Value
		}
	}
	return res
}

// convertECSStats is responsible for converting ecs stats structs to docker style stats
// TODO: get rid of this by supporting ECS stats everywhere we use docker stats only.
func convertECSStats(stats ContainerStats) (docker.CgroupTimesStat, docker.CgroupMemStat, docker.CgroupIOStat) {
	d := time.Duration(stats.CPU.Usage.Total) * time.Nanosecond
	cpu := docker.CgroupTimesStat{
		System: 0, // FIXME: Ignoring system for now
		User:   uint64(d / time.Second),
	}
	mem := docker.CgroupMemStat{
		RSS:             stats.Memory.RSS,
		Cache:           stats.Memory.RSS,
		Pgfault:         stats.Memory.RSS,
		MemUsageInBytes: stats.Memory.Usage,
	}
	io := docker.CgroupIOStat{
		ReadBytes:  stats.IO.ReadBytes,
		WriteBytes: stats.IO.WriteBytes,
	}
	return cpu, mem, io
}
