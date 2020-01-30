PKG = github.com/qlik-oss/sense-installer

# --no-print-directory avoids verbose logging when invoking targets that utilize sub-makes
MAKE_OPTS ?= --no-print-directory

LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT) -X "$(PKG)/pkg.CommitDate=$(COMMIT_DATE)"
XBUILD = CGO_ENABLED=0 go build -a -tags "$(BUILDTAGS)" -ldflags '$(LDFLAGS)'
BINDIR = bin

COMMIT ?= $(shell git rev-parse --short HEAD)
COMMIT_DATE ?= $(shell git show --no-patch --no-notes --pretty='%cd' $(COMMIT) --date=iso)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo v0)
PERMALINK ?= $(shell git describe --tags --exact-match &> /dev/null && echo latest || echo canary)
BUILDTAGS = netgo containers_image_ostree_stub exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp

CLIENT_PLATFORM ?= $(shell go env GOOS)
CLIENT_ARCH ?= $(shell go env GOARCH)
RUNTIME_PLATFORM ?= linux
RUNTIME_ARCH ?= amd64
# NOTE: When we add more to the build matrix, update the regex for porter mixins feed generate
SUPPORTED_PLATFORMS = linux darwin windows
SUPPORTED_ARCHES = amd64

MIXIN = qliksense

ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
else ifeq ($(RUNTIME_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

.PHONY: build
build: clean generate
	mkdir -p $(BINDIR)
	go build -ldflags '$(LDFLAGS)' -tags "$(BUILDTAGS)" -o $(BINDIR)/$(MIXIN)$(FILE_EXT) ./cmd/$(MIXIN)

xbuild-all: clean generate
	$(foreach OS, $(SUPPORTED_PLATFORMS), \
    	$(foreach ARCH, $(SUPPORTED_ARCHES), \
            	$(MAKE) $(MAKE_OPTS) CLIENT_PLATFORM=$(OS) CLIENT_ARCH=$(ARCH) MIXIN=$(MIXIN) xbuild; \
    	))
	$(MAKE) clean-packr
xbuild: $(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)
$(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT):
	mkdir -p $(dir $@)
	GOOS=$(CLIENT_PLATFORM) GOARCH=$(CLIENT_ARCH) $(XBUILD) -o $@ ./cmd/$(MIXIN)


generate: get-crds packr2
	go generate ./...

HAS_PACKR2 := $(shell command -v packr2)
packr2:
ifndef HAS_PACKR2
	go get -u github.com/gobuffalo/packr/v2/packr2
endif

clean: clean-packr
	-rm -fr bin/
	-rm -rf /tmp/operator
	-rm -fr pkg/qliksense/crds

clean-packr: packr2
	cd pkg/qliksense && packr2 clean

get-crds: 
	git clone git@github.com:qlik-oss/qliksense-operator.git -b ms-3 /tmp/operator
	mkdir -p pkg/qliksense/crds/cr
	mkdir -p pkg/qliksense/crds/crd
	mkdir -p pkg/qliksense/crds/crd-deploy
	cp /tmp/operator/deploy/*.yaml pkg/qliksense/crds/crd-deploy
	cp /tmp/operator/deploy/crds/*_crd.yaml pkg/qliksense/crds/crd
	cp /tmp/operator/deploy/crds/*_cr.yaml pkg/qliksense/crds/cr
