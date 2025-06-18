FROM golang:1.24

RUN apt-get update && apt-get install -y gcc

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o events-service ./cmd/main.go

# FROM alpine:latest  
# RUN apk --no-cache add ca-certificates gcc
# WORKDIR /app
# COPY --from=builder /app/events-service /app/events-service
# COPY config.yml /app/config.yml
EXPOSE 8082

CMD ["./events-service"]