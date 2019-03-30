FROM golang:1.12-alpine as builder

RUN apk add --no-cache git

WORKDIR /project
COPY . .

RUN go build

FROM alpine

COPY --from=builder /project/docker-performance-snapshot /usr/bin/docker-performance-snapshot

CMD ["/usr/bin/docker-performance-snapshot"]