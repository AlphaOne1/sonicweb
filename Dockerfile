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

ARG USER=appuser

FROM ubuntu:latest@sha256:9cbed754112939e914291337b5e554b07ad7c392491dba6daf25eef1332a22e8 AS builder

ARG TARGETARCH
ARG USER

RUN useradd --home     "/nonexistent"      \
            --shell    "/usr/sbin/nologin" \
            --user-group                   \
            --uid 65532                    \
            --gid 65532                    \
            -r                             \
            "${USER}"

RUN mkdir -p /tmp/bin /tmp/tmp /tmp/www
RUN chmod 1777 /tmp/tmp

COPY --chmod=0755 sonic-linux-${TARGETARCH} /tmp/bin/sonicweb
RUN getent passwd "${USER}" > /tmp/passwd
RUN getent group  "${USER}" > /tmp/group

################################################################################
FROM scratch AS sonicweb

LABEL org.opencontainers.image.source=https://github.com/AlphaOne1/sonicweb \
      org.opencontainers.image.title="SonicWeb"                             \
      org.opencontainers.image.description="SonicWeb web server"            \
      org.opencontainers.image.licenses=MPL-2.0

ARG USER

COPY --from=builder                         /tmp/passwd \
                                            /tmp/group  /etc/

COPY --from=builder --chown=${USER}:${USER} /tmp/bin    \
                                            /tmp/tmp    \
                                            /tmp/www    /
VOLUME  /www
ENV     HOME=/www
WORKDIR /www

EXPOSE 8080/tcp \
       8081/tcp

USER ${USER}:${USER}

ENTRYPOINT ["/bin/sonicweb", "--root=/www"]