# Copyright the SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

############################
# Usage:
#
# This dockerfile relies on a previously build os and architecture fitting executable.
# It can be generated as follows:
#     $ GOOS=linux GOARCH=amd64 make
# Copy or mount the web content to the /www directory.
# After starting the image the content of this directory will be served.

LABEL org.opencontainers.image.source=https://github.com/AlphaOne1/sonicweb
LABEL org.opencontainers.image.description="SonciWeb webserver"
LABEL org.opencontainers.image.licenses=MPL-2.0

ARG USER=appuser

FROM ubuntu:latest@sha256:9cbed754112939e914291337b5e554b07ad7c392491dba6daf25eef1332a22e8 AS builder

ARG TARGETARCH
ARG USER

RUN useradd --home     "/nonexistent"  \
            --shell    "/sbin/nologin" \
            "${USER}"

RUN mkdir -p /tmp/bin /tmp/tmp /tmp/www

COPY sonic-linux-${TARGETARCH} /tmp/bin/sonicweb
RUN chmod +x /tmp/bin/sonicweb

################################################################################
FROM scratch AS sonicweb

ARG USER

COPY --from=builder                         /etc/passwd /etc/passwd
COPY --from=builder                         /etc/group  /etc/group

COPY --from=builder --chown=${USER}:${USER} /tmp/bin    /bin
COPY --from=builder --chown=${USER}:${USER} /tmp/tmp    /tmp
COPY --from=builder --chown=${USER}:${USER} /tmp/www    /www

VOLUME /www

EXPOSE 8080/tcp
EXPOSE 8081/tcp
USER ${USER}:${USER}

ENTRYPOINT ["/bin/sonicweb", "--root=/www"]