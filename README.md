SonicWeb
========

*SonicWeb* is a lightweight, easy-to-use web server for static content.

Features
--------

* statically linked, apt to be used in scratch containers (~19MB)
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
|                         | defaults to `debug`                          |
| -logstyle  \<style\>    | log style (auto, text, json),                |
|                         | defaults to `auto`                           |
