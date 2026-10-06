# ---- build ----
FROM golang:1.25 AS builder
WORKDIR /src
ENV GOFLAGS=-mod=mod GOPROXY=https://goproxy.cn,direct
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/rin-server ./cmd/funauth

# ---- run ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/rin-server ./rin-server
ENV FUNAUTH_ADDR=:8080
EXPOSE 8080
CMD ["./rin-server"]
