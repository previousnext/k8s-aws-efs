FROM golang:1.8

ADD Makefile /go/Makefile
ADD hack /go/hack
ADD vendor /go/vendor

ADD client /go/src/github.com/previousnext/client
ADD provisioner /go/src/github.com/previousnext/provisioner
ADD status /go/src/github.com/previousnext/status
ADD cli /go/src/github.com/previousnext/cli

RUN make && \
      mv bin/linux/cli/client /usr/local/bin/cli && \
      mv bin/linux/server/provisioner /usr/local/bin/provisioner && \
      mv bin/linux/server/status /usr/local/bin/status && \
      chmod a+x /usr/local/bin/cli && \
      chmod a+x /usr/local/bin/provisioner && \
      chmod a+x /usr/local/bin/status

# Cleanup.
RUN rm -fR /go

CMD ["cli"]
