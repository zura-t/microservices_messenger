# STAGE 1. BUILD STAGE

FROM golang:1.23-alpine3.19 AS build

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/main ./cmd/api-gateway

# STAGE 2. FINAL STAGE

FROM scratch AS final

WORKDIR /

COPY --from=build /bin/main /main

EXPOSE 8080

ENTRYPOINT ["/main"]