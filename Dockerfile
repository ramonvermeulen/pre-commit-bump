FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pre-commit-bump .

FROM alpine:edge

WORKDIR /app

COPY --from=build /app/pre-commit-bump .
COPY gha-entrypoint.sh /app/gha-entrypoint.sh

RUN chmod +x /app/gha-entrypoint.sh

ENTRYPOINT ["/app/gha-entrypoint.sh"]