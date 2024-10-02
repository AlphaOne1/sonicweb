# Copyright the sonicweb contributors.
# SPDX-License-Identifier: MPL-2.0

.PHONY: all clean

all: sonic

IGOOS=       $(shell go env GOOS)
IGOARCH=     $(shell go env GOARCH)
ICGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0)

sonic: *.go logo.tmpl
	CGO_ENABLED=$(ICGO_ENABLED) go build -trimpath -ldflags "-s -w" -o $@-$(IGOOS)-$(IGOARCH)

clean:
	-rm -f sonic-$(IGOOS)-$(IGOARCH)