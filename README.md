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

| Paraeter                | Description                                  |
|-------------------------|----------------------------------------------|
| -root      \<path\>     | root directory of content,                   |
|                         | defaults to `/www`                           |
| -base      \<path\>     | base path to publish the content,            |
|                         | defaults to `/`                              |
| -port      \<port\>     | port to listen on for web requests,          |
|                         | defaults to `8080`                           |
| -address   \<address\>  | address to listen on for web requests,       |
|                         | defaults to all                              |
| -iport     \<port\>     | port to listen on for telemetry requests,    |
|                         | defaults to `8081`                           |
| -iaddress  \<address\>  | address to listen on for telemetry requests, |
|                         | defaults to all                              |
| -telemetry {true,false} | enable/disable telemetry support,            |
|                         | defaults to `true`                           |
| -pprof     {true,false} | enable/disable pprof support,                |
|                         | defaults to `false`                          |
| -log       \<level\>    | log level (debug, info, warn, error),        |
|                         | defaults to `info`                           |
| -logstyle  \<style\>    | log style (auto, text, json),                |
|                         | defaults to `auto`                           |

Example call, to serve the content of `testroot/` on the standard base path `/`:

```
$ ./sonic -root testroot/
               /\______
            .-//\\     `'--__
          /´ // ||        _,/
        /´  //__||      ,/
______/´    __         /____             _     _       __     __
\    /    /'_ '\      / ___/____  ____  (_)___| |     / /__  / /_
 \  /    / / '\ \     \__ \/ __ \/ __ \/ / ___/ | /| / / _ \/ __ \
  \/      / _  \     ___/ / /_/ / / / / / /__ | |/ |/ /  __/ /_/ /
   \ .   / | \ /_   /____/\____/_/ /_/_/\___/ |__/|__/\___/_.___/
    \|\  |  \ // \       `\
     \ \-'__-/ _/ \        `\   Version: 3e2283cb1a2b3322493002265b9f932aecb0efc5*
     @@_       _--/-----_    `\      of: 2024-06-11T15:55:44Z
        `-----´          `'-_  \  using: go1.22.4
                             `-_\

time=2024-06-12T01:42:20.092282 level=INFO source=/home/dev/sonic/main.go:119 msg=logging level=debug
time=2024-06-12T01:42:20.093238 level=INFO source=/home/dev/sonic/main.go:133 msg="using root directory" root=testroot/
time=2024-06-12T01:42:20.093335 level=INFO source=/home/dev/sonic/main.go:148 msg="using base path" path=/
time=2024-06-12T01:42:20.093776 level=INFO source=/home/dev/sonic/main.go:166 msg="registering handler for FileServer"
time=2024-06-12T01:42:20.093884 level=INFO source=/home/dev/sonic/main.go:195 msg="starting server"
time=2024-06-12T01:42:20.094049 level=INFO source=/home/dev/sonic/instrumentation.go:69 msg="serving pprof disabled"
time=2024-06-12T01:42:20.094170 level=INFO source=/home/dev/sonic/instrumentation.go:73 msg="serving telemetry" address=:8081/metrics
```

Building
--------

For easier management a `Makefile` is included, using it, the build is as easy as:

```sh
make
```

If your operating system does not provide a usable form of `make`, you can also do:

```sh
CGO_ENABLED=0 go build -ldflags "-s -w"
```
