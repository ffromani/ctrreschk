FROM docker.io/golang:1.22 AS builder

WORKDIR /go/src/github.com/ffromani/ctrreschk
COPY . .

RUN make build

FROM alpine:3.20
COPY --from=builder /go/src/github.com/ffromani/ctrreschk/_out /usr/local/bin
ENTRYPOINT ["/bin/sh"]
