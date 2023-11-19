FROM golang:1.21.4-alpine AS build
ADD . /src
WORKDIR /src
RUN GOOS=linux go build -o lambda cmd/lambda/main.go

FROM alpine
WORKDIR /app
COPY --from=build /src/lambda /app
