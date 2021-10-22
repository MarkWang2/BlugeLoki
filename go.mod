module github.com/MarkWang2/BlugeLoki

go 1.14

require (
	github.com/blugelabs/bluge v0.1.7
	github.com/blugelabs/bluge_segment_api v0.2.0
	github.com/cortexproject/cortex v1.4.1-0.20201012150016-9e8beee8cacb
	github.com/go-kit/kit v0.10.0
	github.com/grafana/loki v1.6.1
	github.com/klauspost/compress v1.9.5
	github.com/kr/text v0.2.0 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.14.0
	github.com/stretchr/testify v1.6.1
	github.com/weaveworks/common v0.0.0-20200914083218-61ffdd448099
	go.etcd.io/bbolt v1.3.5-0.20200615073812-232d8fc87f50
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/hpcloud/tail => github.com/grafana/tail v0.0.0-20201004203643-7aa4e4a91f03

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible

replace k8s.io/client-go => k8s.io/client-go v0.18.3

// >v1.2.0 has some conflict with prometheus/alertmanager. Hence prevent the upgrade till it's fixed.
replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0

// Use fork of gocql that has gokit logs and Prometheus metrics.
replace github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85

// Same as Cortex, we can't upgrade to grpc 1.30.0 until go.etcd.io/etcd will support it.
replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

// Same as Cortex
// Using a 3rd-party branch for custom dialer - see https://github.com/bradfitz/gomemcache/pull/86
replace github.com/bradfitz/gomemcache => github.com/themihai/gomemcache v0.0.0-20180902122335-24332e2d58ab
