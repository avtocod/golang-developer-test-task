# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.14-alpine as builder

RUN set -x \
    && mkdir /src \
    # SSL ca certificates (ca-certificates is required to call HTTPS endpoints)
    && apk add --no-cache ca-certificates \
    && update-ca-certificates

WORKDIR /src

COPY ./go.* ./

# Burn modules cache
RUN set -x \
    && go version \
    && go mod download \
    && go mod verify

COPY . /src

RUN set -x \
    && go version \
    && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tmp/app .

# Image page: <https://hub.docker.com/_/alpine>
FROM alpine:latest as runtime

RUN set -x \
    # Unprivileged user creation <https://stackoverflow.com/a/55757473/12429735RUN>
    && adduser \
        --disabled-password \
        --gecos "" \
        --home "/nonexistent" \
        --shell "/sbin/nologin" \
        --no-create-home \
        --uid "10001" \
        "appuser"

# Use an unprivileged user
USER appuser:appuser

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /tmp/app /bin/app

ENTRYPOINT ["/bin/app"]
