FROM  quay.io/wasilak/golang:1.24-alpine as builder
ARG VERSION=main

LABEL org.opencontainers.image.source="https://github.com/wasilak/cloudflare-ddns"

RUN apk add --no-cache git

WORKDIR /src

COPY ./ .

RUN go build -ldflags "-X github.com/wasilak/notes-manager/libs/common.Version=${VERSION}" -o /src/cloudflare-ddns .

FROM quay.io/wasilak/alpine:3

COPY --from=builder /src/cloudflare-ddns /bin/cloudflare-ddns

CMD ["/bin/cloudflare-ddns"]
