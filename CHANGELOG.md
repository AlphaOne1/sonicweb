Release 1.5.0
=============

- added `-clientca` option to enable mTLS connections

Release 1.4.1
=============

- update dependencies
- build using Go 1.24.3, that also fixes a vulnerability in os.Root

Release 1.4.0
=============

- added `-tryfile` option, to replace original URI with a given list of
  files to try first
- added single-page application example to illustrate the try-file usage
- added `-tlscert` and `-tlskey` parameter, enabling serving via TLS
- fixed manpage headers
- add support for automatic certificates via Let's Encrypt
  via `-acmedomain`, `-certcache` and `-acmeendpoint` parameters

Release 1.3.0
=============

- additional-header parameter `-header`
- additional-header-from-file parameter `-headerFile`
- parameter `-wafcfg` to add configurations to the Coraza Web Application Firewall

Release 1.2.1
=============

- use AlphaOne1/geany for logo output
- dependency updates
- source location logging only for DEBUG log level
- removed unused utility functions

Release 1.2.0
=============

- update to Go 1.24
- use os.Root for file access protection
- update dependencies

Release 1.1.0
=============

- add Dockerfile
- added packaging for deb and rpm
- add helm chart
- add opentelemetry tracing support
- added english, german and spanish manpages

Release 1.0.0
=============

Initial release

- command line configurable webserver
- access logging
- correlation id
- coraza waf middleware
