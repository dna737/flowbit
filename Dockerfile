# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS build

# Set the working directory inside the container
WORKDIR /src

# Copy dependency files first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the backend source code
COPY . ./

# Build the binary
# -trimpath and ldflags help reduce binary size and remove local file paths
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Use Google's distroless image for a tiny, secure runtime environment
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api

# Cloud Run defaults to port 8080
EXPOSE 8080

# Run as non-privileged user for security
USER nonroot:nonroot

ENTRYPOINT ["/api"]
