FROM golang:1.11.5-stretch

ENV OS linux
ENV GO111MODULE on
ENV PKGPATH github.com/lopezator/migrator
ENV CGO_ENABLED 0

# golangci-lint
RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $GOPATH/bin v1.15.0

# copy current workspace
WORKDIR ${GOPATH}/src/${PKGPATH}
COPY . ${GOPATH}/src/${PKGPATH}