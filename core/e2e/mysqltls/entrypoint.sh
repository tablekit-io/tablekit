#!/bin/bash
# Write the server cert/key from base64 env, fix ownership/perms, then start
# mysqld with TLS required so only encrypted TCP connections are accepted.
set -e
mkdir -p /certs
printf '%s' "$TLS_CERT" | base64 -d > /certs/server.crt
printf '%s' "$TLS_KEY" | base64 -d > /certs/server.key
chown mysql:mysql /certs/server.crt /certs/server.key
chmod 600 /certs/server.key
exec docker-entrypoint.sh mysqld \
    --ssl-cert=/certs/server.crt \
    --ssl-key=/certs/server.key \
    --require_secure_transport=ON
