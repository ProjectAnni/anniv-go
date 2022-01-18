FROM tetafro/golang-gcc:1.16-alpine AS build
RUN apk add upx
COPY go.mod go.sum /app/
WORKDIR /app/
ENV GOPROXY=https://goproxy.io,direct
RUN go mod download
COPY ./ /app/
RUN --mount=type=cache,target=/root/.cache/go-build\
     go build ./meta/cmd/ && upx cmd

FROM alpine:latest
WORKDIR /app
VOLUME /app/repo
COPY --from=build /app/cmd /app/
ENTRYPOINT /app/cmd /app/repo
