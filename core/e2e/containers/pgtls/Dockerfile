# postgres:17 with TLS enabled from certs passed in via base64 env vars. Used by
# the e2e TLS test, which can't bind-mount host files (docker-out-of-docker).
FROM postgres:17
COPY entrypoint.sh /tls-entrypoint.sh
RUN chmod +x /tls-entrypoint.sh
ENTRYPOINT ["/tls-entrypoint.sh"]
