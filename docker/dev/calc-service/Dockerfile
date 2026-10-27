FROM golang:1.27.3-alpine3.24

ARG WGO_VERSION=v0.7.2
ARG DELVE_VERSION=v1.27.2
ARG GOLANGCI_LINT_VERSION=v2.15.0

ARG DEV_UID=1000
ARG DEV_GID=1000

RUN apk add --no-cache ca-certificates git gcc musl-dev make \
    && GOBIN=/usr/local/bin go install github.com/bokwoon95/wgo@${WGO_VERSION} \
    && GOBIN=/usr/local/bin go install github.com/go-delve/delve/cmd/dlv@${DELVE_VERSION} \
    && GOBIN=/usr/local/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} \
    && addgroup -g ${DEV_GID} developer \
    && adduser -D -u ${DEV_UID} -G developer developer \
    && mkdir -p /go/pkg/mod /home/developer/.cache/go-build \
    && chown -R developer:developer /go /home/developer

ENV GOCACHE=/home/developer/.cache/go-build

USER developer
WORKDIR /workspace/apps/bridge

CMD ["make", "debug"]
