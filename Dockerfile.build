FROM golang:1.17

ENV OS linux
ENV PKGPATH github.com/lopezator/migrator
ENV CGO_ENABLED 0

# golangci-lint
RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $GOPATH/bin v1.44.2

# copy current workspace
WORKDIR ${GOPATH}/src/${PKGPATH}
COPY . ${GOPATH}/src/${PKGPATH}