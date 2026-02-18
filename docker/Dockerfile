FROM golang:1.24.3-alpine AS build

WORKDIR /app
COPY .. ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/app ./cmd/app

FROM scratch

WORKDIR /app

COPY --from=build /app/bin/app ./bin/
EXPOSE 8080
ENTRYPOINT ["./bin/app"]
