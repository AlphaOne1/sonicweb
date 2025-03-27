SonicWeb
========

*SonicWeb* is a lightweight, easy-to-use web server for static content.

Features
--------

* statically linked, suitable for use in scratch containers (~13MB)
* focused purpose, thus little attack surface
* usage of OWASP [Coraza](https://github.com/corazawaf/coraza) middleware
  to follow best security practises
* easy integration in monitoring using [Prometheus](prometheus.io) and/or
  [Jaeger Tracing](jaegertracing.io)
* no complications with configuration files

Getting Started
---------------

*SonicWeb* is controlled solely by command line arguments. They are as follows:

| Parameter                    | Description                                        | Default |
|------------------------------|----------------------------------------------------|---------|
| -root           \<path\>     | root directory of content                          | `/www`  |
| -base           \<path\>     | base path to publish the content                   | `/`     |
| -port           \<port\>     | port to listen on for web requests                 | `8080`  |
| -address        \<address\>  | address to listen on for web requests              | all     |
| -tlscert        \<certfile\> | TLS certificate file                               | n/a     |
| -tlskey         \<keyfile\>  | TLS key file                                       | n/a     |
| -acmedomain     \<domain\>   | allowed domain for automatic certificate retrieval | n/a     |
| -header         \<header\>   | additional header                                  | n/a     |
| -headerfile     \<file\>     | file containing additional headers                 | n/a     |
| -tryfile        \<fileexp\>  | always try to load file expression first           | n/a     |
| -wafcfg         \<file-glob> | configuration for Web Application Firewall         | n/a     |
| -iport          \<port\>     | port to listen on for telemetry requests           | `8081`  |
| -iaddress       \<address\>  | address to listen on for telemetry requests        | all     |
| -telemetry      {true,false} | enable/disable telemetry support                   | `true`  |
| -trace-endpoint {address}    | endpoint to send trace data to                     | `""`    |
| -pprof          {true,false} | enable/disable pprof support                       | `false` |
| -log            \<level\>    | log level (debug, info, warn, error)               | `info`  |
| -logstyle       \<style\>    | log style (auto, text, json)                       | `auto`  |
| -help                        | print the argument overview and exit               | n/a     |
| -version                     | print just version information and exit            | n/a     |

Example call, to serve the content of `testroot/` on the standard base path `/`:

```text
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
   |/   \    _______\ Version: v1.3.0
         \  |              of: 2025-03-24T10:34:12Z
          \ |           using: go1.24.1
           \|
time=2025-03-24T11:53:38.789677 level=INFO msg="maxprocs: Leaving GOMAXPROCS=4: CPU quota undefined"
time=2025-03-24T11:53:38.790198 level=INFO msg=logging level=info
time=2025-03-24T11:53:38.790235 level=INFO msg="using root directory" root=testroot/
time=2025-03-24T11:53:38.790268 level=INFO msg="using base path" path=/
time=2025-03-24T11:53:38.790285 level=INFO msg="tracing disabled"
time=2025-03-24T11:53:38.790815 level=INFO msg="registering handler for FileServer"
time=2025-03-24T11:53:38.792743 level=INFO msg="serving pprof disabled"
time=2025-03-24T11:53:38.793455 level=INFO msg="serving telemetry" address=:8081/metrics
time=2025-03-24T11:53:38.795349 level=INFO msg="starting server" address=:8080 t_init=6.141082ms
```

Additional Headers
------------------

In some situations, it is necessary to add HTTP headers to the response.
*SonicWeb* provides the `-header` parameter to facilitate this.

```shell
$ ./sonic-linux-amd64 --root testroot/ -header "Environment: production"
```

To add a huge amount of headers the `-headerfile` parameter can be used:

```shell
$ ./sonic-linux-amd64 --root testroot/ -headerfile additional_headers.conf
```

The file should be formatted as follows:

```text
<HeaderKey>: <HeaderValue>
 <nextLine, if multi-line, starts with space>
```

Headers can be specified multiple times, with the last entry taking precedence.
*SonicWeb* sets the `Server` header to its name and version. By providing an own version of the `Server` header,
it can be replaced, e.g. to misguide potential attackers.

Try Files
---------

The `-tryfile` option is specially aimed at single-page applications that use URIs to encode functionality.
When used, *SonicWeb* tries the given file expressions in order. There is a special value that can be used:

| Value         | Description        |
|---------------|--------------------|
| $uri          | URI of the request |

If none of the expressions matches a real file, a 404 is returned. If one of the expressions ends with `/index.html`,
that suffix is truncated—replaced by the final `/`—to prevent redirection loops caused by Go's handling of
`/index.html`. (Go’s FileHandler redirects to `/` when it encounters `/index.html`; therefore, attempting to load
`/index.html` would trigger a redirect and repeatedly try to load `/index.html` instead of `/`, resulting in a loop.)

An invocation of *SonicWeb* could then be as follows:

```shell
$ ./sonic-linux-amd64 --root testroot/ -tryfile \$uri -tryfile /
```

Web Application Firewall
------------------------

*SonicWeb* integrates the [Coraza](https://github.com/corazawaf/coraza) Web Application Firewall middleware. It uses
rules to determine actions on the incoming (and outgoing) HTTP traffic. This project does not include the rulesets.
The rules can be activated using the `-wafcfg` parameter. It expects, for each invocation, a file containing a Coraza
configuration file. A good base ruleset can be obtained from [coreruleset.org](https://coreruleset.org).
There is also extensive documentation on how to write new rules.

*SonicWeb* can be started as follows:

```shell
$ ./sonic-linux-amd64 -root testroot/                          \
                      -wafcfg /etc/crs4/crs-setup.conf         \
                      -wafcfg /etc/crs4/plugins/\*-config.conf \
                      -wafcfg /etc/crs4/plugins/\*-before.conf \
                      -wafcfg /etc/crs4/rules/\*.conf          \
                      -wafcfg /etc/crs4/plugins/\*-after.conf
```

Building
--------

For easier management, a `Makefile` is included, using it, the build is as easy as:

```sh
make
```

If your operating system does not provide a usable form of `make`, you can also do:

```sh
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"
```
