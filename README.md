<!-- markdownlint-disable MD013 MD033 MD041 -->
<!-- SPDX-FileCopyrightText: 2026 The SonicRed contributors.
     SPDX-License-Identifier: MPL-2.0
-->
<p align="center">
    <img src="sonicred_logo.svg" width="60%" alt="Logo"><br>
    <a href="https://github.com/AlphaOne1/sonicred/blob/HEAD/go.mod"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://img.shields.io/github/go-mod/go-version/AlphaOne1/sonicred"
             alt="Go Version">
    </a>
    <a href="https://github.com/AlphaOne1/sonicred/releases"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://img.shields.io/github/v/release/AlphaOne1/sonicred"
             alt="Latest Release">
    </a>
    <a href="https://github.com/AlphaOne1/sonicred/actions/workflows/test.yml"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/sonicred/actions/workflows/test.yml/badge.svg"
             alt="Test Pipeline Result">
    </a>
    <a href="https://github.com/AlphaOne1/sonicred/actions/workflows/github-code-scanning/codeql"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/sonicred/actions/workflows/github-code-scanning/codeql/badge.svg"
             alt="CodeQL Pipeline Result">
    </a>
    <a href="https://github.com/AlphaOne1/sonicred/actions/workflows/security.yml"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://github.com/AlphaOne1/sonicred/actions/workflows/security.yml/badge.svg"
             alt="Security Pipeline Result">
    </a>
    <a href="https://goreportcard.com/report/github.com/AlphaOne1/sonicred"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://goreportcard.com/badge/github.com/AlphaOne1/sonicred"
             alt="Go Report Card">
    </a>
    <a href="https://app.codecov.io/gh/AlphaOne1/sonicred"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://codecov.io/gh/AlphaOne1/sonicred/graph/badge.svg"
             alt="Code Coverage">
    </a>
    <a href="https://coderabbit.ai"
       rel="external noopener noreferrer"
       target="_blank">
       <img src="https://img.shields.io/coderabbit/prs/github/AlphaOne1/sonicred"
            alt="CodeRabbit Reviews">
    </a>
    <!--<a href="https://www.bestpractices.dev/projects/0000"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://www.bestpractices.dev/projects/0000/badge"
             alt="OpenSSF Best Practices">
    </a>-->
    <a href="https://scorecard.dev/viewer/?uri=github.com/AlphaOne1/sonicred"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://api.scorecard.dev/projects/github.com/AlphaOne1/sonicred/badge"
             alt="OpenSSF Scorecard">
    </a>
    <a href="https://api.reuse.software/info/github.com/AlphaOne1/sonicred"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://api.reuse.software/badge/github.com/AlphaOne1/sonicred"
            alt="REUSE compliance">
    </a>
    <a href="https://slsa.dev"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://slsa.dev/images/gh-badge-level3.svg"
             alt="SLSA Level 3">
    </a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicred?ref=badge_shield&issueType=license"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicred.svg?type=shield&issueType=license"
            alt="FOSSA License Status">
    </a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicred?ref=badge_shield&issueType=security"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2FAlphaOne1%2Fsonicred.svg?type=shield&issueType=security"
             alt="FOSSA Security Status">
    </a>
    <a href="https://pkg.go.dev/github.com/AlphaOne1/sonicred"
       rel="external noopener noreferrer"
       target="_blank">
        <img src="https://pkg.go.dev/badge/github.com/AlphaOne1/sonicred.svg"
             alt="Go Reference">
    </a>
</p>
<!-- markdownlint-enable MD013 MD033 MD041 -->

> [!IMPORTANT]
> **Project Renaming:** This project was formerly known as **SonicWeb**. All repository paths,
> Go modules, and binaries have been updated to **SonicRed**. If you have cloned this repository
> prior to the rename, please update your remote URLs.

*SonicRed* is a lightweight, easy-to-use web server for static content.


Features
--------

