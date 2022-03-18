FROM tetafro/golang-gcc:1.16-alpine AS build
RUN apk add upx
WORKDIR /app/
COPY ./ /app/
RUN --mount=type=cache,target=/root/go\
    go mod download
RUN --mount=type=cache,target=/root/.cache/go-build\
    go build && upx anniv-go

FROM node:current-alpine AS frontend-build
COPY frontend/ /app/
WORKDIR /app/
RUN --mount=type=cache,target=/app/node_modules\
    yarn install && yarn build

FROM alpine:latest
WORKDIR /app
VOLUME /app/data
COPY --from=build /app/anniv-go /app/
COPY --from=frontend-build /app/dist /app/frontend
ENV GIN_MODE=release
ENTRYPOINT /app/anniv-go --conf /app/data/config.yml --db /app/data/data.db
EXPOSE 8080/tcp
