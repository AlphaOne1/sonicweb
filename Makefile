# vim: set smartindent ts=4:

# SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

PROJECT_NAME=       SonicWeb
EXEC_PREFIX=		sonicweb
PACKAGE_FILE_PREFIX=$(PROJECT_NAME)
PACKAGE_NAME=		$(EXEC_PREFIX)
THIRD_PARTY_NAME=   third_party_licenses
IGOOS:=				$(shell go env GOOS)
IGOARCH:=			$(shell go env GOARCH)
EXEC_SUFFIX=		$(if $(filter windows,$(IGOOS)),.exe,)
ICGO_ENABLED=		$(if $(CGO_ENABLED),$(CGO_ENABLED),0)

# recognize if git is available and set IBUILDTAG accordingly
GIT_AVAILABLE := $(if $(shell git status 2>/dev/null),yes,no)
ifeq ($(GIT_AVAILABLE),yes)
    IBUILDTAG :=  $(or $(strip $(shell git describe --tags 2>/dev/null)),v0.0)
	GIT_REF_DATE:=$(strip $(shell git log -1 --date=format:"%B %Y" --format="%ad" 2>/dev/null))
endif
IBUILDTAG?=		unknown
GIT_REF_DATE?=	$(strip $(shell date +"%B %Y"))

ifdef DEBUG
$(info PROJECT_NAME:        $(PROJECT_NAME))
$(info EXEC_PREFIX:         $(EXEC_PREFIX))
$(info EXEC_SUFFIX:         $(EXEC_SUFFIX))
$(info PACKAGE_FILE_PREFIX: $(PACKAGE_FILE_PREFIX))
$(info IGOOS:               $(IGOOS))
$(info IGOARCH:             $(IGOARCH))
$(info ICGO_ENABLED:        $(ICGO_ENABLED))
$(info GIT_AVAILABLE:       $(GIT_AVAILABLE))
$(info IBUILDTAG:           $(IBUILDTAG))
$(info GIT_REF_DATE:        $(GIT_REF_DATE))
endif

PATH:=       $(PATH):$(shell go env GOPATH)/bin
MANPAGES=    man/$(EXEC_PREFIX).1.gz		\
             man/$(EXEC_PREFIX)_de.1.gz	\
             man/$(EXEC_PREFIX)_es.1.gz
SOURCES_FMT= '{{ range .GoFiles }} {{$$.Dir}}/{{.}} {{ end }}'
SOURCES:=    $(shell go list -f $(SOURCES_FMT) ./... ) go.mod dir_index.html.tmpl logo.tmpl


.PHONY: all clean docker fuzz helm package test testreport tls
.DELETE_ON_ERROR:

all: $(EXEC_PREFIX)-$(IGOOS)-$(IGOARCH)$(EXEC_SUFFIX)
# Build the default Linux/amd64 image; use docker-<os>-<arch> for other targets.
docker: docker-linux-amd64
package: $(PACKAGE_FILE_PREFIX)-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb \
		$(PACKAGE_FILE_PREFIX)-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm
helm: $(PACKAGE_FILE_PREFIX)-$(IBUILDTAG).tgz

%.tar: %
	COPYFILE_DISABLE=1 tar -cf $@ $<

%.gz: %
	gzip -k -f -9 $<

%.xz: %
	xz -k -f -6 $<

define osNamePrefix
$(firstword $(subst -, ,$(basename $(patsubst $(2)-%,%,$(1)))))
endef

define osName
$(call osNamePrefix,$(1),$(EXEC_PREFIX))
endef

define archNamePrefix
$(firstword $(subst -, ,$(basename $(patsubst $(2)-$(call osNamePrefix,$(1),$(2))-%,%,$(1)))))
endef

define archName
$(call archNamePrefix,$(1),$(EXEC_PREFIX))
endef

define validSuffix
$(if $(or \
	$(findstring --,$(basename $(1))),\
	$(filter -%,$(basename $(1))),\
	$(filter %-,$(basename $(1))),\
	$(filter 0 1,$(words $(subst -, ,$(basename $(1)))))\
),$(error Invalid executable name '$(strip $(1))'; expected pattern '$(EXEC_PREFIX)-<os>-<arch>[-<variant>...]'))
endef

$(EXEC_PREFIX)-%: $(SOURCES)
	$(call validSuffix,$(patsubst $(EXEC_PREFIX)-%,%,$@))
	go mod download
	GOOS="$(call osName,$@)"									\
	GOARCH="$(call archName,$@)"								\
	CGO_ENABLED="$(ICGO_ENABLED)"								\
	go build -trimpath											\
			 -ldflags "-s -w -X main.buildInfoTag=$(IBUILDTAG)"	\
			 -o $@

$(THIRD_PARTY_NAME)-%.tar.xz: $(THIRD_PARTY_NAME)-%-dir
	ln -s $< $(patsubst %-dir,%,$<)
	COPYFILE_DISABLE=1 tar -ch $(patsubst %-dir,%,$<) | xz -9 > $@
	rm -rf $(patsubst %-dir,%,$<)

