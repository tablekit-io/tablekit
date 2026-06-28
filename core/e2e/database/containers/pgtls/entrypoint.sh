#!/usr/bin/env bash
# Write the server cert/key from base64 env, fix ownership/perms, then start
# postgres with TLS on. Key must be 0600 and owned by the db user.
set -e
mkdir -p /certs
printf '%s' "$TLS_CERT" | base64 -d > /certs/server.crt
printf '%s' "$TLS_KEY" | base64 -d > /certs/server.key
chown postgres:postgres /certs/server.crt /certs/server.key
chmod 600 /certs/server.key
exec docker-entrypoint.sh postgres \
    -c ssl=on \
    -c ssl_cert_file=/certs/server.crt \
    -c ssl_key_file=/certs/server.key
