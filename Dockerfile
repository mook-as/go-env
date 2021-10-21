FROM golang:1-alpine AS builder

WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /go-env .

RUN mkdir -p /root/etc
RUN chmod a+x /root/etc

FROM scratch

COPY --from=builder /go-env /
COPY --from=builder /root/etc /etc

ENTRYPOINT ["/go-env"]
