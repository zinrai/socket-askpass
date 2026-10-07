#!/usr/bin/env python3
"""Runs ssh on a jump host to reach a destination, with socket-askpass on the
jump host answering its password prompt. The secret is served from this
machine, by ASKPASS_SECRET_COMMAND, only while this script runs.

See README.md in this directory for what it needs."""

import os
import secrets
import shlex
import signal
import socket
import subprocess
import sys
import tempfile
import threading


def serve_secret(listener, secret_command):
    while True:
        try:
            conn, _ = listener.accept()
        except OSError:
            return
        with conn:
            # Not the terminal: ssh reads from it at the same time
            result = subprocess.run(
                secret_command,
                shell=True,
                stdin=subprocess.DEVNULL,
                stdout=subprocess.PIPE,
            )
            # Not sent when the command fails: part of a secret, or a message,
            # would be taken as the password
            if result.returncode == 0:
                conn.sendall(result.stdout)


def command_on_jump_host(remote_socket, ssh_args):
    # Not sent with SetEnv: the jump host's sshd would have to accept them
    env = (
        f"SOCKET_ASKPASS_PATH={shlex.quote(remote_socket)}"
        " SSH_ASKPASS=socket-askpass SSH_ASKPASS_REQUIRE=force"
    )
    # Not asked to confirm a new host key: the helper would answer with the
    # secret, and ssh would keep asking
    target_options = ["-o", "StrictHostKeyChecking=accept-new"]
    # Not passed as separate arguments: ssh joins them into one line for the
    # remote shell, which would split them again
    ssh = shlex.join(["ssh", *target_options, *ssh_args])
    # Not left to sshd: it keeps the socket file after the connection ends
    cleanup = f"status=$?; rm -f {shlex.quote(remote_socket)}; exit $status"
    return f"{env} {ssh}; {cleanup}"


def exit_on_signal(signum, _frame):
    sys.exit(128 + signum)


def ssh_to_jump_host(jump, ssh_args, local_socket):
    remote_socket = f"/tmp/socket-askpass-{secrets.token_hex(8)}.sock"
    # Not only warned of: a failed forward would surface later as a rejected
    # password
    jump_options = ["-o", "ExitOnForwardFailure=yes"]
    jump_options += ["-R", f"{remote_socket}:{local_socket}"]
    if sys.stdin.isatty():
        jump_options.append("-t")
    remote = command_on_jump_host(remote_socket, ssh_args)
    status = subprocess.run(["ssh", *jump_options, jump, remote]).returncode
    return status if status >= 0 else 128 - status


def main():
    secret_command = os.environ.get("ASKPASS_SECRET_COMMAND")
    if len(sys.argv) < 3 or not secret_command:
        print(
            f"usage: ASKPASS_SECRET_COMMAND=<command> {sys.argv[0]}"
            " <jump> <ssh argument>...",
            file=sys.stderr,
        )
        return 2
    jump, ssh_args = sys.argv[1], sys.argv[2:]

    # Not left at the default: SIGHUP and SIGTERM would end the script without
    # removing the local socket
    signal.signal(signal.SIGHUP, exit_on_signal)
    signal.signal(signal.SIGTERM, exit_on_signal)

    listener = socket.socket(socket.AF_UNIX)
    with tempfile.TemporaryDirectory() as local_dir, listener:
        local_socket = os.path.join(local_dir, "s")
        listener.bind(local_socket)
        listener.listen()
        threading.Thread(
            target=serve_secret, args=(listener, secret_command), daemon=True
        ).start()
        try:
            return ssh_to_jump_host(jump, ssh_args, local_socket)
        except KeyboardInterrupt:
            return 130


if __name__ == "__main__":
    sys.exit(main())
