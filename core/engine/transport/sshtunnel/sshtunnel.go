// Package sshtunnel forwards a local TCP listener through an SSH bastion to a
// database host:port. One mechanism serves every driver, so no driver-specific
// dial hook or global registration is needed.
package sshtunnel

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"core/engine/config"

	"golang.org/x/crypto/ssh"
)

// DialTimeout bounds the bastion handshake. Drivers also reuse it as their own
// connect timeout when tunnelling.
const DialTimeout = 10 * time.Second

// Open dials the SSH bastion described by cfg and starts a local TCP listener
// that forwards every accepted connection through the bastion to target (the
// real database host:port). It returns the local address to point the driver at
// and a cleanup that tears down the listener and SSH client.
//
// Host-key verification is intentionally skipped for v1.
func Open(cfg *config.SSHSettings, target string) (localAddr string, cleanup func(), err error) {
	keyPEM, err := os.ReadFile(cfg.SSHKeyFilePath)
	if err != nil {
		return "", nil, fmt.Errorf("read ssh key %q: %w", cfg.SSHKeyFilePath, err)
	}

	var signer ssh.Signer
	if cfg.Passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyPEM, []byte(cfg.Passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyPEM)
	}
	if err != nil {
		return "", nil, fmt.Errorf("parse ssh key %q: %w", cfg.SSHKeyFilePath, err)
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)), &ssh.ClientConfig{
		User:            cfg.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         DialTimeout,
	})
	if err != nil {
		return "", nil, fmt.Errorf("ssh dial %q: %w", cfg.Host, err)
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
			go forward(local, client, target)
		}
	}()

	cleanup = func() {
		listener.Close()
		client.Close()
	}
	return listener.Addr().String(), cleanup, nil
}

// forward pipes a local connection to target over the SSH client.
func forward(local net.Conn, client *ssh.Client, target string) {
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
