package checks

import (
	"github.com/DataDog/gopsutil/cpu"
	"github.com/DataDog/gopsutil/host"
	"github.com/DataDog/gopsutil/mem"

	"github.com/DataDog/datadog-agent/pkg/process/config"
	"github.com/DataDog/datadog-agent/pkg/process/model"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// CollectSystemInfo collects a set of system-level information that will not
// change until a restart. This bit of information should be passed along with
// the process messages.
func CollectSystemInfo(cfg *config.AgentConfig) (*model.SystemInfo, error) {
	hi, err := host.Info()
	if err != nil {
		return nil, err
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	mi, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	cpus := make([]*model.CPUInfo, 0, len(cpuInfo))
	log.Infof("CollectSystemInfo - CPU count: %d", len(cpuInfo))
	for _, c := range cpuInfo {
		cpus = append(cpus, &model.CPUInfo{
			Number:     c.CPU,
			Vendor:     c.VendorID,
			Family:     c.Family,
			Model:      c.Model,
			PhysicalId: c.PhysicalID,
			CoreId:     c.CoreID,
			Cores:      c.Cores,
			Mhz:        int64(c.Mhz),
			CacheSize:  c.CacheSize,
		})
		log.Infof("CollectSystemInfo - CPU cores: %d", c.Cores)
	}

	log.Infof("CollectSystemInfo - cpus count: %d", len(cpus))

	return &model.SystemInfo{
		Uuid: hi.HostID,
		Os: &model.OSInfo{
			Name:          hi.OS,
			Platform:      hi.Platform,
			Family:        hi.PlatformFamily,
			Version:       hi.PlatformVersion,
			KernelVersion: hi.KernelVersion,
		},
		Cpus:        cpus,
		TotalMemory: int64(mi.Total),
	}, nil
}
