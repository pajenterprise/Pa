// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package dogstatsd

import (
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/metrics"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/datadog-agent/pkg/dogstatsd/listeners"
)

func mockAggregator() *aggregator.BufferedAggregator {
	agg := aggregator.NewBufferedAggregator(
		serializer.NewSerializer(nil),
		"hostname",
		time.Millisecond*10,
	)
	return agg
}

func buildPacketContent(numberOfMetrics int) []byte {
	rawPacket := "daemon:666|h|@0.5|#sometag1:somevalue1,sometag2:somevalue2"
	packets := rawPacket
	for i := 1; i < numberOfMetrics; i++ {
		packets += "\n" + rawPacket
	}
	return []byte(packets)
}

func BenchmarkParsePackets(b *testing.B) {
	// our logger will log dogstatsd packet by default if nothing is setup
	config.SetupLogger("", "off", "", "", false, true, false)
	rawPacket := buildPacketContent(20 * 32)

	agg := mockAggregator()
	s, _ := NewServer(agg)
	defer s.Stop()

	done := make(chan struct{})
	go func() {
		s, _, _ := agg.GetBufferedChannels()
		for {
			select {
			case <-s:
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	b.RunParallel(func(pb *testing.PB) {
		batcher := newBatcher(agg)
		parser := newParser()
		// 32 packets of 20 samples
		packet := listeners.Packet{
			Contents: rawPacket,
			Origin:   listeners.NoOrigin,
		}

		packets := listeners.Packets{&packet}
		for pb.Next() {
			packet.Contents = rawPacket
			s.parsePackets(batcher, parser, packets)
		}
	})
}

var m metrics.MetricSample

func BenchmarkParseMetricMessage(b *testing.B) {
	// our logger will log dogstatsd packet by default if nothing is setup
	config.SetupLogger("", "off", "", "", false, true, false)

	agg := mockAggregator()
	s, _ := NewServer(agg)
	defer s.Stop()

	done := make(chan struct{})
	go func() {
		s, _, _ := agg.GetBufferedChannels()
		for {
			select {
			case <-s:
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	parser := newParser()
	originTagger := originTags{}
	message := []byte("daemon:666|h|@0.5|#sometag1:somevalue1,sometag2:somevalue2")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m, _ = s.parseMetricMessage(parser, message, originTagger.getTags)
		}
	})
}

func BenchmarkWithMapper(b *testing.B) {
	datadogYaml := `
dogstatsd_mapper_profiles:
  - name: airflow
    prefix: 'airflow.'
    mappings:
      - match: "airflow.job.duration.*.*"       # metric format: airflow.job.duration.<job_type>.<job_name>
        name: "airflow.job.duration"            # remap the metric name
        tags:
          job_type: "$1"
          job_name: "$2"
      - match: "airflow.job.size.*.*"           # metric format: airflow.job.size.<job_type>.<job_name>
        name: "airflow.job.size"                # remap the metric name
        tags:
          foo: "$1"
          bar: "$2"
`
	config.Datadog.SetConfigType("yaml")
	err := config.Datadog.ReadConfig(strings.NewReader(datadogYaml))
	assert.NoError(b, err)

	BenchmarkMapperControl(b)
}
func BenchmarkMapperControl(b *testing.B) {
	port, err := getAvailableUDPPort()
	require.NoError(b, err)
	config.Datadog.SetDefault("dogstatsd_port", port)

	// our logger will log dogstatsd packet by default if nothing is setup
	config.SetupLogger("", "off", "", "", false, true, false)

	agg := mockAggregator()
	s, _ := NewServer(agg)
	defer s.Stop()

	done := make(chan struct{})
	go func() {
		s, _, _ := agg.GetBufferedChannels()
		for {
			select {
			case <-s:
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	batcher := newBatcher(agg)
	parser := newParser()

	for n := 0; n < b.N; n++ {
		packet := listeners.Packet{
			Contents: []byte("airflow.job.duration.my_job_type.my_job_name:666|g"),
			Origin:   listeners.NoOrigin,
		}
		packets := listeners.Packets{&packet}
		s.parsePackets(batcher, parser, packets)
	}

	b.ReportAllocs()
}
