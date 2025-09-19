# vim: set smartindent ts=4:

# SPDX-FileCopyrightText: 2025 The SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

IGOOS=       $(shell go env GOOS)
IGOARCH=     $(shell go env GOARCH)
EXEC_SUFFIX= $(if $(filter windows,$(IGOOS)),.exe,)
ICGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0)

# recognize if git is available and set IBUILDTAG accordingly
GIT_AVAILABLE := $(if $(shell command -v git >/dev/null 2>&1 && echo yes),yes,no)
ifeq ($(GIT_AVAILABLE),yes)
    IBUILDTAG := $(strip $(shell git describe --tags 2>/dev/null))
	GIT_REF_DATE=$(strip $(shell git log -1 --date=format:"%B %Y" --format="%ad" 2>/dev/null))
endif
IBUILDTAG?=		unknown
GIT_REF_DATE?=	$(strip $(shell date +"%B %Y"))

PATH:=       $(PATH):$(shell go env GOPATH)/bin
MANPAGES=    man/sonicweb.1.gz		\
             man/sonicweb_de.1.gz	\
             man/sonicweb_es.1.gz
SOURCES_FMT= '{{ range .GoFiles }} {{$$.Dir}}/{{.}} {{ end }}'
SOURCES=     $(shell go list -f "$(SOURCES_FMT)" ./... ) logo.tmpl


.PHONY: all clean docker fuzz helm package test tls
.DELETE_ON_ERROR:

all: sonic-$(IGOOS)-$(IGOARCH)$(EXEC_SUFFIX)
docker: docker-linux-amd64
package: SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm
helm: SonicWeb-$(IBUILDTAG).tgz

sonic-%: $(SOURCES)
	$(if $(filter 3,$(words $(subst -, ,$@))),,$(error Invalid executable name '$(strip $@)'; expected pattern 'sonic-<os>-<arch>'))
	go mod download
	GOOS="$(word   2, $(subst -, ,$(basename $@)))"				\
	GOARCH="$(word 3, $(subst -, ,$(basename $@)))"				\
	CGO_ENABLED="$(ICGO_ENABLED)"								\
	go build -trimpath											\
			 -ldflags "-s -w -X main.buildInfoTag=$(IBUILDTAG)"	\
			 -o $@

docker-%: sonic-%
	TARGET_OS="$(word   2, $(subst -, ,$(basename $<)))"	\
	TARGET_ARCH="$(word 3, $(subst -, ,$(basename $<)))"	\
	docker build --platform=$${TARGET_OS}/$${TARGET_ARCH}	\
	             -t sonicweb:$(IBUILDTAG)					\
	             --squash									\
	             .

SonicWeb-$(IBUILDTAG).tgz: $(shell find helm -type f)
	helm package --app-version "$(IBUILDTAG)" --version "$(IBUILDTAG)" helm

SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb: nfpm-$(IGOOS)-$(IGOARCH).yaml sonic-$(IGOOS)-$(IGOARCH) $(MANPAGES)
	nfpm package --config $< --packager deb --target $@

SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm: nfpm-$(IGOOS)-$(IGOARCH).yaml sonic-$(IGOOS)-$(IGOARCH) $(MANPAGES)
	nfpm package --config $< --packager rpm --target $@

nfpm-%.yaml: nfpm.yaml.tmpl
	TARGET_OS="$(word   2, $(subst -, ,$(basename $@)))"	\
	TARGET_ARCH="$(word 3, $(subst -, ,$(basename $@)))"	\
	TARGET_VERSION="$(IBUILDTAG)"							\
	envsubst < $< > $@

%.1: %.1.tmpl
	GIT_REF_DATE="$(GIT_REF_DATE)"	\
	GIT_TAG="$(IBUILDTAG)"			\
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
	@-rm -vrf	sonic-*-*			\
				nfpm-*.yaml			\
				testcert            \
				man/sonicweb*.1.gz	\
				man/sonicweb*.1		\
				SonicWeb-*.* | sed -E s/"(.*)"/"cleaning \\1"/
