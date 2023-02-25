module github.com/atsu/goat

require (
	cloud.google.com/go v0.54.0 // indirect
	cloud.google.com/go/storage v1.6.0
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/aws/aws-sdk-go v1.29.20
	github.com/confluentinc/confluent-kafka-go v1.1.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/google/btree v1.0.0
	github.com/googleapis/gnostic v0.3.1
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/openshift/api v0.0.0-20180409145114-80a1e2bf1695
	github.com/openshift/client-go v0.0.0-20180409152027-b3f4c8b4682c
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/shirou/gopsutil v3.21.3+incompatible
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	golang.org/x/crypto v0.1.0
	google.golang.org/api v0.20.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20181004124137-fd83cbc87e76
	k8s.io/apimachinery v0.0.0-20180913025736-6dd46049f395
	k8s.io/client-go v9.0.0+incompatible
)

// use replace here to prevent other dependencies from mucking with the k8s requirements, thus preventing this build error:
// build github.com/atsu/kuba: cannot load k8s.io/api/admissionregistration/v1alpha1: cannot find module providing package k8s.io/api/admissionregistration/v1alpha1
replace (
	k8s.io/api => k8s.io/api v0.0.0-20181004124137-fd83cbc87e76
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20180913025736-6dd46049f395
	k8s.io/client-go => k8s.io/client-go v9.0.0+incompatible
	k8s.io/kubernetes => k8s.io/kubernetes v1.12.0
	k8s.io/metrics => k8s.io/metrics v0.0.0-20181004124939-8b490d15bb19
)

go 1.13
