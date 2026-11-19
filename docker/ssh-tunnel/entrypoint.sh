#!/bin/sh
set -eu

# Копия нужна для корректных прав ключа при bind mount из Docker Desktop.
umask 077
cp /run/ssh/key /tmp/ssh-key
chmod 600 /tmp/ssh-key

exec ssh -F /dev/null -N -T \
  -i /tmp/ssh-key \
  -o BatchMode=yes \
  -o IdentitiesOnly=yes \
  -o StrictHostKeyChecking=yes \
  -o UserKnownHostsFile=/run/ssh/known_hosts \
  -o ExitOnForwardFailure=yes \
  -o ConnectTimeout=10 \
  -o ServerAliveInterval=15 \
  -o ServerAliveCountMax=3 \
  "$@"
