SonicWeb
========

*SonicWeb* is a lightweight, easy-to-use web server for static content.

Features
--------

* statically linked, apt to be used in scratch containers (~13MB)
* focused purpose, thus little attack surface
* usage of OWASP [coraza](https://github.com/corazawaf/coraza) middleware
  to follow best security practises
* easy integration in monitoring using [Prometheus](prometheus.io) and/or
  [Jaeger Tracing](jaegertracing.io)
* no complications with configuration files

Usage
-----

*SonicWeb* is controlled solely by command line arguments. They are as follows:

| Paraeter                | Description                                 | Default |
|-------------------------|---------------------------------------------|---------|
| -root      \<path\>     | root directory of content                   | `/www`  |
| -base      \<path\>     | base path to publish the content            | `/`     |
| -port      \<port\>     | port to listen on for web requests          | `8080`  |
| -address   \<address\>  | address to listen on for web requests       | all     |
| -iport     \<port\>     | port to listen on for telemetry requests    | `8081`  |
| -iaddress  \<address\>  | address to listen on for telemetry requests | all     |
| -telemetry {true,false} | enable/disable telemetry support            | `true`  |
| -pprof     {true,false} | enable/disable pprof support                | `false` |
| -log       \<level\>    | log level (debug, info, warn, error)        | `info`  |
| -logstyle  \<style\>    | log style (auto, text, json)                | `auto`  |

Example call, to serve the content of `testroot/` on the standard base path `/`:

```
             |\
             ||\
    _________||\\
    \            \   /|
     \     ___    \ / |
    /     /.-.\   _V__|             _     _       __     __
   /     //   \  / ___/____  ____  (_)___| |     / /__  / /_
  /___  // _  |  \__ \/ __ \/ __ \/ / ___/ | /| / / _ \/ __ \
     |   \(_)/  ___/ / /_/ / / / / / /__ | |/ |/ /  __/ /_/ /
     |  , \_/  /____/\____/_/ /_/_/\___/ |__/|__/\___/_.___/
     | / \           \
     |/   \    _______\ Version: 926891b9d3a0672044a7de3e12d567f0e6763af6*
           \  |              of: 2024-10-02T22:10:33Z
            \ |           using: go1.23.1
             \|
time=2024-10-03T09:19:29.031699 level=INFO source=sonic/main.go:75 msg="maxprocs: Leaving GOMAXPROCS=4: CPU quota undefined"
time=2024-10-03T09:19:29.032050 level=INFO source=sonic/main.go:104 msg=logging level=info
time=2024-10-03T09:19:29.032357 level=INFO source=sonic/main.go:118 msg="using root directory" root=testroot/
time=2024-10-03T09:19:29.032433 level=INFO source=sonic/main.go:133 msg="using base path" path=/
time=2024-10-03T09:19:29.032769 level=INFO source=sonic/instrumentation.go:72 msg="serving pprof disabled"
time=2024-10-03T09:19:29.032837 level=INFO source=sonic/instrumentation.go:76 msg="serving telemetry" address=:8081/metrics
time=2024-10-03T09:19:29.033072 level=INFO source=sonic/main.go:151 msg="registering handler for FileServer"
time=2024-10-03T09:19:29.033176 level=INFO source=sonic/main.go:180 msg="starting server"
```

Building
--------

For easier management a `Makefile` is included, using it, the build is as easy as:

```sh
make
```

If your operating system does not provide a usable form of `make`, you can also do:

```sh
CGO_ENABLED=0 go build -trimpath-ldflags "-s -w"
```
