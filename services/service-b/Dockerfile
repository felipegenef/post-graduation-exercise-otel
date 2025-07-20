# Build binary
FROM golang:alpine as build

WORKDIR /build

COPY go.mod go.sum ./
COPY . .

# Add to certs to enable http requests on schratch image
RUN apk add --no-cache ca-certificates

RUN GOOS=linux go build -ldflags="-w -s" -o ./main main.go

FROM scratch


WORKDIR "/app"
COPY --from=build /build/main /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/



CMD ["/app/main"]