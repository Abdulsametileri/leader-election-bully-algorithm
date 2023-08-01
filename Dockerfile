FROM golang:alpine AS builder
# Create non-root user.
RUN adduser -D -g '' appuser
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o node .

# Final Image.
FROM alpine:latest
# Change to non-root user.
USER appuser
WORKDIR /app/
# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app/node .
ENTRYPOINT [ "/app/node" ]
# it will be overriden
CMD [ "node-01" ]