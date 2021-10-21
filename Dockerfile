FROM golang:1-alpine AS builder

WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /go-env .

FROM scratch

COPY --from=builder /go-env /

ENTRYPOINT ["/go-env"]
