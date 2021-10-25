package api

import (
	"context"
	"encoding/json"
	"github.com/MarkWang2/BlugeLoki/storage/stores/shipper"
	"github.com/MarkWang2/BlugeLoki/storage/stores/util"
	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/cortexproject/cortex/pkg/chunk/local"
	"github.com/prometheus/common/model"
	"github.com/weaveworks/common/server"
	"io/ioutil"
	"net/http"
	"time"
)

type API struct {
	server *server.Server
}

func New() *API {
	// cfg.HTTPListenAddress, cfg.HTTPListenPort
	cfg := server.Config{HTTPListenPort: 8080}
	wws, _ := server.New(cfg)
	s := &API{server: wws}
	s.RegisterRoute()
	//s.server.Run()
	return &API{server: wws}
}

// Register
func (a *API) RegisterRoute() {
	a.server.HTTP.Path("/").Handler(http.HandlerFunc(a.WriteLogs))
}

func (a *API) Run() {
	a.server.Run()
}

//Run()

// ready serves the ready endpoint
func (a *API) WriteLogs(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var t interface{}
	decoder.Decode(&t)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	fsObjectClient, _ := local.NewFSObjectClient(local.FSConfig{Directory: "./obstore"})
	config := shipper.Config{
		ActiveIndexDirectory: "snpsegindex",
		SharedStoreType:      "filesystem",
		CacheLocation:        "dcache",
		CacheTTL:             30 * time.Second,
		ResyncInterval:       20 * time.Second,
		IngesterName:         "wang",
		Mode:                 shipper.ModeReadWrite,
	}

	//body = []byte(`{"@timestamp":"0001-01-01T00:00:00.000+00:00","@metadata":{"beat":"test","type":"_doc","version":"1.2.3"},"msg":"message"}`)

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
	//data := []byte(`
	//{
	//	"id": 123,
	//	"fname": "John",
	//	"height": 1.75,
	//	"male": true,
	//	"languages": null,
	//	"subjects": [ "Math", "Science" ],
	//	"profile": {
	//		"uname": "johndoe91",
	//		"f_count": 1975
	//	}
	//}`)
	//json.Unmarshal()

	var event util.Event

	json.Unmarshal(body, &event)

	cfg := chunk.DefaultSchemaConfig("", "v10", 0)
	ts, _ := time.Parse(time.RFC3339, "1970-09-16T00:00:00Z")
	tbName := cfg.Configs[0].IndexTables.TableFor(model.TimeFromUnix(ts.Unix()))
	s, _ := shipper.NewShipper(config, fsObjectClient, nil)
	wr := s.NewWriteBatch()
	event.Fields["@timestamp"] = event.Timestamp

	tbName = cfg.Configs[0].IndexTables.TableFor(model.TimeFromUnix(event.Timestamp.Unix()))
	wr.AddEvent(tbName, event.Fields)
	//wr.AddJson(tbName, body)
	s.BatchWrite(context.Background(), wr)

	//if s.healthCheckTarget && !s.tms.Ready() {
	//	http.Error(rw, readinessProbeFailure, http.StatusInternalServerError)
	//	return
	//}
	rw.WriteHeader(http.StatusOK)
}

func (a *API) foo() {
	fsObjectClient, _ := local.NewFSObjectClient(local.FSConfig{Directory: "./obstore"})
	config := shipper.Config{
		ActiveIndexDirectory: "snpsegindex",
		SharedStoreType:      "filesystem",
		CacheLocation:        "dcache",
		CacheTTL:             30 * time.Second,
		ResyncInterval:       20 * time.Second,
		IngesterName:         "wang",
		Mode:                 shipper.ModeReadWrite,
	}

	cfg := chunk.DefaultSchemaConfig("", "v10", 0)
	ts, _ := time.Parse(time.RFC3339, "1970-09-16T00:00:00Z")
	tbName := cfg.Configs[0].IndexTables.TableFor(model.TimeFromUnix(ts.Unix()))
	s, _ := shipper.NewShipper(config, fsObjectClient, nil)
	// new struct
	wr := s.NewWriteBatch()
	// create new table 没有对应表明创建
	// wr.Add(tbName, "test", []byte("test"), []byte("test"))
	data := []byte(`
	{
		"id": 123,
		"fname": "John",
		"height": 1.75,
		"male": true,
		"languages": null,
		"subjects": [ "Math", "Science" ],
		"profile": {
			"uname": "johndoe91",
			"f_count": 1975
		}
	}`)

	wr.AddJson(tbName, data)

	s.BatchWrite(context.Background(), wr)
}

//config := shipper.Config{
//	ActiveIndexDirectory: "snpsegindex",
//	SharedStoreType:      "filesystem",
//	CacheLocation:        "dcache",
//	CacheTTL:             30 * time.Second,
//	ResyncInterval:       20 * time.Second,
//	IngesterName:         "wang",
//	Mode:                 shipper.ModeReadWrite,
//}
//
//cfg := chunk.DefaultSchemaConfig("", "v10", 0)
//ts, _ := time.Parse(time.RFC3339, "1970-09-16T00:00:00Z")
//tbName := cfg.Configs[0].IndexTables.TableFor(model.TimeFromUnix(ts.Unix()))
//s, _ := shipper.NewShipper(config, fsObjectClient, nil)
//// new struct
//wr := s.NewWriteBatch()
//// create new table 没有对应表明创建
//// wr.Add(tbName, "test", []byte("test"), []byte("test"))
//data := []byte(`
//{
//	"id": 123,
//	"fname": "John",
//	"height": 1.75,
//	"male": true,
//	"languages": null,
//	"subjects": [ "Math", "Science" ],
//	"profile": {
//		"uname": "johndoe91",
//		"f_count": 1975
//	}
//}`)
//
//wr.AddJson(tbName, data)
//
//s.BatchWrite(context.Background(), wr)

//// cfg.HTTPListenAddress, cfg.HTTPListenPort
//cfg := server.Config{HTTPListenPort: 8080}
//wws, _ := server.New(cfg)
//
//s := &API{server: wws}
//s.foo()
//s.server.Run()
