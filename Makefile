# Copyright the SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

IGOOS=       $(shell go env GOOS)
IGOARCH=     $(shell go env GOARCH)
ICGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0)
IBUILDTAG=   $(shell git describe --tags)

.PHONY: all docker clean

all: sonic-$(IGOOS)-$(IGOARCH) docker

sonic-%: *.go logo.tmpl
	CGO_ENABLED=$(ICGO_ENABLED) go build						\
			-trimpath											\
			-ldflags "-s -w -X main.buildInfoTag=$(IBUILDTAG)"	\
			-o $@

docker: sonic-linux-amd64
	docker build --platform=`echo $< | sed -r s/'sonic-([^-]+)-([^-]+)'/'\1\/\2'/` -t sonicweb:$(IBUILDTAG) --squash .

clean:
	@-rm -vf sonic-*-* | sed -r s/"(.*)"/"cleaning \\1"/