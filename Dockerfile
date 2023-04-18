# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.20-alpine AS build

ARG NAME
ARG VERSION
ARG REVISION

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build \
	-ldflags="\
	-X 'github.com/steadybit/extension-kit/extbuild.ExtensionName=${NAME}' \
	-X 'github.com/steadybit/extension-kit/extbuild.Version=${VERSION}' \
	-X 'github.com/steadybit/extension-kit/extbuild.Revision=${REVISION}'" \
	-o /extension-postman

##
## Runtime
##
FROM node:16-alpine

ENV LC_ALL="en_US.UTF-8" LANG="en_US.UTF-8" LANGUAGE="en_US.UTF-8" ALPINE_NODE_REPO="oznu/alpine-node"

RUN npm install -g newman@5.3.2 newman-reporter-json-summary@1.0.14 newman-reporter-htmlextra@1.22.11

ARG USERNAME=steadybit
ARG USER_UID=1000

RUN adduser -u $USER_UID -D $USERNAME

USER $USERNAME

WORKDIR /

COPY --from=build /extension-postman /extension-postman

EXPOSE 8086

ENTRYPOINT ["/extension-postman"]
