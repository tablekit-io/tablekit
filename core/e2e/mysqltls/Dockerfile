# mysql:8.4 with TLS required, certs passed in via base64 env vars. Used by the
# e2e TLS test, which can't bind-mount host files (docker-out-of-docker).
FROM mysql:8.4
COPY entrypoint.sh /tls-entrypoint.sh
RUN chmod +x /tls-entrypoint.sh
ENTRYPOINT ["/tls-entrypoint.sh"]
