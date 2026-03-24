FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o gitea-hooks .

FROM alpine:3.21
RUN apk add --no-cache git openssh-client nodejs npm
RUN npm install -g @anthropic-ai/claude-code
WORKDIR /app
COPY --from=builder /app/gitea-hooks .
RUN mkdir -p /data/reviews
EXPOSE 8080
CMD ["./gitea-hooks"]
