# Copyright the SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

IGOOS=       $(shell go env GOOS)
IGOARCH=     $(shell go env GOARCH)
ICGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0)
IBUILDTAG=   $(shell git describe --tags)

.PHONY: all docker clean

all: sonic-$(IGOOS)-$(IGOARCH)
docker: docker-linux-amd64

sonic-%: *.go logo.tmpl
	CGO_ENABLED=$(ICGO_ENABLED) go build						\
			-trimpath											\
			-ldflags "-s -w -X main.buildInfoTag=$(IBUILDTAG)"	\
			-o $@

docker-%: sonic-%
	export TARGET_OS=`  echo $< | sed -r s/'sonic-([^-]+)-([^-]+)'/'\1'/`; \
	export TARGET_ARCH=`echo $< | sed -r s/'sonic-([^-]+)-([^-]+)'/'\2'/`; \
	docker build --platform=$${TARGET_OS}/$${TARGET_ARCH}                  \
	             -t sonicweb:$(IBUILDTAG)                                  \
	             --squash                                                  \
	             .

clean:
	@-rm -vf sonic-*-* | sed -r s/"(.*)"/"cleaning \\1"/