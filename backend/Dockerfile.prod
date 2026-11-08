FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/createadmin ./cmd/createadmin

FROM alpine:3.23

RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app \
    && mkdir -p /data/uploads \
    && chown app:app /data/uploads

COPY --from=build /out/api /out/createadmin /usr/local/bin/

USER app
ENV HTTP_ADDR=:8080 UPLOAD_DIR=/data/uploads GIN_MODE=release
EXPOSE 8080
CMD ["api"]
