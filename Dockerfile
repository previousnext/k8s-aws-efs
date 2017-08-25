FROM golang:1.8
ADD workspace /go
RUN go get github.com/mitchellh/gox
RUN make build

FROM alpine:3.6
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/k8s-aws-efs_linux_amd64 /usr/local/bin/k8s-aws-efs
CMD ["k8s-aws-efs"]
