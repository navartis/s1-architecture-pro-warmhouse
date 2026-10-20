FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/device-service .

FROM alpine:3.20

RUN adduser -D -u 10001 app
USER app

COPY --from=build /out/device-service /usr/local/bin/device-service

EXPOSE 8082

CMD ["device-service"]
