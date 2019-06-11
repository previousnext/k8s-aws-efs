FROM golang:1.12
ENV GO111MODULE=on
ADD . /go/src/github.com/previousnext/k8s-aws-efs
WORKDIR /go/src/github.com/previousnext/k8s-aws-efs
RUN go get github.com/mitchellh/gox
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/src/github.com/previousnext/k8s-aws-efs/bin/k8s-aws-efs_linux_amd64 /usr/local/bin/k8s-aws-efs
CMD ["k8s-aws-efs"]
