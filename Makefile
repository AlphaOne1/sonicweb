# vim: set smartindent ts=4:

# Copyright the SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

IGOOS=       $(shell go env GOOS)
IGOARCH=     $(shell go env GOARCH)
EXEC_SUFFIX= $(if $(filter windows,$(IGOOS)),.exe,)
ICGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0)

# recognize if git is available and set IBUILDTAG accordingly
GIT_AVAILABLE := $(if $(shell command -v git >/dev/null 2>&1 && echo yes),yes,no)
ifeq ($(GIT_AVAILABLE),yes)
    IBUILDTAG := $(strip $(shell git describe --tags 2>/dev/null))
endif
IBUILDTAG?= unknown

PATH:=       $(PATH):$(shell go env GOPATH)/bin
MANPAGES=    man/sonicweb.1.gz		\
             man/sonicweb_de.1.gz	\
             man/sonicweb_es.1.gz
SOURCES_FMT= '{{ range .GoFiles }} {{$$.Dir}}/{{.}} {{ end }}'
SOURCES=     $(shell go list -f $(SOURCES_FMT) ./... ) logo.tmpl


.PHONY: all clean docker test tls

all: sonic-$(IGOOS)-$(IGOARCH)$(EXEC_SUFFIX)
docker: docker-linux-amd64
package: SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm
helm: SonicWeb-$(IBUILDTAG).tgz

sonic-%: $(SOURCES)
	$(if $(filter 3,$(words $(subst -, ,$@))),,$(error Invalid executable name '$(strip $@)'; expected pattern 'sonic-<os>-<arch>'))
	go get
	GOOS=$(word   2, $(subst -, ,$@))                           \
	GOARCH=$(word 3, $(subst -, ,$@))                           \
	CGO_ENABLED=$(ICGO_ENABLED)									\
	go build -trimpath											\
			 -ldflags "-s -w -X main.buildInfoTag=$(IBUILDTAG)"	\
			 -o $@

docker-%: sonic-%
	export TARGET_OS=`  echo $< | sed -r s/'sonic-([^-]+)-([^-]+)'/'\1'/`;	\
	export TARGET_ARCH=`echo $< | sed -r s/'sonic-([^-]+)-([^-]+)'/'\2'/`;	\
	docker build --platform=$${TARGET_OS}/$${TARGET_ARCH}					\
	             -t sonicweb:$(IBUILDTAG)									\
	             --squash													\
	             .

SonicWeb-$(IBUILDTAG).tgz: $(shell find helm -type f)
	helm package --app-version "$(IBUILDTAG)" --version "$(IBUILDTAG)" helm

SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb: nfpm-$(IGOOS)-$(IGOARCH).yaml sonic-$(IGOOS)-$(IGOARCH) $(MANPAGES)
	nfpm package --config $< --packager deb --target $@

SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm: nfpm-$(IGOOS)-$(IGOARCH).yaml sonic-$(IGOOS)-$(IGOARCH) $(MANPAGES)
	nfpm package --config $< --packager rpm --target $@

nfpm-%.yaml: nfpm.yaml.tmpl
	export TARGET_OS=`  echo $@ | sed -r s/'nfpm-([^-]+)-([^-]+).yaml'/'\1'/`;	\
	export TARGET_ARCH=`echo $@ | sed -r s/'nfpm-([^-]+)-([^-]+).yaml'/'\2'/`;	\
	export TARGET_VERSION="$(IBUILDTAG)";										\
	cat $< | envsubst > $@

%.1: %.1.tmpl
	export GIT_REF_DATE=`git log -1 --date=format:"%B %Y" --format="%ad"`;	\
	export GIT_TAG=`git describe --tags`;									\
	cat $< | envsubst > $@

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

clean:
	@-rm -vf	sonic-*-*			\
				nfpm-*.yaml			\
				testcert            \
				man/sonicweb*.1.gz	\
				man/sonicweb*.1		\
				SonicWeb-*.* | sed -r s/"(.*)"/"cleaning \\1"/
