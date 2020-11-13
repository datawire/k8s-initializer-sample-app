# This image is used to build executables and run tests
FROM golang:alpine AS builder

# Git is required for fetching the dependencies.
RUN apk add --no-cache git ca-certificates

# Download (cache) dependencies
WORKDIR $GOPATH/src/github.com/datawire/k8s-initializer-sample-app
COPY go.mod .
COPY go.sum .
RUN go mod download

# These copies are for the compile and tests
COPY main.go main.go

# These copies are for the second build step
COPY static /sample-app/static
COPY templates /sample-app/templates

# Build the static binaries.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/sample-app

FROM scratch
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /sample-app/static /go/workdir/static
COPY --from=builder /sample-app/templates /go/workdir/templates
COPY --from=builder /go/bin/sample-app /go/bin/sample-app
WORKDIR /go/workdir
ENTRYPOINT ["/go/bin/sample-app"]
