# ── build stage ─────────────────────────────────────────────────
FROM golang:1.22-alpine AS build
WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o chaincode ./...

# ── runtime stage ───────────────────────────────────────────────
FROM alpine:3.20
WORKDIR /app
COPY --from=build /workspace/chaincode .
USER 1000
CMD ["./chaincode"]
