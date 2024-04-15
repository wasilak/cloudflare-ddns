FROM  quay.io/wasilak/golang:1.22-alpine as builder

LABEL org.opencontainers.image.source="https://github.com/wasilak/cloudflare-ddns"

RUN apk add --no-cache git

WORKDIR /src

COPY ./ .

RUN go build .

FROM quay.io/wasilak/alpine:3

COPY --from=builder /src/cloudflare-ddns /bin/cloudflare-ddns

CMD ["/bin/cloudflare-ddns"]
