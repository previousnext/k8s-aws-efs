FROM golang:1.8

ADD . /go
RUN make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=0 /go/bin/linux/cli/client /usr/local/bin/client
COPY --from=0 /go/bin/linux/server/provisioner /usr/local/bin/provisioner
COPY --from=0 /go/bin/linux/server/status /usr/local/bin/status

CMD ["cli"]
