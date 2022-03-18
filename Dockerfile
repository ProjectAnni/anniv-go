FROM tetafro/golang-gcc:1.16-alpine AS build
RUN apk add upx
COPY go.mod go.sum /app/
WORKDIR /app/
RUN go mod download
COPY ./ /app/
RUN --mount=type=cache,target=/root/.cache/go-build\
     go build && upx anniv-go

FROM node:current-alpine AS frontend-build
COPY frontend/package.json frontend/yarn.lock /app/
WORKDIR /app/
RUN yarn install
COPY frontend /app
RUN yarn build

FROM alpine:latest
WORKDIR /app
VOLUME /app/data
COPY --from=build /app/anniv-go /app/
COPY --from=frontend-build /app/dist /app/frontend
ENV GIN_MODE=release
ENTRYPOINT /app/anniv-go --conf /app/data/config.yml --db /app/data/data.db
EXPOSE 8080/tcp
