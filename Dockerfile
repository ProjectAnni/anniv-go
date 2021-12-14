FROM tetafro/golang-gcc:1.16-alpine AS build
COPY go.mod go.sum /app/
WORKDIR /app/
ENV GOPROXY=https://goproxy.io,direct
RUN go mod download
COPY ./ /app/
RUN --mount=type=cache,target=/root/.cache/go-build\
     go build

FROM alpine:latest
WORKDIR /app
VOLUME /app/data
COPY --from=build /app/anniv-go /app/
ENV GIN_MODE=release
ENTRYPOINT /app/anniv-go --conf /app/data/config.yml --db /app/data/data.db
EXPOSE 8080/tcp
