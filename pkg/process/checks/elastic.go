package checks

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"

	model "github.com/DataDog/agent-payload/process"
	"github.com/DataDog/datadog-agent/pkg/process/config"
	"github.com/DataDog/datadog-agent/pkg/util/log"

	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
	"github.com/tidwall/gjson"
)

// Elasticsearch is a singleton ElasticCheck.
var Elasticsearch = &ElasticCheck{}

// ElasticCheck collects shard + cluster information
type ElasticCheck struct {
	sync.Mutex

	sysInfo *model.SystemInfo
	lastRun time.Time

	nodeName    string
	clusterName string
	clusterUUID string

	isLeaderNode    bool
	lastLeaderCheck time.Time

	es *elasticsearch6.Client
}

// Init initializes the singleton ElasticCheck.
func (e *ElasticCheck) Init(_ *config.AgentConfig, info *model.SystemInfo) {
	e.sysInfo = info

	log.Infof("elasticsearch client version: %s", elasticsearch6.Version)

	// TODO: Does this work with ES5 if we're only accessing _cat endpoints?
	// TODO: We should wrap this up in a retrier so that we try and handle intermittent cluster communication failures
	client, err := elasticsearch6.NewDefaultClient() // Accesses localhost:9200
	if err != nil {
		_ = log.Errorf("failed to create elasticsearch client: %s", err)
		return
	}

	e.es = client
	e.getClusterInfo()
	e.isLeaderNode = e.isLeader()

	log.Infof("elasticsearch node: %s (leader: %t)", e.nodeName, e.isLeaderNode)
}

// Not my fault: https://github.com/elastic/go-elasticsearch/blob/master/_examples/encoding/gjson.go
func readJsonBlob(r io.Reader) []byte {
	var b bytes.Buffer
	_, _ = b.ReadFrom(r) // TODO: Ayy
	return b.Bytes()
}

func (e *ElasticCheck) getClusterInfo() {
	esInfo, err := e.es.Info()
	if err != nil {
		log.Errorf("failed to get elasticsearch info: %s", err)
		return
	}

	defer esInfo.Body.Close()
	jsonBlob := readJsonBlob(esInfo.Body)

	// Example output:
	//{
	//  "name" : "i-ABC",
	//  "cluster_name" : "dd-test",
	//  "cluster_uuid" : "HckBgZQNOJgy8eQG8HYOSz",
	//  "version" : {
	//    "number" : "5.6.2",
	//    "build_hash" : "57e20f3",
	//    "build_date" : "2017-09-23T13:16:45.703Z",
	//    "build_snapshot" : false,
	//    "lucene_version" : "6.6.1"
	//  },
	//  "tagline" : "You Know, for Search"
	//}

	// Cluster name
	e.clusterName = gjson.GetBytes(jsonBlob, "cluster_name").String()
	if e.clusterName == "" {
		e.clusterName = "unknown"
		log.Warnf("unable to find elasticsearch cluster name")
	}

	// Cluster UUID
	e.clusterUUID = gjson.GetBytes(jsonBlob, "cluster_uuid").String()
	if e.clusterUUID == "" {
		e.clusterUUID = "unknown"
		log.Warnf("unable to find elasticsearch cluster UUID")
	}

	// Node name
	e.nodeName = gjson.GetBytes(jsonBlob, "name").String()
	if e.nodeName == "" {
		e.nodeName = "unknown"
		log.Warnf("unable to find elasticsearch node name")
	}

	log.Infof("elasticsearch cluster: %s (%s)", e.clusterName, e.clusterUUID)
}

// TODO: We can probably save cycles and skip this check on data nodes and client nodes
func (e *ElasticCheck) isLeader() bool {
	now := time.Now() // We want to check every 5m in case the leader changed
	if now.Sub(e.lastLeaderCheck) >= 5*time.Minute {
		e.lastLeaderCheck = now

		currentState, err := e.leaderCheck()
		if err != nil { // If we're error'ing now - lets use last known state.
			return e.isLeaderNode
		}

		e.isLeaderNode = currentState
	}

	return e.isLeaderNode
}

// Note: this doesn't really deal with split brains and multiple nodes thinking they're leaders... \o/
func (e *ElasticCheck) leaderCheck() (bool, error) {
	leaderInfo, err := e.es.Cat.Master(func(request *esapi.CatMasterRequest) {
		request.Format = "json"
	})

	if err != nil {
		return false, log.Warnf("failed to get elasticsearch leader info: %s", err)
	}
	defer leaderInfo.Body.Close()

	// Example output:
	//	[{"id":"8iGt13GbTR63qMBN4F4imQ","host":"172.21.119.104","ip":"172.21.119.104","node":"i-ABDE"}]

	jsonBlob := readJsonBlob(leaderInfo.Body)
	leaderNode := gjson.GetBytes(jsonBlob, "0.node").String()
	if leaderNode == "" {
		return false, log.Warnf("unable to find elasticsearch leader, defaulting to false")
	}

	return leaderNode == e.nodeName, nil
}

