package dogstatsd

import (
	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/metrics"
)

// batcher batches multiple metrics before submission
// this struct is not safe for concurrent use
type batcher struct {
	samples      []metrics.MetricSample
	samplesCount int

	events        []*metrics.Event
	serviceChecks []*metrics.ServiceCheck

	shardedAgg *aggregator.ShardedAggregator

	// output channels
	choutSamples       chan<- []metrics.MetricSample
	choutEvents        chan<- []*metrics.Event
	choutServiceChecks chan<- []*metrics.ServiceCheck
}

func newBatcher(aggregator *aggregator.ShardedAggregator) *batcher {
	s, e, sc := aggregator.First().GetBufferedChannels()
	return &batcher{
		samples:            metrics.GlobalMetricSamplePool.GetBatch(),
		choutSamples:       s,
		choutEvents:        e,
		choutServiceChecks: sc,
		// XXX(remy):
		shardedAgg: aggregator,
	}
}

func (b *batcher) appendSample(sample metrics.MetricSample) {
	if b.samplesCount == len(b.samples) {
		b.flushSamples()
	}
	b.samples[b.samplesCount] = sample
	b.samplesCount++
}

func (b *batcher) appendEvent(event *metrics.Event) {
	b.events = append(b.events, event)
}

func (b *batcher) appendServiceCheck(serviceCheck *metrics.ServiceCheck) {
	b.serviceChecks = append(b.serviceChecks, serviceCheck)
}

func (b *batcher) flushSamples() {
	if b.samplesCount > 0 {
		//	// XXX(remy): very costly to do that with a BufferedAggregator
		//	b.choutSamples <- b.samples[:b.samplesCount]
		b.shardedAgg.PushSamples(b.samples[:b.samplesCount])

		b.samplesCount = 0
		b.samples = metrics.GlobalMetricSamplePool.GetBatch()
	}
}

func (b *batcher) flush() {
	b.flushSamples()
	if len(b.events) > 0 {
		b.choutEvents <- b.events
		b.events = []*metrics.Event{}
	}
	if len(b.serviceChecks) > 0 {
		b.choutServiceChecks <- b.serviceChecks
		b.serviceChecks = []*metrics.ServiceCheck{}
	}
}
