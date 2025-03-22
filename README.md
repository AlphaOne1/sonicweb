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

| Paraeter                     | Description                                 | Default |
|------------------------------|---------------------------------------------|---------|
| -root           \<path\>     | root directory of content                   | `/www`  |
| -base           \<path\>     | base path to publish the content            | `/`     |
| -port           \<port\>     | port to listen on for web requests          | `8080`  |
| -address        \<address\>  | address to listen on for web requests       | all     |
| -iport          \<port\>     | port to listen on for telemetry requests    | `8081`  |
| -iaddress       \<address\>  | address to listen on for telemetry requests | all     |
| -telemetry      {true,false} | enable/disable telemetry support            | `true`  |
| -trace-endpoint {address}    | endpoint to send trace data to              | `""`    |
| -pprof          {true,false} | enable/disable pprof support                | `false` |
| -log            \<level\>    | log level (debug, info, warn, error)        | `info`  |
| -logstyle       \<style\>    | log style (auto, text, json)                | `auto`  |
| -help                        | print the argument overview and exit        | n/a     |
| -version                     | print just version information and exit     | n/a     |

Example call, to serve the content of `testroot/` on the standard base path `/`:

```
$ ./sonic-linux-amd64 --root testroot/
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
   |/   \    _______\ Version: v1.2.1
         \  |              of: 2025-03-21T22:57:06Z
          \ |           using: go1.24.1
           \|
time=2025-03-22T00:54:37.701993 level=INFO msg="maxprocs: Leaving GOMAXPROCS=4: CPU quota undefined"
time=2025-03-22T00:54:37.702252 level=INFO msg=logging level=info
time=2025-03-22T00:54:37.702714 level=INFO msg="using root directory" root=testroot
time=2025-03-22T00:54:37.702760 level=INFO msg="using base path" path=/
time=2025-03-22T00:54:37.702775 level=INFO msg="tracing disabled"
time=2025-03-22T00:54:37.703103 level=INFO msg="registering handler for FileServer"
time=2025-03-22T00:54:37.703381 level=INFO msg="serving pprof disabled"
time=2025-03-22T00:54:37.703521 level=INFO msg="serving telemetry" address=:8081/metrics
time=2025-03-22T00:54:37.705278 level=INFO msg="starting server" address=:8080
```

Building
--------

For easier management a `Makefile` is included, using it, the build is as easy as:

```sh
make
```

If your operating system does not provide a usable form of `make`, you can also do:

```sh
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"
```
