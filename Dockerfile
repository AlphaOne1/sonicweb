# Copyright the SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

############################
# Usage:
#
# This dockerfile relies on a previously build linux executable, it can be generated as follows:
#     $ GOOS=linux make
# Copy the web content to the /www directory.
# After starting the image the content of this directory will be served.

FROM ubuntu:latest AS builder

ENV USER=appuser

RUN useradd --home     "/nonexistent"  \
            --shell    "/sbin/nologin" \
            "${USER}"

RUN mkdir -p /tmp/bin /tmp/tmp /tmp/www

COPY sonic-*-* .
RUN mv sonic-linux-* /tmp/bin/sonicweb
RUN chmod +x /tmp/bin/sonicweb

################################################################################
FROM scratch AS sonicweb

COPY --from=builder                         /etc/passwd /etc/passwd
COPY --from=builder                         /etc/group  /etc/group

COPY --from=builder --chown=appuser:appuser /tmp/bin    /bin
COPY --from=builder --chown=appuser:appuser /tmp/tmp    /tmp
COPY --from=builder --chown=appuser:appuser /tmp/www    /www

EXPOSE 8080
EXPOSE 8081
USER appuser:appuser

ENTRYPOINT ["/bin/sonicweb", "--root=/www"]