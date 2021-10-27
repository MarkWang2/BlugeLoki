package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MarkWang2/BlugeLoki/storage/stores/shipper"
	"github.com/MarkWang2/BlugeLoki/storage/stores/util"
	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/cortexproject/cortex/pkg/chunk/aws"
	"github.com/cortexproject/cortex/pkg/chunk/azure"
	"github.com/cortexproject/cortex/pkg/chunk/gcp"
	"github.com/cortexproject/cortex/pkg/chunk/local"
	"github.com/cortexproject/cortex/pkg/chunk/openstack"
	"github.com/prometheus/common/model"
	"github.com/weaveworks/common/server"
	"io/ioutil"
	"net/http"
	"time"
)

type API struct {
	server *server.Server
	config *Config
}

const (
	StorageTypeAWS        = "aws"
	StorageTypeAzure      = "azure"
	StorageTypeInMemory   = "inmemory"
	StorageTypeFileSystem = "filesystem"
	StorageTypeGCS        = "gcs"
	StorageTypeS3         = "s3"
	StorageTypeSwift      = "swift"
)

// Config chooses which storage client to use.
type Config struct {
	ObjectStoreName    string
	HTTPListenPort     int
	Engine             string                  `yaml:"engine"`
	AWSStorageConfig   aws.StorageConfig       `yaml:"aws"`
	AzureStorageConfig azure.BlobStorageConfig `yaml:"azure"`
	GCPStorageConfig   gcp.Config              `yaml:"bigtable"`
	GCSConfig          gcp.GCSConfig           `yaml:"gcs"`
	FSConfig           local.FSConfig          `yaml:"filesystem"`
	Swift              openstack.SwiftConfig   `yaml:"swift"`
}

func (a *API) ObjectClient() (chunk.ObjectClient, error) {
	cfg := a.config
	switch cfg.ObjectStoreName {
	case StorageTypeAWS, StorageTypeS3:
		return aws.NewS3ObjectClient(cfg.AWSStorageConfig.S3Config)
	case StorageTypeGCS:
		return gcp.NewGCSObjectClient(context.Background(), cfg.GCSConfig)
	case StorageTypeAzure:
		return azure.NewBlobStorage(&cfg.AzureStorageConfig)
	case StorageTypeSwift:
		return openstack.NewSwiftObjectClient(cfg.Swift)
	case StorageTypeInMemory:
		return chunk.NewMockStorage(), nil
	case StorageTypeFileSystem:
		return local.NewFSObjectClient(cfg.FSConfig)
	default:
		return nil, fmt.Errorf("Unrecognized storage client %v, choose one of: %v, %v, %v, %v, %v", cfg.ObjectStoreName, StorageTypeAWS, StorageTypeS3, StorageTypeGCS, StorageTypeAzure, StorageTypeFileSystem)
	}
}

func (a *API) NewShipper() (*shipper.Shipper, error) {
	objectClient, _ := a.ObjectClient()
	config := shipper.Config{
		ActiveIndexDirectory: "snpsegindex",
		SharedStoreType:      a.config.ObjectStoreName,
		CacheLocation:        "dcache",
		CacheTTL:             30 * time.Second,
		ResyncInterval:       20 * time.Second,
		IngesterName:         "wang",
		Mode:                 shipper.ModeReadWrite,
	}
	return shipper.NewShipper(config, objectClient, nil)
}

func New() *API {
	config := &Config{HTTPListenPort: 8080, ObjectStoreName: StorageTypeFileSystem}
	cfg := server.Config{HTTPListenPort: 8080}
	wws, _ := server.New(cfg)
	s := &API{server: wws, config: config}
	s.RegisterRoute()
	return s
}

// Register
func (a *API) RegisterRoute() {
	a.server.HTTP.Path("/").Handler(http.HandlerFunc(a.WriteLogs))
}

func (a *API) Run() {
	a.server.Run()
}

// ready serves the ready endpoint
func (a *API) WriteLogs(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var t interface{}
	decoder.Decode(&t)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	body = []byte(`{
		"Timestamp": "2017-06-29T13:58:28Z",
		"Meta": null,
		"Fields": {
		"destination": {
			"ip": "10.99.252.50",
				"locality": "internal",
				"port": 53
		},
		"event": {
			"action": "netflow_flow",
				"category": [
	"network_traffic",
	"network"
	],
	"duration": 20269000000,
	"kind": "event",
	"type": [
	"connection"
	]
	},
	"flow": {
	"id": "2vFIarATx_4",
	"locality": "internal"
	},
	"netflow": {
	"destination_ipv4_address": "10.99.252.50",
	"destination_transport_port": 53,
	"egress_interface": 26092,
	"exporter": {
	"address": "192.0.2.1:4444",
	"source_id": 0,
	"timestamp": "2017-06-29T13:58:28Z",
	"uptime_millis": 0,
	"version": 10
	},
	"firewall_event": 2,
	"flow_duration_milliseconds": 20269,
	"flow_end_sys_up_time": 2395395322,
	"flow_start_sys_up_time": 2395375053,
	"ingress_interface": 48660,
	"octet_delta_count": 0,
	"octet_total_count": 65,
	"packet_delta_count": 0,
	"packet_total_count": 1,
	"protocol_identifier": 17,
	"source_ipv4_address": "10.99.130.239",
	"source_mac_address": "00:00:00:00:00:00",
	"source_transport_port": 65105,
	"type": "netflow_flow"
	},
	"network": {
	"bytes": 0,
	"community_id": "1:hn30QwbDmwNihxKr9rCALGUWPgE=",
	"direction": "unknown",
	"iana_number": 17,
	"packets": 0,
	"transport": "udp"
	},
	"observer": {
	"ip": "192.0.2.1"
	},
	"related": {
	"ip": [
	"10.99.130.239",
	"10.99.252.50"
	]
	},
	"source": {
	"bytes": 0,
	"ip": "10.99.130.239",
	"locality": "internal",
	"mac": "00:00:00:00:00:00",
	"packets": 0,
	"port": 65105
	}
	},
	"Private": null,
	"TimeSeries": false
	}`)

	var event util.Event

	json.Unmarshal(body, &event)

	cfg := chunk.DefaultSchemaConfig("", "v10", 0)
	s, _ := a.NewShipper()
	wr := s.NewWriteBatch()
	event.Fields["@timestamp"] = event.Timestamp

	tbName := cfg.Configs[0].IndexTables.TableFor(model.TimeFromUnix(event.Timestamp.Unix()))
	wr.AddEvent(tbName, event.Fields)
	s.BatchWrite(context.Background(), wr)

	//if s.healthCheckTarget && !s.tms.Ready() {
	//	http.Error(rw, readinessProbeFailure, http.StatusInternalServerError)
	//	return
	//}
	rw.WriteHeader(http.StatusOK)
}
