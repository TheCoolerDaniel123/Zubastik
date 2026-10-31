FROM golang:1.25-alpine AS builder

WORKDIR /workspace

COPY go.work go.work.sum ./
COPY shared/go.mod ./shared/
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY user-service/go.mod user-service/go.sum ./user-service/
COPY apigateway-service/go.mod apigateway-service/go.sum ./apigateway-service/
COPY e2e/go.mod ./e2e/

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./user-service/cmd/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /server /server
EXPOSE 8081 50051
CMD ["/server"]
