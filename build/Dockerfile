FROM nginx:1.19.10 AS base

ENV GOLANG_VERSION 1.16.4
ENV GOLANG_CHECKSUM 7154e88f5a8047aad4b80ebace58a059e36e7e2e4eb3b383127a28c711b4ff59
ENV GOROOT /opt/go
ENV GOPATH /build/gopath
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/go/bin:/build/gopath/bin

COPY bin/install_debian_dependencies.sh /usr/local/bin/install_debian_dependencies.sh

# Setup basic build skeleton and tools
RUN set -eux \
    export DEBIAN_FRONTEND=noninteractive; \
    chmod +x /usr/local/bin/install_debian_dependencies.sh; \
    /usr/local/bin/install_debian_dependencies.sh; \
    rm /usr/local/bin/install_debian_dependencies.sh; \
    curl -o /tmp/golang.tar.gz --retry 6 -Ls "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
    echo "${GOLANG_CHECKSUM}  /tmp/golang.tar.gz" | sha256sum -c; \
    tar -C /opt -xzf /tmp/golang.tar.gz; \
    rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/*; \
    mkdir -p /build/gopath;

# Preinstall linters
RUN set -eux \
    export DEBIAN_FRONTEND=noninteractive; \
    go get github.com/wadey/gocovmerge; \
    go get github.com/axw/gocov; \
    go get github.com/AlekSi/gocov-xml; \
    go get github.com/tebeka/go2xunit; \
    GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint; \
    go get github.com/aevea/commitsar; \
    tar --create --gzip --preserve-permissions --file /build/gopath_base.tar.gz build/gopath/bin; \
    rm -rf /build/gopath; \
    echo "#!/bin/sh\nif [ ! -d /build/gopath/bin ]; then tar --extract --gunzip --file /build/gopath_base.tar.gz --directory /; else echo 'gopath already present - not extracting base files'; fi" > /build/extract_gopath.sh; \
    chmod +x /build/extract_gopath.sh

VOLUME /build/gopath
VOLUME /build/src
WORKDIR /build/src