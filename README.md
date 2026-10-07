# socket-askpass

An askpass helper that answers a password prompt with a secret read from a Unix
socket.

`ssh` and `sudo` can run a helper to get a password, named by `SSH_ASKPASS` and
`SUDO_ASKPASS`. `socket-askpass` is such a helper. Whatever listens on the
socket decides the answer, and where the secret is kept is up to it.

[examples](examples/README.md) shows one use: answering a prompt on a jump host
with a password kept on your machine.

## Background

A password handed to a command through an environment variable is readable in
`/proc/<pid>/environ` for as long as the process lives. Written to a file, it
stays on the disk. Read from a socket, it is handed over when it is asked for,
and stored nowhere.

## Usage

```
SOCKET_ASKPASS_PATH=<socket> socket-askpass [prompt]
```

`socket-askpass` connects to the socket, reads until the other end closes the
connection, and writes what it read to stdout unchanged. The prompt argument is
ignored; it exists because `ssh` and `sudo` pass one.

It writes nothing and exits non-zero when the socket cannot be reached, sends
nothing, or does not close within a minute.

## Limitations

- Every prompt gets the same answer, including one that asks for a
  confirmation rather than a password
- Anyone who can connect to the socket can read the secret. Keep it where only
  you can reach it

## License

This project is licensed under the [MIT License](./LICENSE).
