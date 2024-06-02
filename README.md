SonicWeb
========

*SonicWeb* is a lightweight webserver for static content.

Usage
-----

*SonicWeb* is controlled solely by command line arguments. They are as follows:

| Paraeter                | Description                                                 |
|-------------------------|-------------------------------------------------------------|
| -root      \<path\>     | root directory of content, defaults to `/www`               |
| -base      \<path\>     | base path to publish the content, defaults to `/`           |
| -port      \<port\>     | port to listen on for web requests, defaults to `8080`      |
| -address   \<address\>  | address to listen on for web requests, defaults to all      |
| -iport     \<port\>     | port to listen on for telemetry requests, defaults to `8081`|
| -iaddress  \<address\>  | address to listen on for telemetry requests, defaults to all|
| -telemetry {true,false} | enable/disable telemetry support, defaults to `true`        |
| -pprof     {true,false} | enable/disable pprof support, defaults to `false`           |
| -log       \<level\>    | log level (debug, info, warn, error), defaults to `debug`   |
