FROM golang:1.8

ADD Makefile /go/Makefile
ADD hack /go/hack
ADD vendor /go/vendor

ADD src /go/src

RUN make build && \
      mv bin/linux/cli/client /usr/local/bin/cli && \
      mv bin/linux/server/provisioner /usr/local/bin/provisioner && \
      mv bin/linux/server/status /usr/local/bin/status && \
      chmod a+x /usr/local/bin/cli && \
      chmod a+x /usr/local/bin/provisioner && \
      chmod a+x /usr/local/bin/status

# Cleanup.
RUN rm -fR /go

CMD ["cli"]
