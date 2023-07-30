FROM golang:alpine AS builder
RUN adduser -D -g '' appuser
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o node .

FROM alpine:latest
USER appuser
WORKDIR /app/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app/node .
ENTRYPOINT [ "/app/node" ]
CMD [ "1" ]