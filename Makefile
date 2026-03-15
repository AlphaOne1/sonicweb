# vim: set smartindent ts=4:

# SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

EXEC_PREFIX=   sonicweb
PACKAGE_PREFIX=SonicWeb
IGOOS:=        $(shell go env GOOS)
IGOARCH:=      $(shell go env GOARCH)
EXEC_SUFFIX=   $(if $(filter windows,$(IGOOS)),.exe,)
ICGO_ENABLED=  $(if $(CGO_ENABLED),$(CGO_ENABLED),0)

# recognize if git is available and set IBUILDTAG accordingly
GIT_AVAILABLE := $(if $(shell command -v git >/dev/null 2>&1 && echo yes),yes,no)
ifeq ($(GIT_AVAILABLE),yes)
    IBUILDTAG :=  $(strip $(shell git describe --tags 2>/dev/null))
	GIT_REF_DATE:=$(strip $(shell git log -1 --date=format:"%B %Y" --format="%ad" 2>/dev/null))
endif
IBUILDTAG?=		unknown
GIT_REF_DATE?=	$(strip $(shell date +"%B %Y"))

$(info EXEC_PREFIX:   $(EXEC_PREFIX))
$(info EXEC_SUFFIX:   $(EXEC_SUFFIX))
$(info IGOOS:         $(IGOOS))
$(info IGOARCH:       $(IGOARCH))
$(info ICGO_ENABLED:  $(ICGO_ENABLED))
$(info GIT_AVAILABLE: $(GIT_AVAILABLE))
$(info IBUILDTAG:     $(IBUILDTAG))
$(info GIT_REF_DATE:  $(GIT_REF_DATE))

PATH:=       $(PATH):$(shell go env GOPATH)/bin
MANPAGES=    man/$(EXEC_PREFIX).1.gz		\
             man/$(EXEC_PREFIX)_de.1.gz	\
             man/$(EXEC_PREFIX)_es.1.gz
SOURCES_FMT= '{{ range .GoFiles }} {{$$.Dir}}/{{.}} {{ end }}'
SOURCES:=    $(shell go list -f $(SOURCES_FMT) ./... ) go.mod logo.tmpl


.PHONY: all clean docker fuzz helm package test tls
.DELETE_ON_ERROR:

all: $(EXEC_PREFIX)-$(IGOOS)-$(IGOARCH)$(EXEC_SUFFIX)
docker: docker-linux-amd64
package:	$(PACKAGE_PREFIX)-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb\
 			$(PACKAGE_PREFIX)-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm
helm: $(PACKAGE_PREFIX)-$(IBUILDTAG).tgz

define archName
$(lastword $(subst -, ,$(basename $(1))))
endef

define osName
$(lastword $(filter-out $(call archName,$(1)),$(subst -, ,$(basename $(1)))))
endef

$(EXEC_PREFIX)-%: $(SOURCES)
	$(if $(filter-out 1 2,$(words $(subst -, ,$@))),,$(error Invalid executable name '$(strip $@)'; expected pattern '$(EXEC_PREFIX)-<os>-<arch>'))
	go mod download
	GOOS="$(call osName,$@)"									\
	GOARCH="$(call archName,$@)"								\
	CGO_ENABLED="$(ICGO_ENABLED)"								\
	go build -trimpath											\
			 -ldflags "-s -w -X main.buildInfoTag=$(IBUILDTAG)"	\
			 -o $@

docker-%: $(EXEC_PREFIX)-%
	TARGET_OS="$(call osName,$<)"							\
	TARGET_ARCH="$(call archName,$<)"						\
	docker build --platform=$${TARGET_OS}/$${TARGET_ARCH}	\
	             -t $(EXEC_PREFIX):$(IBUILDTAG)					\
	             --squash									\
	             .

$(PACKAGE_PREFIX)-$(IBUILDTAG).tgz: $(shell find helm -type f)
	helm package --app-version "$(IBUILDTAG)" --version "$(IBUILDTAG)" helm

$(PACKAGE_PREFIX)-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb: nfpm-$(IGOOS)-$(IGOARCH).yaml $(EXEC_PREFIX)-$(IGOOS)-$(IGOARCH) $(MANPAGES)
	nfpm package --config $< --packager deb --target $@

$(PACKAGE_PREFIX)-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm: nfpm-$(IGOOS)-$(IGOARCH).yaml $(EXEC_PREFIX)-$(IGOOS)-$(IGOARCH) $(MANPAGES)
	nfpm package --config $< --packager rpm --target $@

nfpm-%.yaml: nfpm.yaml.tmpl
	TARGET_OS="$(call osName,$@)"		\
	TARGET_ARCH="$(call archName,$@)"	\
	TARGET_VERSION="$(IBUILDTAG)"		\
	EXEC_PREFIX="$(EXEC_PREFIX)"		\
	PACKAGE_PREFIX="$(PACKAGE_PREFIX)"	\
	envsubst < $< > $@

%.1: %.1.tmpl
	GIT_REF_DATE="$(GIT_REF_DATE)"		\
	GIT_TAG="$(IBUILDTAG)"				\
	EXEC_PREFIX="$(EXEC_PREFIX)"		\
	PACKAGE_PREFIX="$(PACKAGE_PREFIX)"	\
	envsubst < $< > $@

%.gz: %
	gzip -k -9 $<

tls:
	mkdir -p testcert
	openssl genrsa -out testcert/ca-key.pem 4096
	openssl req -new -x509 -nodes -days 7 -subj "/CN=localhost" \
				-key testcert/ca-key.pem			\
				-out testcert/ca-cert.pem
	openssl req -newkey rsa:4096 -nodes -days 7 -subj "/CN=localhost" \
				-keyout testcert/server-key.pem		\
				-out    testcert/server-req.pem
	openssl x509 -req -days 7 -set_serial 01		\
				-in    testcert/server-req.pem		\
				-out   testcert/server-cert.pem		\
				-CA    testcert/ca-cert.pem			\
				-CAkey testcert/ca-key.pem
	openssl req -newkey rsa:4096 -nodes -days 7 -subj "/CN=localhost" \
  				-keyout testcert/client-key.pem		\
   				-out    testcert/client-req.pem
	openssl x509 -req -days 7 -set_serial 01		\
				-in    testcert/client-req.pem		\
				-out   testcert/client-cert.pem		\
				-CA    testcert/ca-cert.pem			\
				-CAkey testcert/ca-key.pem
	rm testcert/*-req.pem

	#openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 7 -nodes -subj "/CN=sw-example.ex"

test:
	go test ./...


fuzz:
	go test -fuzz=Fuzz -fuzztime="30s" -fuzzminimizetime="10s" -run "^$$"

clean:
	@-rm -vrf	$(EXEC_PREFIX)-*-*		\
				nfpm-*.yaml				\
				testcert				\
				man/$(EXEC_PREFIX)*.1.gz\
				man/$(EXEC_PREFIX)*.1	\
				$(PACKAGE_PREFIX)-*.*	| sed -E s/"(.*)"/"cleaning \\1"/
