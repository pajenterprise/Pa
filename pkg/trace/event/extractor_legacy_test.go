package event

import (
	"testing"

	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/trace/traceutil"
	"github.com/stretchr/testify/assert"
)

func TestLegacyCases(t *testing.T) {
	assert := assert.New(t)
	e := NewLegacyExtractor(map[string]float64{"serviCE1": 1})
	span := &pb.Span{Service: "SeRvIcE1"}
	traceutil.SetTopLevel(span, true)

	rate, ok := e.Extract(span, 0)
	assert.Equal(rate, 1.)
	assert.True(ok)
}
