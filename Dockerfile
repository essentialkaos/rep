## CREATEREPO_C BUILDER ########################################################

FROM essentialkaos/alpine:3.14 as cr-builder

ARG VERSION=0.15.11

# hadolint ignore=DL3003,DL3018
RUN apk add --no-cache bash-completion bzip2-dev cmake curl-dev expat-dev \
        file-dev gcc git glib-dev libxml2-dev make musl-dev openssl-dev \
        python3-dev rpm-dev scanelf sqlite-dev upx xz-dev zlib-dev && \
    git clone --depth=1 --branch="$VERSION" \
        https://github.com/rpm-software-management/createrepo_c.git createrepo_c && \
    mkdir createrepo_c/build && cd createrepo_c/build && \
    cmake .. -DWITH_ZCHUNK=NO -DWITH_LIBMODULEMD=NO -DENABLE_DRPM=OFF \
             -DBUILD_LIBCREATEREPO_C_SHARED=OFF && \
    make && upx src/createrepo_c

## GO BUILDER ##################################################################

FROM golang:alpine as go-builder

WORKDIR /go/src/github.com/essentialkaos/rep

COPY . .

# hadolint ignore=DL3018
RUN apk add --no-cache git make gcc musl-dev upx && \
    make deps && make all && upx rep

## FINAL IMAGE #################################################################

FROM essentialkaos/alpine:3.14

LABEL org.opencontainers.image.title="rep" \
      org.opencontainers.image.description="YUM repository management utility" \
      org.opencontainers.image.vendor="ESSENTIAL KAOS" \
      org.opencontainers.image.authors="Anton Novojilov" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.url="https://kaos.sh/rep" \
      org.opencontainers.image.source="https://github.com/essentialkaos/rep"

COPY --from=go-builder /go/src/github.com/essentialkaos/rep/rep /usr/bin/
COPY --from=cr-builder /createrepo_c/build/src/createrepo_c /usr/bin/

COPY common/rep-docker.knf /etc/rep.knf

# hadolint ignore=DL3018
RUN ln -sf /rep/conf /etc/rep.d && \
    apk add --no-cache curl glib libxml2 rpm sqlite zlib

VOLUME /rep
VOLUME /input

WORKDIR /input

ENTRYPOINT ["rep"]

################################################################################