func (e *ElasticCheck) getShards() ([]*model.ESShard, error) {
	shardInfo, err := e.es.Cat.Shards(func(request *esapi.CatShardsRequest) {
		request.Format = "json"
		request.Bytes = "b"
		request.H = []string{"index", "shard", "prirep", "state", "docs", "store", "node", "id", "unassigned.reason"}
	})

	if err != nil {
		return nil, log.Warnf("failed to get elasticsearch shard info: %s", err)
	}
	defer shardInfo.Body.Close()

	// Example output:
	//[
	//  {
	//    "index": "my-index",
	//    "shard": "9",
	//    "prirep": "r",
	//    "state": "STARTED",
	//    "docs": "19076",
	//    "store": "54.5mb",
	//    "ip": "178.21.21.240",
	//    "node": "i-ABC"
	//  },
	//  ...
	//]

	shards := make([]*model.ESShard, 0)

	jsonBlob := readJsonBlob(shardInfo.Body)
	shardsValue := gjson.GetBytes(jsonBlob, "@this")
	if shardsValue.IsArray() {
		for _, sv := range shardsValue.Array() {
			// Get the node that this shard is on, if it is allocated
			nodes := make([]*model.ESNode, 0)
			if nodeName := gjson.Get(sv.Raw, "node").String(); nodeName != "" {
				nodes = append(nodes, &model.ESNode{
					Name: nodeName,
					Id:   gjson.Get(sv.Raw, "id").String(),
				})
			}

			// Get info for each shard
			shards = append(shards, &model.ESShard{
				IndexName:        gjson.Get(sv.Raw, "index").String(),
				Number:           int32(gjson.Get(sv.Raw, "shard").Int()),
				Primary:          gjson.Get(sv.Raw, "prirep").String() == "p",
				State:            gjson.Get(sv.Raw, "state").String(),
				Nodes:            nodes,
				DocsCount:        uint32(gjson.Get(sv.Raw, "docs").Int()),
				DocsSize:         uint32(gjson.Get(sv.Raw, "store").Int()),
				UnassignedReason: gjson.Get(sv.Raw, "unassigned.reason").String(),
			})
		}
	}

	return shards, nil
}

// Name returns the name of the ElasticCheck.
func (e *ElasticCheck) Name() string { return "elastic" }

// RealTime indicates if this check only runs in real-time mode.
func (e *ElasticCheck) RealTime() bool { return false }

func (e *ElasticCheck) Run(cfg *config.AgentConfig, groupID int32) ([]model.MessageBody, error) {
	e.Lock()
	defer e.Unlock()

	// Don't run the check if the client failed to connect
	if e.es == nil {
		return nil, errors.New("no elasticsearch client configured")
	}

	e.lastRun = time.Now()

	// If we're the leader, then great lets continue (this will also handle continuous checks in case of leader failover)
	if !e.isLeader() {
		log.Trace("this node isn't the leader, skipping check...")
		return nil, nil
	}

	shards, err := e.getShards()
	if err != nil {
		return nil, nil
	}

	// Batching shards in groups of 250, then submitting
	msgs := batchShards(cfg, 250, groupID, e.clusterName, shards)

	log.Infof("collected %d shards (submitting via %d batches)", len(shards), len(msgs))

	return msgs, nil
}

// Shards are split up into a chunks of a configured size per message to limit the message size on intake.
func batchShards(
	cfg *config.AgentConfig,
	maxBatchSize int,
	groupID int32,
	clusterName string,
	shards []*model.ESShard,
) []model.MessageBody {
	groupSize := groupSize(len(shards), maxBatchSize)
	batches := make([]model.MessageBody, 0, groupSize)

	for len(shards) > 0 {
		batchSize := min(maxBatchSize, len(shards))
		batchShards := shards[:batchSize] // Shards for this particular batch

		cs := &model.CollectorESShard{
			HostName:    cfg.HostName,
			Shards:      batchShards,
			GroupId:     groupID,
			GroupSize:   groupSize,
			ClusterName: clusterName,
		}

		batches = append(batches, cs)
		shards = shards[batchSize:]
	}

	return batches
}
