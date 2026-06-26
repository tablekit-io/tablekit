FROM golang:1.25

WORKDIR /app

# Air for live-reload (project moved from cosmtrek/air to air-verse/air)
RUN go install github.com/air-verse/air@latest

# Resolve deps at container start (bind-mount shadows build-time /app),
# then run Air which rebuilds on source changes.
CMD ["sh", "-c", "go mod tidy && air -c .air.toml"]
