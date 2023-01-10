FROM ghcr.io/void-linux/void-linux:20220530rc01-mini-x86_64-musl AS prep
RUN xbps-install -MSuy xbps
RUN xbps-install -MSuy go

FROM prep AS build
COPY . .
RUN go build

FROM scratch as available
COPY --from=build /mdbook-d2-go /
