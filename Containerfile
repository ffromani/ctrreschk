FROM docker.io/golang:1.25 AS builder

WORKDIR /go/src/github.com/ffromani/ctrreschk
COPY . .

RUN make build

FROM alpine:3.23
COPY --from=builder /go/src/github.com/ffromani/ctrreschk/_out /bin
ENTRYPOINT ["/bin/ctrreschk", "align"]
