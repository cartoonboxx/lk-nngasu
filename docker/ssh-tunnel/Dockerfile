FROM alpine:3.22

RUN apk add --no-cache openssh-client postgresql-client

COPY entrypoint.sh /usr/local/bin/ssh-tunnel
ENTRYPOINT ["sh", "/usr/local/bin/ssh-tunnel"]
