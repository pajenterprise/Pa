package dogstatsd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceCheckMinimal(t *testing.T) {
	sc, err := parseServiceCheck([]byte("_sc|agent.up|0"))

	assert.Nil(t, err)
	assert.Equal(t, []byte("agent.up"), sc.name)
	assert.Equal(t, int64(0), sc.timestamp)
	assert.Equal(t, serviceCheckStatusOk, sc.status)
	assert.Equal(t, []byte(nil), sc.message)
	assert.Equal(t, 0, sc.tags.tagsCount)
}

func TestServiceCheckError(t *testing.T) {
	// not enough information
	_, err := parseServiceCheck([]byte("_sc|agent.up"))
	assert.Error(t, err)

	_, err = parseServiceCheck([]byte("_sc|agent.up|"))
	assert.Error(t, err)

	_, err = parseServiceCheck([]byte("_sc||"))
	assert.Error(t, err)

	_, err = parseServiceCheck([]byte("_sc|"))
	assert.Error(t, err)

	// not invalid status
	_, err = parseServiceCheck([]byte("_sc|agent.up|OK"))
	assert.Error(t, err)

	// not unknown status
	_, err = parseServiceCheck([]byte("_sc|agent.up|21"))
	assert.Error(t, err)

	// invalid timestamp
	_, err = parseServiceCheck([]byte("_sc|agent.up|0|d:some_time"))
	assert.NoError(t, err)

	// unknown metadata
	_, err = parseServiceCheck([]byte("_sc|agent.up|0|u:unknown"))
	assert.NoError(t, err)
}
func TestServiceCheckMetadataTimestamp(t *testing.T) {
	sc, err := parseServiceCheck([]byte("_sc|agent.up|0|d:21"))

	require.Nil(t, err)
	assert.Equal(t, []byte("agent.up"), sc.name)
	assert.Equal(t, int64(21), sc.timestamp)
	assert.Equal(t, serviceCheckStatusOk, sc.status)
	assert.Equal(t, []byte(nil), sc.message)
	assert.Equal(t, 0, sc.tags.tagsCount)
}

func TestServiceCheckMetadataHostname(t *testing.T) {
	sc, err := parseServiceCheck([]byte("_sc|agent.up|0|h:localhost"))

	require.Nil(t, err)
	assert.Equal(t, []byte("agent.up"), sc.name)
	assert.Equal(t, []byte("localhost"), sc.hostname)
	assert.Equal(t, int64(0), sc.timestamp)
	assert.Equal(t, serviceCheckStatusOk, sc.status)
	assert.Equal(t, []byte(nil), sc.message)
	assert.Equal(t, 0, sc.tags.tagsCount)
}

func TestServiceCheckMetadataTags(t *testing.T) {
	sc, err := parseServiceCheck([]byte("_sc|agent.up|0|#tag1,tag2:test,tag3"))

	require.Nil(t, err)
	assert.Equal(t, []byte("agent.up"), sc.name)
	assert.Equal(t, int64(0), sc.timestamp)
	assert.Equal(t, serviceCheckStatusOk, sc.status)
	assert.Equal(t, []byte(nil), sc.message)
	assert.Equal(t, 3, sc.tags.tagsCount)
	assert.Equal(t, []byte("tag1"), sc.tags.tags[0])
	assert.Equal(t, []byte("tag2:test"), sc.tags.tags[1])
	assert.Equal(t, []byte("tag3"), sc.tags.tags[2])
}

func TestServiceCheckMetadataMessage(t *testing.T) {
	sc, err := parseServiceCheck([]byte("_sc|agent.up|0|m:this is fine"))

	require.Nil(t, err)
	assert.Equal(t, []byte("agent.up"), sc.name)
	assert.Equal(t, int64(0), sc.timestamp)
	assert.Equal(t, serviceCheckStatusOk, sc.status)
	assert.Equal(t, []byte("this is fine"), sc.message)
	assert.Equal(t, 0, sc.tags.tagsCount)
}

func TestServiceCheckMetadataMultiple(t *testing.T) {
	// all type
	sc, err := parseServiceCheck([]byte("_sc|agent.up|0|d:21|h:localhost|#tag1:test,tag2|m:this is fine"))
	require.Nil(t, err)
	assert.Equal(t, []byte("agent.up"), sc.name)
	assert.Equal(t, []byte("localhost"), sc.hostname)
	assert.Equal(t, int64(21), sc.timestamp)
	assert.Equal(t, serviceCheckStatusOk, sc.status)
	assert.Equal(t, []byte("this is fine"), sc.message)
	assert.Equal(t, 2, sc.tags.tagsCount)
	assert.Equal(t, []byte("tag1:test"), sc.tags.tags[0])
	assert.Equal(t, []byte("tag2"), sc.tags.tags[1])

	// multiple time the same tag
	sc, err = parseServiceCheck([]byte("_sc|agent.up|0|d:21|h:localhost|h:localhost2|d:22"))
	require.Nil(t, err)
	assert.Equal(t, []byte("agent.up"), sc.name)
	assert.Equal(t, []byte("localhost2"), sc.hostname)
	assert.Equal(t, int64(22), sc.timestamp)
	assert.Equal(t, serviceCheckStatusOk, sc.status)
	assert.Equal(t, []byte(nil), sc.message)
	assert.Equal(t, 0, sc.tags.tagsCount)
}
