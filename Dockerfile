FROM golang:1.20-alpine AS build
RUN apk add --no-cache gcc g++ upx git
WORKDIR /app/
COPY ./ /app/
RUN --mount=type=cache,target=/go\
    --mount=type=cache,target=/root/.cache/go-build\
    go mod download && go build && upx anniv-go

FROM node:current-alpine AS frontend-build
COPY frontend/ /app/
WORKDIR /app/
RUN --mount=type=cache,target=/app/node_modules\
    yarn install && yarn build

FROM alpine:latest
WORKDIR /app
VOLUME /app/data
VOLUME /app/tmp
COPY --from=ghcr.io/projectanni/annil:latest /app/anni /usr/bin
COPY --from=build /app/anniv-go /app/
COPY --from=frontend-build /app/dist /app/frontend
ENV GIN_MODE=release DB_VENDOR=sqlite DB_PATH=/app/data/data.db CONF=/app/data/config.yml
ENTRYPOINT /app/anniv-go
EXPOSE 8080/tcp
