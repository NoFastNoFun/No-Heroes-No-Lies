# ── build stage ──
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY cmd ./cmd
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app /src/cmd/server

# ── run stage ──
FROM alpine:3.20
ENV PORT=8080
COPY --from=build /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
