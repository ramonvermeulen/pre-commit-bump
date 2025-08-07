FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pre-commit-bump .

FROM alpine:edge

WORKDIR /app

COPY --from=build /app/pre-commit-bump .

ENTRYPOINT ["/app/pre-commit-bump"]