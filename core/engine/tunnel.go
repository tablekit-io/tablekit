package engine

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// sshDialTimeout bounds the bastion handshake.
const sshDialTimeout = 10 * time.Second

// openTunnel dials the SSH bastion and starts a local TCP listener that forwards
// every accepted connection through the bastion to target (the real database
// host:port). It returns the local address to point the driver at and a cleanup
// that tears down the listener and SSH client. This one mechanism serves every
// engine, so no driver-specific dial hook or global registration is needed.
//
// Host-key verification is intentionally skipped for v1.
func openTunnel(cfg *sshSettings, target string) (localAddr string, cleanup func(), err error) {
	keyPEM, err := os.ReadFile(cfg.sshKeyFilePath)
	if err != nil {
		return "", nil, fmt.Errorf("read ssh key %q: %w", cfg.sshKeyFilePath, err)
	}

	var signer ssh.Signer
	if cfg.passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyPEM, []byte(cfg.passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyPEM)
	}
	if err != nil {
		return "", nil, fmt.Errorf("parse ssh key %q: %w", cfg.sshKeyFilePath, err)
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(cfg.host, strconv.Itoa(cfg.port)), &ssh.ClientConfig{
		User:            cfg.username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         sshDialTimeout,
	})
	if err != nil {
		return "", nil, fmt.Errorf("ssh dial %q: %w", cfg.host, err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		client.Close()
		return "", nil, fmt.Errorf("open local tunnel listener: %w", err)
	}

	go func() {
		for {
			local, acceptErr := listener.Accept()
			if acceptErr != nil {
				return // listener closed by cleanup
			}
			go forwardThroughTunnel(local, client, target)
		}
	}()

	cleanup = func() {
		listener.Close()
		client.Close()
	}
	return listener.Addr().String(), cleanup, nil
}

// forwardThroughTunnel pipes a local connection to target over the SSH client.
func forwardThroughTunnel(local net.Conn, client *ssh.Client, target string) {
	remote, err := client.Dial("tcp", target)
	if err != nil {
		local.Close()
		return
	}
	go func() {
		io.Copy(remote, local)
		remote.Close()
	}()
	io.Copy(local, remote)
	local.Close()
}
