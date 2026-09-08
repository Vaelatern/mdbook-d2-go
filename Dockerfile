# syntax=docker/dockerfile:1
FROM ghcr.io/void-linux/void-musl:20260901R1 AS prep
RUN xbps-install -MSuy xbps
RUN xbps-install -MSuy go

FROM prep AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /mdbook-d2-go

FROM alpine:3.24 AS test
RUN apk add --no-cache curl tar
RUN curl -sSL https://github.com/rust-lang/mdBook/releases/download/v0.5.4/mdbook-v0.5.4-x86_64-unknown-linux-musl.tar.gz | tar -xz -C /usr/local/bin
COPY --from=build /mdbook-d2-go /usr/local/bin/
COPY example-with-d2 /book
WORKDIR /book
RUN mdbook build

FROM scratch AS available
COPY --from=build /mdbook-d2-go /
