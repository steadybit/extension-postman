# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /extension-postman

##
## Runtime
##
FROM loadimpact/k6:latest

RUN  npm install -g newman-reporter-json-summary
RUN  npm install -g newman-reporter-htmlextra

WORKDIR /

COPY --from=build /extension-postman /extension-postman

EXPOSE 8086

ENTRYPOINT ["/extension-postman"]
