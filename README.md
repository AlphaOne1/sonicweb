<!-- markdownlint-disable MD013 MD033 MD041 -->
<p align="center">
    <img src="sonicweb_logo.svg" width="60%" alt="Logo"><br>
    <a href="https://github.com/AlphaOne1/sonicweb/actions/workflows/test.yml"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/sonicweb/actions/workflows/test.yml/badge.svg"
             alt="Test Pipeline Result">
    </a>
    <a href="https://github.com/AlphaOne1/sonicweb/actions/workflows/github-code-scanning/codeql"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/sonicweb/actions/workflows/github-code-scanning/codeql/badge.svg"
             alt="CodeQL Pipeline Result">
    </a>
    <a href="https://github.com/AlphaOne1/sonicweb/actions/workflows/security.yml"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/sonicweb/actions/workflows/security.yml/badge.svg"
             alt="Security Pipeline Result">
    </a>
    <a href="https://goreportcard.com/report/github.com/AlphaOne1/sonicweb"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://goreportcard.com/badge/github.com/AlphaOne1/sonicweb"
             alt="Go Report Card">
    </a>
    <a href="https://app.codecov.io/gh/AlphaOne1/sonicweb"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://codecov.io/gh/AlphaOne1/sonicweb/graph/badge.svg"
             alt="Code Coverage">
    </a>
    <a href="https://coderabbit.ai"
       rel="external noopener noreferrer"
       target="_blank">
       <img src="https://img.shields.io/coderabbit/prs/github/AlphaOne1/sonicweb"
            alt="CodeRabbit Reviews">
    </a>
    <!--<a href="https://www.bestpractices.dev/projects/0000"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://www.bestpractices.dev/projects/0000/badge"
             alt="OpenSSF Best Practices">
    </a>-->
    <a href="https://scorecard.dev/viewer/?uri=github.com/AlphaOne1/sonicweb"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://api.scorecard.dev/projects/github.com/AlphaOne1/sonicweb/badge"
             alt="OpenSSF Scorecard">
    </a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicweb?ref=badge_shield&issueType=license"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicweb.svg?type=shield&issueType=license"
            alt="FOSSA License Status">
    </a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicweb?ref=badge_shield&issueType=security"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicweb.svg?type=shield&issueType=security"
             alt="FOSSA Security Status">
    </a>
</p>
<!-- markdownlint-enable MD013 MD033 MD041 -->

*SonicWeb* is a lightweight, easy-to-use web server for static content.


Features
--------

