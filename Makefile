.PHONY: all clean

all: sonic

sonic: *.go logo.tmpl
	CGO_ENABLED=0 go build -ldflags "-s -w"

clean:
	-rm sonic