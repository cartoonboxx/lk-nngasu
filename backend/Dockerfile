FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
# go.sum не хранится в архиве-каркасе: сгенерируйте его один раз (`make tidy`) и закоммитьте,
# тогда эту строку можно заменить на `go mod download`.
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/createadmin ./cmd/createadmin

FROM alpine:3.20
RUN adduser -D -u 10001 app && mkdir -p /data/uploads && chown app /data/uploads
COPY --from=build /out/api /out/createadmin /usr/local/bin/
USER app
ENV UPLOAD_DIR=/data/uploads HTTP_ADDR=:8080
EXPOSE 8080
CMD ["api"]
