# syntax=docker/dockerfile:1

##
## Build
##
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

ARG TARGETOS
ARG TARGETARCH
ARG NAME
ARG VERSION
ARG REVISION
ARG ADDITIONAL_BUILD_PARAMS
ARG SKIP_LICENSES_REPORT=false
ARG VERSION=unknown
ARG REVISION=unknown

WORKDIR /app

RUN apk add build-base
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="\
    -X 'github.com/steadybit/extension-kit/extbuild.ExtensionName=${NAME}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Version=${VERSION}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Revision=${REVISION}'" \
    -o ./extension \
    ${ADDITIONAL_BUILD_PARAMS}
RUN make licenses-report

##
## Runtime
##
FROM node:26-alpine

ARG VERSION=unknown
ARG REVISION=unknown

LABEL "steadybit.com.discovery-disabled"="true"
LABEL "version"="${VERSION}"
LABEL "revision"="${REVISION}"
RUN echo "$VERSION" > /version.txt && echo "$REVISION" > /revision.txt

ENV LC_ALL="en_US.UTF-8" LANG="en_US.UTF-8" LANGUAGE="en_US.UTF-8" ALPINE_NODE_REPO="oznu/alpine-node"

# newman is installed as a local project (instead of -g) so that "overrides" can force patched transitive dependencies
COPY <<EOF /opt/newman/package.json
{
  "private": true,
  "dependencies": {
    "newman": "6.2.2",
    "newman-reporter-json-summary": "1.0.14",
    "newman-reporter-htmlextra": "1.23.1"
  },
  "overrides": {
    "handlebars": "4.7.10"
  }
}
EOF

# npm is only needed to install newman; removing it keeps its bundled dependencies (e.g. undici) out of the runtime image
RUN npm install -g --ignore-scripts npm@12.1.0 && \
    npm install --prefix /opt/newman --ignore-scripts --omit=dev --no-audit --no-fund && \
    npm cache clean --force && \
    rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

ENV PATH="/opt/newman/node_modules/.bin:${PATH}"

ARG USERNAME=steadybit
ARG USER_UID=10000

RUN apk update && apk upgrade --no-cache && rm -rf /var/cache/apk/* && \
    adduser -u $USER_UID -D $USERNAME

USER $USER_UID

WORKDIR /

COPY --from=build /app/extension /extension
COPY --from=build /app/licenses /licenses

EXPOSE 8086
EXPOSE 8087

ENTRYPOINT ["/extension"]
