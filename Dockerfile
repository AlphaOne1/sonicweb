# SPDX-FileCopyrightText: 2025 The SonicWeb contributors.
# SPDX-License-Identifier: MPL-2.0

############################
# Usage:
#
# This dockerfile relies on a previously build os and architecture fitting executable.
# It can be generated as follows:
#     $ make sonic-linux-amd64
# Copy or mount the web content to the /www directory.
# After starting the image the content of this directory will be served.

ARG USER=appuser

FROM ubuntu:latest@sha256:66460d557b25769b102175144d538d88219c077c678a49af4afca6fbfc1b5252 AS builder

ARG TARGETARCH
ARG USER

# Centralized versions and checksums for third-party web assets
ARG HLJS_VER=11.11.1
ARG MARKED_VER=16.3.0
ARG MARKED_HL_VER=2.2.2
ARG GHMD_VER=5.8.1

ARG HLJS_JS_SHA256=c4a399dd6f488bc97a3546e3476747b3e714c99c57b9473154c6fb8d259b9381
ARG MARKED_SHA256=fe19dcc22695007cccbd794f859676e9d25356d48be2fe1a158650405a34e81f
ARG MARKED_HL_SHA256=94854921cc0771c9b51277240ea326368d24ad05d334e8fdb0f896c68526f9b7
ARG GHMD_SHA256=c47f5a601c095973e19c0a7d0418d35b2b209098955d2cc4136eb274f9083cc4
ARG HLJS_CSS_SHA256=3a9a5def8b9c311e5ae43abde85c63133185eed4f0d9f67fea4b00a8308cf066

RUN useradd --home     "/nonexistent"      \
            --shell    "/usr/sbin/nologin" \
            --user-group                   \
            --uid 65532                    \
            -r                             \
            "${USER}"

RUN mkdir -p /tmp/root/bin     \
             /tmp/root/etc     \
             /tmp/root/tmp     \
             /tmp/root/www     \
             /tmp/root/www/css \
             /tmp/root/www/js

COPY --chmod=0755 sonic-linux-${TARGETARCH} /tmp/root/bin/sonicweb
COPY --chmod=0444 docker_root/              \
                  README.md                 \
                  sonicweb_logo.svg         /tmp/root/www/

ADD --chmod=0444                                                                     \
    --checksum=sha256:${HLJS_JS_SHA256}                                              \
    https://cdnjs.cloudflare.com/ajax/libs/highlight.js/${HLJS_VER}/highlight.min.js \
    /tmp/root/www/js/

ADD --chmod=0444                                                                      \
    --checksum=sha256:${MARKED_SHA256}                                                \
    https://cdnjs.cloudflare.com/ajax/libs/marked/${MARKED_VER}/lib/marked.umd.min.js \
    /tmp/root/www/js/

ADD --chmod=0444                                                                              \
    --checksum=sha256:${MARKED_HL_SHA256}                                                     \
    https://cdnjs.cloudflare.com/ajax/libs/marked-highlight/${MARKED_HL_VER}/index.umd.min.js \
    /tmp/root/www/js/marked-highlight.umd.min.js

ADD --chmod=0444                                                                                   \
    --checksum=sha256:${GHMD_SHA256}                                                               \
    https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/${GHMD_VER}/github-markdown.min.css \
    /tmp/root/www/css/

ADD --chmod=0444                                                                          \
    --checksum=sha256:${HLJS_CSS_SHA256}                                                  \
    https://cdnjs.cloudflare.com/ajax/libs/highlight.js/${HLJS_VER}/styles/github.min.css \
    /tmp/root/www/css/

RUN getent passwd "${USER}" > /tmp/root/etc/passwd &&\
    getent group  "${USER}" > /tmp/root/etc/group  &&\
                                                     \
    chown -R ${USER}:${USER}  /tmp/root/bin          \
                              /tmp/root/tmp          \
                              /tmp/root/www        &&\
    chmod 1777                /tmp/root/tmp        &&\
    sed -i '1,/<\/p>/{/<a href.*/,/<\/a>/d}' /tmp/root/www/README.md

################################################################################
FROM scratch AS sonicweb

# Defaults for local builds; CI should override these via --build-arg
ARG VERSION=dev
ARG REVISION=unknown
ARG CREATED=1970-01-01T00:00:00Z

LABEL org.opencontainers.image.title="SonicWeb"                             \
      org.opencontainers.image.description="SonicWeb web server"            \
      org.opencontainers.image.licenses=MPL-2.0                             \
      org.opencontainers.image.source=https://github.com/AlphaOne1/sonicweb \
      org.opencontainers.image.documentation=https://github.com/AlphaOne1/sonicweb \
      org.opencontainers.image.url=https://github.com/AlphaOne1/sonicweb    \
      org.opencontainers.image.version="${VERSION}"                         \
      org.opencontainers.image.revision="${REVISION}"                       \
      org.opencontainers.image.created="${CREATED}"

ARG USER

COPY --from=builder /tmp/root   /

# if no volume is mounted, a standard documentation page is shown.
# This page is overlayed by later mounts.
VOLUME  /www
ENV     HOME=/www \
        PATH=/bin \
        TMPDIR=/tmp
WORKDIR /www

EXPOSE  8080/tcp  \
        8081/tcp

USER    ${USER}:${USER}

ENTRYPOINT ["/bin/sonicweb"]