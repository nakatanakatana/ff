FROM golang:1.24 AS builder

WORKDIR /app/source
COPY ./ /app/source

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

RUN mkdir /app/output
RUN go build -o /app/output ./cmd/...

FROM busybox

COPY --from=builder /app/output /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/ff"]
