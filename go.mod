module github.com/qlik-oss/sense-installer

go 1.13

require (
	github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/PaesslerAG/jsonpath v0.1.1 // indirect
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/carolynvs/datetime-printer v0.2.0 // indirect
	github.com/cbroglie/mustache v1.0.1 // indirect
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/containerd/continuity v0.0.0-20191127005431-f65d91d395eb // indirect
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c // indirect
	github.com/containerd/ttrpc v0.0.0-20191028202541-4f1b8fe65a5c // indirect
	github.com/containerd/typeurl v0.0.0-20190911142611-5eb25027c9fd // indirect
	github.com/deislabs/cnab-go v0.7.1-beta1 // indirect
	github.com/deislabs/porter v0.21.1-beta.1
	github.com/docker/cli v0.0.0-20191210190034-08eaead288c1 // indirect
	github.com/docker/cnab-to-oci v0.3.0-beta2 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8 // indirect
	github.com/gobuffalo/packr/v2 v2.7.1 // indirect
	github.com/gogo/googleapis v1.3.0 // indirect
	github.com/google/go-containerregistry v0.0.0-20191209173213-73724ba06bda // indirect
	github.com/gophercloud/gophercloud v0.7.0 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/mmcdole/gofeed v1.0.0-beta2 // indirect
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.0.1 // indirect
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c // indirect
	github.com/qri-io/jsonschema v0.1.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.7 // indirect
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)

replace (
	github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
