# Stage 1: Module Caching
FROM golang:1.24 AS modules
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: Build
FROM golang:1.24 AS builder
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./main.go

# Stage 3: Final minimal image
FROM scratch
COPY --from=builder /app/app /app/app

# ls -l /app
RUN ls -R /app
ENTRYPOINT ["/app/app"]