* statically linked, suitable for use in scratch containers (~13MB)
* focused purpose, thus little attack surface
* integrates OWASP [Coraza](https://github.com/corazawaf/coraza) middleware
  to follow security best practices
* HTTPS using [Let's Encrypt](https://letsencrypt.org) certificates
* easy integration with monitoring and observability tools via OpenTelemetry,
  for example, [Prometheus](https://prometheus.io) and [Jaeger Tracing](https://jaegertracing.io)
* no need for complicated configuration files


Installation
------------

*SonicRed* provides prebuilt binaries and installation packages for Debian and RPM-based distributions.

Builds are secured with SLSA Level 3 provenance via slsa-framework/slsa-github-generator.
The downloaded archive together with the provenance file `multiple.intoto.jsonl`
can be verified using the [slsa-verifier](https://github.com/slsa-framework/slsa-verifier/)
(replace the `<VERSION>` with the one you actually downloaded, e.g., `v1.10.0`):

```bash
slsa-verifier verify-artifact SonicRed-linux-amd64-<VERSION>.deb \
    --provenance-path multiple.intoto.jsonl                      \
    --source-uri github.com/AlphaOne1/sonicred                   \
    --source-tag <VERSION>
```

If a source build is necessary, one could use the following command:

```bash
go install -ldflags "-s -w -X main.buildInfoTag=latest" github.com/AlphaOne1/sonicred@latest
```

Getting Started
---------------

*SonicRed* is controlled solely by command line arguments. They are as follows:

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
| -index          {true,false} | enable directory listing                           | true              |          |
| -header         \<header\>   | additional header                                  | n/a               | &check;  |
| -headerfile     \<file\>     | file containing additional headers                 | n/a               | &check;  |
| -tryfile        \<fileexp\>  | always try to load file expression first           | n/a               | &check;  |
| -wafcfg         \<file-glob> | configuration for Web Application Firewall         | n/a               | &check;  |
| -iport          \<port\>     | port to listen on for telemetry requests           | `8081`            |          |
| -iaddress       \<address\>  | address to listen on for telemetry requests        | all               |          |
| -telemetry      {true,false} | enable/disable telemetry support                   | `true`            |          |
| -trace-endpoint {address}    | deprecated, use OTEL_EXPORTER_OTLP_TRACES_ENDPOINT environment instead | `""`|    |
| -pprof          {true,false} | enable/disable pprof support                       | `false`           |          |
| -log            \<level\>    | log level (debug, info, warn, error)               | `info`            |          |
| -logstyle       \<style\>    | log style (auto, text, json)                       | `auto`            |          |
| -help                        | print the argument overview and exit               | n/a               |          |
| -version                     | print just version information and exit            | n/a               |          |

Example call, to serve the content of `testroot/` on the standard base path `/`:

```text
$ ./sonicred-linux-amd64 -root testroot/
           |\
           ||\
  _________||\\
  \            \   /|
   \     ___    \ / |
  /     /.-.\   _V__|             _      ____           __
 /     //   \  / ___/____  ____  (_)____/ __ \,__  ____/ /
/___  // _  |  \__ \/ __ \/ __ \/ / ___/ /_/ / _ \/ __  /
   |   \(_)/  ___/ / /_/ / / / / / /__/ _, _/  __/ /_/ /
   |  , \_/  /____/\____/_/ /_/_/\___/_/ |_|\___/\__._/
   | / \           \
   |/   \    _______\ Version: v1.11.0
         \  |              of: 2026-05-09T20:35:43Z
          \ |           using: go1.26.3
           \|
time=2026-05-10T23:20:45.064848+02:00 level=INFO msg=logging level=info
time=2026-05-10T23:20:45.065440+02:00 level=INFO msg="using root directory" root=testroot
time=2026-05-10T23:20:45.065505+02:00 level=INFO msg="using base path" path=/
time=2026-05-10T23:20:45.070429+02:00 level=INFO msg="telemetry initialized"
time=2026-05-10T23:20:45.070460+02:00 level=INFO msg="registering handlers for FileServer"
time=2026-05-10T23:20:45.072648+02:00 level=INFO msg="started server" address=:8080 t_init=10.757033ms
time=2026-05-10T23:20:45.072756+02:00 level=INFO msg="waiting for servers to shutdown"
time=2026-05-10T23:20:45.072990+02:00 level=INFO msg="server started" name=SonicRed addr=[::]:8080
```

HTTPS
---

*SonicRed* supports serving HTTPS via TLS. There are two options to enable HTTPS:

1. Manually provide a certificate and a key
2. Enable automatic certificate retrieval via [Let's Encrypt](https://letsencrypt.org)


### Manual Configuration

To use a certificate and key pair, you simply start *SonicRed* as follows:

```sh
./sonicred-linux-amd64 -root testroot/ -tlscert cert.pem -tlskey key.pem
```

The Makefile provides a straightforward way to generate certificates for testing purposes.
For serious use, an official certificate signed by a certificate authority should be considered.

### Manual Configuration with Client Certificate Authentication

To use the client certificate authentication, you simply start *SonicRed* as follows:

```sh
./sonicred-linux-amd64 -root testroot/ -tlscert cert.pem -tlskey key.pem -clientca clientca0.pem
```

### Automatic Certificate Retrieval

Let's Encrypt offers to automatically obtain certificates. For this to work, *SonicRed* holds a list of valid domains,
for which certificate retrieval is allowed. When a client connects to one of these, and no certificate is available,
*SonicRed* sends a certificate request to Let's Encrypt. The valid domains can be specified via the `-acmedomain`
parameter. Only exact domains match, so subdomains must be provided with repeated calls.

Once a certificate is obtained, it is stored in a certificate cache. By default, this cache is in the
operating system's default temporary directory. It can be changed using the `-certcache` parameter.

To start *SonicRed* using automatic certificate retrieval, use the following command:

```sh
./sonicred-linux-amd64 -root testroot/ -acmedomain example.com -acmedomain www.example.com
```

Other acme endpoints can be used, specifying the `-acmeendpoint` parameter. If nothing is specified, the production
endpoint of Let's Encrypt is used. Use the following command for testing:

```sh
./sonicred-linux-amd64 -root testroot/             \
                       -acmedomain example.com     \
                       -acmedomain www.example.com \
                       -acmeendpoint "https://acme-staging-v02.api.letsencrypt.org/directory"
```

Directory Listing
-----------------

When serving files from a directory, it might be useful to enable directory listing.
*SonicRed* supports this via the `-index` parameter. It is enabled by default. When enabled, *SonicRed* will
respond to directory requests with a generated page that contains the entries of the requested directory,
including:

  - file or directory name
  - size (in bytes)
  - last modified time

When disabled, attempting to list a directory's contents will result in a 403 Forbidden response.

Additional Headers
------------------

In some situations, it is necessary to add HTTP headers to the response.
*SonicRed* provides the `-header` parameter to facilitate this.

```sh
./sonicred-linux-amd64 -root testroot/ -header "Environment: production"
```

To add a huge number of headers, the `-headerfile` parameter can be used:

```sh
./sonicred-linux-amd64 -root testroot/ -headerfile additional_headers.conf
```

The file should be formatted as follows:

```text
<HeaderKey>: <HeaderValue>
 <nextLine, if multi-line, starts with space>
```

Headers can be specified multiple times, with the last entry taking precedence.
*SonicRed* sets the `Server` header to its name and version.
By providing a custom `Server` header, it can be replaced, e.g., to mislead potential attackers.


Try Files
---------

The `-tryfile` option is specially aimed at single-page applications that use URIs to encode functionality.
When used, *SonicRed* tries the given file expressions in order. There is a special value that can be used:

| Value         | Description        |
|---------------|--------------------|
| $uri          | URI of the request |

If none of the expressions matches a real file, a 404 is returned. If one of the expressions ends with `/index.html`,
that suffix is truncated—replaced by the final `/`—to prevent redirection loops caused by Go's handling of
`/index.html`. (Go’s FileHandler redirects to `/` when it encounters `/index.html`; therefore, attempting to load
`/index.html` would trigger a redirect and repeatedly try to load `/index.html` instead of `/`, resulting in a loop.)

An invocation of *SonicRed* could then be as follows:

```sh
./sonicred-linux-amd64 -root testroot/ -tryfile \$uri -tryfile /
```

Web Application Firewall
------------------------

*SonicRed* integrates the [Coraza](https://github.com/corazawaf/coraza) Web Application Firewall middleware. It uses
rules to determine actions on the incoming (and outgoing) HTTP traffic. This project does not include the rulesets.
The rules can be activated using the `-wafcfg` parameter. It expects, for each invocation, a file containing a Coraza
configuration file. A good base ruleset can be obtained from [coreruleset.org](https://coreruleset.org).
There is also extensive documentation on how to write new rules.

*SonicRed* can be started as follows:

```sh
./sonicred-linux-amd64 -root testroot/                          \
                       -wafcfg /etc/crs4/crs-setup.conf         \
                       -wafcfg /etc/crs4/plugins/\*-config.conf \
                       -wafcfg /etc/crs4/plugins/\*-before.conf \
                       -wafcfg /etc/crs4/rules/\*.conf          \
                       -wafcfg /etc/crs4/plugins/\*-after.conf
```

Docker Usage
------------

*SonicRed* is also distributed as a Docker image. To start it, one can simply write:

```sh
docker run -p 8080:8080 ghcr.io/alphaone1/sonicred:v1.10.0
```

and it will show this documentation. The entrypoint of the Dockerfile just starts *SonicRed* without any parameters.
Therefore, `/www` is the default web root directory. Every parameter passed after the image name is appended as a
parameter to *SonicRed*. For example, running

```sh
docker run -p 8080:8080 ghcr.io/alphaone1/sonicred:v1.10.0 --log=debug
```

is equivalent to running:

```sh
./sonicred-linux-amd64 --log=debug
```

The Docker image allows new web content to be mounted on `/www`, replacing the default content entirely. A new
web root directory, e.g., `myapp/`, can be mounted like this:

```sh
docker run -p 8080:8080 -v ./myapp:/www:ro ghcr.io/alphaone1/sonicred:v1.10.0
```

Note that without specifying the `:ro` flag, the content will be mounted as read-write. *SonicRed* does not write into
the mounted directory. Nevertheless, it poses a potential risk. Ensure that the mounted content is readable by the
non-root user that *SonicRed* uses (UID 65532).

If telemetry is needed, port 8081 needs to be exposed additionally:

```sh
docker run -p 8080:8080 -p 8081:8081 ghcr.io/alphaone1/sonicred:v1.10.0
```
