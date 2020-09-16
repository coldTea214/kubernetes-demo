FROM golang:1.13 as builder
WORKDIR /app
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -o scheduler-extender

FROM alpine:3
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/scheduler-extender /scheduler-extender
CMD ["/scheduler-extender"]
