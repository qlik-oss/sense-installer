module github.com/qlik-oss/sense-installer

go 1.13

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a

	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51

	sigs.k8s.io/kustomize/api => github.com/qlik-oss/kustomize/api v0.3.3-0.20200206224201-2e697eccbad9
)

require (
	cloud.google.com/go v0.52.0 // indirect
	cloud.google.com/go/storage v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.0.3
	github.com/Shopify/ejson v1.2.1
	github.com/aws/aws-sdk-go v1.28.9 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/containerd/continuity v0.0.0-20191214063359-1097c8bae83b // indirect
	github.com/containers/image/v5 v5.1.0
	github.com/docker/cli v0.0.0-20191212191748-ebca1413117a // indirect
	github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gobuffalo/logger v1.0.3 // indirect
	github.com/gobuffalo/packd v1.0.0 // indirect
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/mattn/go-colorable v0.1.4
	github.com/mattn/go-tty v0.0.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/qlik-oss/k-apis v0.0.25
	github.com/robfig/cron/v3 v3.0.1
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.1
	github.com/ttacon/chalk v0.0.0-20160626202418-22c06c80ed31
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4 // indirect
	golang.org/x/exp v0.0.0-20200119233911-0405dc783f0a // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/tools v0.0.0-20200312194400-c312e98713c2 // indirect
	google.golang.org/genproto v0.0.0-20200128133413-58ce757ed39b // indirect
	google.golang.org/grpc v1.27.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 v3.0.0-20190924164351-c8b7dadae555
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	sigs.k8s.io/kustomize/api v0.3.2
	sigs.k8s.io/yaml v1.1.0
)

exclude github.com/Azure/go-autorest v12.0.0+incompatible
