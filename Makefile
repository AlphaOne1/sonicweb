.PHONY: all clean

all: sonic

sonic: *.go logo.tmpl
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"

clean:
	-rm -f sonic sonic.exe