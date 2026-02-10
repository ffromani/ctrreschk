FROM docker.io/golang:1.25 AS builder

WORKDIR /go/src/github.com/ffromani/ctrreschk
COPY . .

RUN make build

FROM scratch
COPY --from=builder /go/src/github.com/ffromani/ctrreschk/_out /
ENTRYPOINT ["/ctrreschk", "-w", "align"]
