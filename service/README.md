<!-- markdownlint-disable MD013 MD033 MD041 -->
<!-- SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
     SPDX-License-Identifier: MPL-2.0
-->

# SonicWeb Service

The service library of *SonicWeb* facilitates the management of multiple HTTP server instances. Its intent is to lift
the responsibility of managing the lifetimes of these servers to the user.


## Installation

To install the *SonicWeb* service library, you can use the following command:

```sh
go get github.com/AlphaOne1/sonicweb/service
```

Versions of this library are bound to the semantic versioning of *SonicWeb*. This library is intended for public use but
be aware that breaking changes may occur between minor versions. No breaking changes will be introduced between patch
versions.


## Getting Started

The following snippet illustrates the basic usage of the *SonicWeb* service library:

```go
server := http.Server{}

// ... initializations of the server component

g, _ := service.NewGroup(service.WithServer(&server, "webserver"))

// creating a timeout context for demonstration purposes
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

_ = g.StartAll(ctx)

// ... the servers are running now for the specified 10 seconds

g.WaitAllServersShutdown()
```

As one can see, the details of managing a clean server startup and waiting for its correct shutdown are encapsulated in
the `Group` type.


## Managing Multiple Servers

Multiple servers can be managed by a single `Group` instance specifying multiple `service.WithServer`
or using the `service.WithServers` options.

For a small number of servers the `WithServer` option is easier to use:

```go
server0 := http.Server{}
server1 := http.Server{}

// ... initializations of the server components

g, _ := service.NewGroup(
    service.WithServer(&server0, "webserver"),
    service.WithServer(&server1, "instrumentation"))

// creating a timeout context for demonstration purposes
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

_ = g.StartAll(ctx)

// ... the servers are running now for the specified 10 seconds

g.WaitAllServersShutdown()
```

But if the numbers grow, it is recommended to use the `WithServers` option:

```go
server0 := http.Server{}

servers := []*http.Server{&server0}
names := []string{"server0"} // must contain an entry for each server

// ... initializations of the server components

g, _ := service.NewGroup(service.WithServers(servers, names))

// creating a timeout context for demonstration purposes
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

_ = g.StartAll(ctx)

// ... the servers are running now for the specified 10 seconds

g.WaitAllServersShutdown()
```

## Further Options

As *SonicWeb* service library is an integral part of program flow, it is useful to give it the possibility to output its
own log messages. The option `service.WithLogger` allows to specify a `slog.Logger`.

To limit the time the servers have to shut down, a shutdown timeout can be specified using the
`service.WithShutdownTimeout` option.

A typical server usecase, example could look like this (deliberately simplified, no error handling):

```go
server := http.Server{} // generate the webserver

// using SonicWeb instrumentation library to generate an instrumentation server
instrumentationServer, _ := instrumentation.Server(
    "",     // all addresses
    "8081", // port
    nil,    // no prometheus handler
    true,   // enables pprof
    slog.Default())

// create a service group with the servers and the logger
services, _ := service.NewGroup(
    service.WithLogger(slog.Default()),
    service.WithShutdownTimeout(5*time.Second),
    service.WithServer(&server, "webserver"),
    service.WithServer(instrumentationServer, "instrumentation"))

// telling to signalize on a context when the specified signals are received
signalShutdown, signalShutdownFunc := signal.NotifyContext(
    context.Background(),
    syscall.SIGINT,
    syscall.SIGTERM)

defer signalShutdownFunc() // should be called for cleanups

_ = services.StartAll(signalShutdown) // starts the servers and waits for a shutdown signal
services.WaitAllServersShutdown()     // waits for the servers to have shut down
```
