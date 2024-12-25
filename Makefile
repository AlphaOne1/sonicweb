# vim: set smartindent ts=4:

# Copyright the SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

IGOOS=       $(shell go env GOOS)
IGOARCH=     $(shell go env GOARCH)
ICGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0)
IBUILDTAG=   $(shell git describe --tags)
PATH:=       $(PATH):$(shell go env GOPATH)/bin
MANPAGES=	 sonicweb.1.gz		\
             sonicweb_de.1.gz	\
             sonicweb_es.1.gz

.PHONY: all docker clean

all: sonic-$(IGOOS)-$(IGOARCH)
docker: docker-linux-amd64
package: SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).deb SonicWeb-$(IGOOS)-$(IGOARCH)-$(IBUILDTAG).rpm
helm: SonicWeb-$(IBUILDTAG).tgz

sonic-%: *.go logo.tmpl
	CGO_ENABLED=$(ICGO_ENABLED) go build						\
			-trimpath											\
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

clean:
	@-rm -vf	sonic-*-*		\
				nfpm-*.yaml		\
				sonicweb*.1.gz	\
				sonicweb*.1		\
				SonicWeb-*.* | sed -r s/"(.*)"/"cleaning \\1"/