* statically linked, suitable for use in scratch containers (~13MB)
* focused purpose, thus little attack surface
* usage of OWASP [Coraza](https://github.com/corazawaf/coraza) middleware
  to follow best security practices
* HTTPS using [Let's Encrypt](https://letsencrypt.org) certificates
* easy integration in monitoring using [Prometheus](https://prometheus.io) and/or
  [Jaeger Tracing](https://jaegertracing.io)
* no complications with configuration files


Getting Started
---------------

*SonicWeb* is controlled solely by command line arguments. They are as follows:

| Parameter                    | Description                                        | Default           | Multiple |
|------------------------------|----------------------------------------------------|-------------------|----------|
| -root           \<path\>     | root directory of content                          | `/www`            |          |
| -base           \<path\>     | base path to publish the content                   | `/`               |          |
| -port           \<port\>     | port to listen on for web requests                 | `8080`            |          |
| -address        \<address\>  | address to listen on for web requests              | all               |          |
| -tlscert        \<certfile\> | TLS certificate file                               | n/a               |          |
| -tlskey         \<keyfile\>  | TLS key file                                       | n/a               |          |
| -clientca       \<cafile\>   | client certificate authority for mTLS              | n/a               | &check;  |
| -acmedomain     \<domain\>   | allowed domain for automatic certificate retrieval | n/a               | &check;  |
| -certcache      \<path\>     | directory for certificate cache                    | os temp directory |          |
| -acmeendpoint   \<url\>      | endpoint for automatic certificate retrieval       | n/a               |          |
| -header         \<header\>   | additional header                                  | n/a               | &check;  |
| -headerfile     \<file\>     | file containing additional headers                 | n/a               | &check;  |
| -tryfile        \<fileexp\>  | always try to load file expression first           | n/a               | &check;  |
| -wafcfg         \<file-glob> | configuration for Web Application Firewall         | n/a               | &check;  |
| -iport          \<port\>     | port to listen on for telemetry requests           | `8081`            |          |
| -iaddress       \<address\>  | address to listen on for telemetry requests        | all               |          |
| -telemetry      {true,false} | enable/disable telemetry support                   | `true`            |          |
| -trace-endpoint {address}    | endpoint to send trace data to                     | `""`              |          |
| -pprof          {true,false} | enable/disable pprof support                       | `false`           |          |
| -log            \<level\>    | log level (debug, info, warn, error)               | `info`            |          |
| -logstyle       \<style\>    | log style (auto, text, json)                       | `auto`            |          |
| -help                        | print the argument overview and exit               | n/a               |          |
| -version                     | print just version information and exit            | n/a               |          |

Example call, to serve the content of `testroot/` on the standard base path `/`:

```text
$ ./sonic-linux-amd64 -root testroot/
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
   |/   \    _______\ Version: v1.6.0
         \  |              of: 2025-09-18T03:28:31Z
          \ |           using: go1.25.1
           \|
time=2025-09-18T05:43:06.033688+02:00 level=INFO msg=logging level=info
time=2025-09-18T05:43:06.034142+02:00 level=INFO msg="using root directory" root=testroot/
time=2025-09-18T05:43:06.034185+02:00 level=INFO msg="using base path" path=/
time=2025-09-18T05:43:06.034267+02:00 level=INFO msg="tracing disabled"
time=2025-09-18T05:43:06.034286+02:00 level=INFO msg="registering handler for FileServer"
time=2025-09-18T05:43:06.035879+02:00 level=INFO msg="serving pprof disabled"
time=2025-09-18T05:43:06.036019+02:00 level=INFO msg="serving telemetry" address=:8081/metrics
time=2025-09-18T05:43:06.035024+02:00 level=INFO msg="starting server" address=:8080 t_init=1.6175ms
```

HTTPS
---

*SonicWeb* supports serving HTTPS via TLS. There are two options to enable HTTPS:

1. Manually provide a certificate and a key
2. Enable automatic certificate retrieval via [Let's Encrypt](https://letsencrypt.org)


### Manual Configuration

To use a certificate and key pair, you simply start *SonicWeb* as follows:

```shell
./sonic-linux-amd64 -root testroot/ -tlscert cert.pem -tlskey key.pem
```

The Makefile provides a straightforward way to generate certificates for testing purposes.
For serious use, an official certificate signed by a certificate authority should be considered.

### Manual Configuration with Client Certificate Authentication

To use the client certificate authentication, you simply start *SonicWeb* as follows:

```shell
./sonic-linux-amd64 -root testroot/ -tlscert cert.pem -tlskey key.pem -clientca clientca0.pem
```

### Automatic Certificate Retrieval

Let's Encrypt offers to automatically obtain certificates. For this to work, *SonicWeb* holds a list of valid domains,
for which certificate retrieval is allowed. When a client connects to one of these, and no certificate is available,
*SonicWeb* sends a certificate request to Let's Encrypt. The valid domains can be specified via the `-acmedomain`
parameter. Only exact domains match, so subdomains must be provided with repeated calls.

Once a certificate is obtained, it is stored in a certificate cache. By default, this cache is in the
operating system's default temporary directory. It can be changed using the `-certcache` parameter.

To start *SonicWeb* using automatic certificate retrieval, use the following command:

```shell
./sonic-linux-amd64 -root testroot/ -acmedomain example.com -acmedomain www.example.com
```

Other acme endpoints can be used, specifying the `-acmeendpoint` parameter. If nothing is specified, the production
endpoint of Let's Encrypt is used. Use the following command for testing:

```shell
./sonic-linux-amd64 -root testroot/             \
                    -acmedomain example.com     \
                    -acmedomain www.example.com \
                    -acmeendpoint "https://acme-staging-v02.api.letsencrypt.org/directory"
```

Additional Headers
------------------

In some situations, it is necessary to add HTTP headers to the response.
*SonicWeb* provides the `-header` parameter to facilitate this.

```shell
./sonic-linux-amd64 -root testroot/ -header "Environment: production"
```

To add a huge number of headers the `-headerfile` parameter can be used:

```shell
./sonic-linux-amd64 -root testroot/ -headerfile additional_headers.conf
```

The file should be formatted as follows:

```text
<HeaderKey>: <HeaderValue>
 <nextLine, if multi-line, starts with space>
```

Headers can be specified multiple times, with the last entry taking precedence.
*SonicWeb* sets the `Server` header to its name and version. By providing an own version of the `Server` header,
it can be replaced, e.g., to misguide potential attackers.


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
./sonic-linux-amd64 -root testroot/ -tryfile \$uri -tryfile /
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
./sonic-linux-amd64 -root testroot/                          \
                    -wafcfg /etc/crs4/crs-setup.conf         \
                    -wafcfg /etc/crs4/plugins/\*-config.conf \
                    -wafcfg /etc/crs4/plugins/\*-before.conf \
                    -wafcfg /etc/crs4/rules/\*.conf          \
                    -wafcfg /etc/crs4/plugins/\*-after.conf
```

Docker Usage
------------

*SonicWeb* is also distributed as a docker image. To start it, one can simply write:

```shell
docker run -p 8080:8080 ghcr.io/alphaone1/sonicweb:v1.6.0
```

and it will show this documentation. The entrypoint of the dockerfile just starts *SonicWeb* without any parameters.
So `/www` is the default web root directory. Every parameter passed after the image name is appended as a parameter
to *SonicWeb*. So running e.g.

```shell
docker run -p 8080:8080 ghcr.io/alphaone1/sonicweb:v1.6.0 --log=debug
```

is equivalent to running:

```shell
./sonic-linux-amd64 --log=debug
```

The docker image is prepared to have new web content mounted on `/www` replacing the default content entirely. A new
web root directory, e.g. `myapp/` could be mounted like this:

```shell
docker run -p 8080:8080 -v ./myapp:/www:ro ghcr.io/alphaone1/sonicweb:v1.6.0
```

Note that without specifying the `:ro` flag, the content will be mounted as read-write. *SonicWeb* does not write into
the mounted directory. Nevertheless it poses a potential risk. Also take care that the content of the mounts is
readable by the non-root-user that *SonicWeb* uses (UID 65532).

If telemetry is needed, port 8081 needs to be exposed additionally:

```shell
docker run -p 8080:8080 -p 8081:8081 ghcr.io/alphaone1/sonicweb:v1.6.0
```


Building
--------

For easier management, a `Makefile` is included, using it, the build is as easy as:

```shell
make
```

If your operating system does not provide a usable form of `make`, you can also do:

```shell
go get
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"
```
