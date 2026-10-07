package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func copySecret(w io.Writer, path string, timeout time.Duration) error {
	c, err := net.Dial("unix", path)
	if err != nil {
		return fmt.Errorf("cannot reach %s: %w", path, err)
	}
	defer c.Close()

	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	// Not io.Copy: a secret cut short by the deadline would already be on
	// stdout, and a caller that ignores the exit status would send it
	secret, err := io.ReadAll(c)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	// Not passed on as an empty secret: a forwarded socket with nothing
	// listening behind it is accepted and then closed without a byte
	if len(secret) == 0 {
		return fmt.Errorf("cannot read %s: nothing was sent", path)
	}

	_, err = w.Write(secret)
	return err
}
