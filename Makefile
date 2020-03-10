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

DEVNUL := /dev/null
WHICH := which

ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
DEVNUL := NUL
WHICH := where
else ifeq ($(RUNTIME_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

.PHONY: build
build: clean generate
	mkdir -p $(BINDIR)
	go build -ldflags '$(LDFLAGS)' -tags "$(BUILDTAGS)" -o $(BINDIR)/$(MIXIN)$(FILE_EXT) ./cmd/$(MIXIN)
	$(MAKE) clean

.PHONY: test
test: clean generate
ifeq ($(shell ${WHICH} docker-registry 2>${DEVNUL}),)
	$(eval TMP := $(shell mktemp -d))
	git clone https://github.com/docker/distribution.git $(TMP)/docker-distribution
	cd $(TMP)/docker-distribution; git checkout -b v2.7.1; make
	cp $(TMP)/docker-distribution/bin/registry pkg/qliksense/docker-registry
	-rm -rf $(TMP)/docker-distribution
endif
	go test -short -count=1 -tags "$(BUILDTAGS)" -v ./...
	$(MAKE) clean

xbuild-all: clean generate
	$(foreach OS, $(SUPPORTED_PLATFORMS), \
    	$(foreach ARCH, $(SUPPORTED_ARCHES), \
            	$(MAKE) $(MAKE_OPTS) CLIENT_PLATFORM=$(OS) CLIENT_ARCH=$(ARCH) MIXIN=$(MIXIN) xbuild; \
    	))
	$(MAKE) clean
xbuild: $(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)
$(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT):
	mkdir -p $(dir $@)
	GOOS=$(CLIENT_PLATFORM) GOARCH=$(CLIENT_ARCH) $(XBUILD) -o $@ ./cmd/$(MIXIN)
	tar -czvf $(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH).tar.gz -C $(BINDIR)/$(VERSION)/ $(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)
	#tar -C $(BINDIR)/$(VERSION)/ -cvf $(BINDIR)/$(VERSION)/$(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH).tar.gz $(MIXIN)-$(CLIENT_PLATFORM)-$(CLIENT_ARCH)$(FILE_EXT)


generate: get-crds packr2
	go generate ./...

packr2:
ifeq ($(shell ${WHICH} packr2 2>${DEVNUL}),)
	go get -u github.com/gobuffalo/packr/v2/packr2@v2.7.1
endif

clean: clean-packr
	-rm -fr pkg/qliksense/crds

clean-packr: packr2
	cd pkg/qliksense && packr2 clean

get-crds:
	$(eval TMP := $(shell mktemp -d))
	git clone git@github.com:qlik-oss/qliksense-operator.git -b ms-3 $(TMP)/operator
	mkdir -p pkg/qliksense/crds/cr
	mkdir -p pkg/qliksense/crds/crd
	mkdir -p pkg/qliksense/crds/crd-deploy
	cp $(TMP)/operator/deploy/*.yaml pkg/qliksense/crds/crd-deploy
	cp $(TMP)/operator/deploy/crds/*_crd.yaml pkg/qliksense/crds/crd
	cp $(TMP)/operator/deploy/crds/*_cr.yaml pkg/qliksense/crds/cr
	-rm -rf $(TMP)/operator
