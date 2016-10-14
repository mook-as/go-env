FROM alpine

ADD bin/go-env /

ENTRYPOINT ["/go-env"]