.PRECIOUS: $(THIRD_PARTY_NAME)-%-dir
$(THIRD_PARTY_NAME)-%-dir: go.mod
	rm -rf $@
	GOOS="$(call osNamePrefix,$@,$(THIRD_PARTY_NAME))"		;	\
	GOARCH="$(call archNamePrefix,$@,$(THIRD_PARTY_NAME))"	;	\
	export TMP_DIR=`mktemp -d`								&&	\
	go-licenses save ./...		\
		--force					\
		--save_path $${TMP_DIR}	\
		--ignore `go list -m`								&&	\
	go-licenses report ./...				\
	    --template license_report.md.tmpl	\
	    --ignore `go list -m`				\
	    > $${TMP_DIR}/license_report.md						&&	\
	mv $${TMP_DIR} $@										||	\
	rm -rf $${TMP_DIR}

docker-%: $(EXEC_PREFIX)-% $(THIRD_PARTY_NAME)-%.tar.xz
	TARGET_OS="$(call osName,$<)"							\
	TARGET_ARCH="$(call archName,$<)"						\
	docker build --platform=$${TARGET_OS}/$${TARGET_ARCH}	\
	             -t $(EXEC_PREFIX):$(IBUILDTAG)					\
	             --squash									\
	             .

$(PACKAGE_FILE_PREFIX)-$(IBUILDTAG).tgz: $(wildcard helm/* helm/**/*)
	helm package --app-version "$(IBUILDTAG)" --version "$(IBUILDTAG)" helm

$(PACKAGE_FILE_PREFIX)-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).%: nfpm-$(IGOOS)-$(IGOARCH).yaml $(EXEC_PREFIX)-$(IGOOS)-$(IGOARCH) $(MANPAGES) $(THIRD_PARTY_NAME)-$(IGOOS)-$(IGOARCH)-dir
	$(if $(filter deb rpm,$*),,$(error "Package type $* not supported"))
	nfpm package --config $< --packager $* --target $@

nfpm-%.yaml: nfpm.yaml.tmpl
	TARGET_OS="$(call osNamePrefix,$@,nfpm)"		\
	TARGET_ARCH="$(call archNamePrefix,$@,nfpm)"	\
	TARGET_VERSION="$(IBUILDTAG)"					\
	EXEC_PREFIX="$(EXEC_PREFIX)"					\
	PACKAGE_NAME="$(PACKAGE_NAME)"					\
	PROJECT_NAME="$(PROJECT_NAME)"					\
	envsubst < $< > $@

%.1: %.1.tmpl
	GIT_REF_DATE="$(GIT_REF_DATE)"	\
	GIT_TAG="$(IBUILDTAG)"			\
	EXEC_PREFIX="$(EXEC_PREFIX)"	\
	PROJECT_NAME="$(PROJECT_NAME)"	\
	envsubst < $< > $@

tls:
	mkdir -p testcert
	openssl ecparam -name prime256v1 -genkey -noout -out testcert/ca-key.pem
	openssl req -new -x509 -nodes -days 7 -subj "/CN=localhost" \
				-key testcert/ca-key.pem			\
				-out testcert/ca-cert.pem
	openssl ecparam -name prime256v1 -genkey -noout -out testcert/server-key.pem
	openssl req -new -subj "/CN=localhost" 			\
				-key    testcert/server-key.pem		\
				-out    testcert/server-req.pem
	openssl x509 -req -days 7 -set_serial 01		\
				-in    testcert/server-req.pem		\
				-out   testcert/server-cert.pem		\
				-CA    testcert/ca-cert.pem			\
				-CAkey testcert/ca-key.pem
	openssl ecparam -name prime256v1 -genkey -noout -out testcert/client-key.pem
	openssl req -new -subj "/CN=localhost" 			\
  				-key    testcert/client-key.pem		\
   				-out    testcert/client-req.pem
	openssl x509 -req -days 7 -set_serial 01		\
				-in    testcert/client-req.pem		\
				-out   testcert/client-cert.pem		\
				-CA    testcert/ca-cert.pem			\
				-CAkey testcert/ca-key.pem
	rm testcert/*-req.pem

	#openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -keyout server.key -out server.crt -days 7 -nodes -subj "/CN=sw-example.ex"

test:
	go test ./...

testreport:
	go tool gotestsum --junitfile junit.xml --	\
	    -race									\
	    -v `go list ./... | grep -v example`	\
	    --covermode=atomic						\
	    --coverpkg=./...						\
	    --coverprofile=coverage.txt

fuzz:
	go test -fuzz=Fuzz -fuzztime="30s" -fuzzminimizetime="10s" -run "^$$"

clean:
	@-rm -vrf	$(EXEC_PREFIX)-*-*			\
				nfpm-*.yaml					\
				testcert					\
				$(THIRD_PARTY_NAME)*       \
				man/$(EXEC_PREFIX)*.1.gz	\
				man/$(EXEC_PREFIX)*.1		\
				$(PACKAGE_FILE_PREFIX)-*.*	| sed -E s/"(.*)"/"cleaning \\1"/
