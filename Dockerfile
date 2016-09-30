FROM scratch

ADD bin/go-env /

ENTRYPOINT ["/go-env"]

