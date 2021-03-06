module github.com/MarkWang2/BlugeLoki

go 1.14

require (
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blugelabs/bluge v0.1.7
	github.com/blugelabs/bluge_segment_api v0.2.0
	github.com/cortexproject/cortex v1.4.1-0.20201012150016-9e8beee8cacb
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/golang/snappy v0.0.1
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/klauspost/compress v1.9.5
	github.com/kr/text v0.2.0 // indirect
	github.com/miekg/dns v1.1.31 // indirect
	github.com/pierrec/lz4 v2.5.3-0.20200429092203-e876bbd321b3+incompatible
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.14.0
	github.com/prometheus/procfs v0.1.3 // indirect
	github.com/prometheus/prometheus v1.8.2-0.20200923143134-7e2db3d092f3
	github.com/stretchr/testify v1.7.0
	github.com/thanos-io/thanos v0.13.1-0.20200923175059-57035bf8f843 // indirect
	github.com/weaveworks/common v0.0.0-20200914083218-61ffdd448099
	go.etcd.io/bbolt v1.3.5-0.20200615073812-232d8fc87f50
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/grpc v1.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

//replace github.com/hpcloud/tail => github.com/grafana/tail v0.0.0-20201004203643-7aa4e4a91f03
//
//replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible
//
//replace k8s.io/client-go => k8s.io/client-go v0.18.3
//
//// >v1.2.0 has some conflict with prometheus/alertmanager. Hence prevent the upgrade till it's fixed.
//replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0

//replace k8s.io/client-go => k8s.io/client-go v0.18.3
//
//// >v1.2.0 has some conflict with prometheus/alertmanager. Hence prevent the upgrade till it's fixed.
//replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0
//
//// Use fork of gocql that has gokit logs and Prometheus metrics.
//replace github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85
//
//// Same as Cortex, we can't upgrade to grpc 1.30.0 until go.etcd.io/etcd will support it.
//replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

// Same as Cortex
// Using a 3rd-party branch for custom dialer - see https://github.com/bradfitz/gomemcache/pull/86
replace github.com/bradfitz/gomemcache => github.com/themihai/gomemcache v0.0.0-20180902122335-24332e2d58ab
