FROM golang:alpine as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN apk --no-cache add ca-certificates
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/main /app/
WORKDIR /app

CMD ["./main"]
