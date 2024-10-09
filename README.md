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
$ ./sonic-linux-amd64 --root testroot
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
   |/   \    _______\ Version: v1.0.0-32-g77369ff*
         \  |              of: 2024-10-08T23:24:30Z
          \ |           using: go1.23.2
           \|
time=2024-10-10T01:17:15.409549 level=INFO source=sonic/main.go:78 msg="maxprocs: Leaving GOMAXPROCS=4: CPU quota undefined"
time=2024-10-10T01:17:15.411215 level=INFO source=sonic/main.go:157 msg=logging level=info
time=2024-10-10T01:17:15.412224 level=INFO source=sonic/main.go:171 msg="using root directory" root=testroot
time=2024-10-10T01:17:15.412313 level=INFO source=sonic/main.go:186 msg="using base path" path=/
time=2024-10-10T01:17:15.412579 level=INFO source=sonic/main.go:195 msg="registering handler for FileServer"
time=2024-10-10T01:17:15.413270 level=INFO source=sonic/main.go:204 msg="starting server"
time=2024-10-10T01:17:15.413799 level=INFO source=sonic/instrumentation.go:72 msg="serving pprof disabled"
time=2024-10-10T01:17:15.413839 level=INFO source=sonic/instrumentation.go:76 msg="serving telemetry" address=:8081/metrics
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
