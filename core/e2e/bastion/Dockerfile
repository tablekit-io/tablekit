# Minimal SSH bastion for e2e tunnel tests. On start it writes the public key
# from $AUTHORIZED_KEY into root's authorized_keys, generates host keys, and runs
# sshd in the foreground with TCP forwarding enabled (no passwords, key only).
FROM alpine:3.20

RUN apk add --no-cache openssh-server

# Key-only root login; allow the local port forwarding the tunnel relies on.
RUN sed -ri \
    -e 's/^#?PermitRootLogin.*/PermitRootLogin prohibit-password/' \
    -e 's/^#?PasswordAuthentication.*/PasswordAuthentication no/' \
    -e 's/^#?AllowTcpForwarding.*/AllowTcpForwarding yes/' \
    /etc/ssh/sshd_config

# Entrypoint: install the authorized key, ensure host keys, then exec sshd.
RUN printf '%s\n' \
    '#!/bin/sh' \
    'set -e' \
    'mkdir -p /root/.ssh' \
    'chmod 700 /root/.ssh' \
    'printf "%s\n" "$AUTHORIZED_KEY" > /root/.ssh/authorized_keys' \
    'chmod 600 /root/.ssh/authorized_keys' \
    'ssh-keygen -A' \
    'exec /usr/sbin/sshd -D -e' \
    > /entrypoint.sh && chmod +x /entrypoint.sh

EXPOSE 22
ENTRYPOINT ["/entrypoint.sh"]
