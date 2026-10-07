package main

import (
	"bytes"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func serve(t *testing.T, payload []byte, keepOpen bool) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "s")
	l, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			c.Write(payload)
			if keepOpen {
				<-release
			}
			c.Close()
		}
	}()
	return path
}

func TestSecretIsPassedThroughVerbatim(t *testing.T) {
	secret := []byte(" pa ss\tword\n")

	var out bytes.Buffer
	if err := copySecret(&out, serve(t, secret, false), time.Second); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), secret) {
		t.Errorf("got %q, want %q", out.Bytes(), secret)
	}
}

func TestServerThatNeverClosesFailsWithNothingWritten(t *testing.T) {
	path := serve(t, []byte("secret\n"), true)

	var out bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- copySecret(&out, path, 100*time.Millisecond) }()

	select {
	case err := <-done:
		if err == nil {
			t.Error("got no error")
		}
		if out.Len() != 0 {
			t.Errorf("got %q on stdout, want nothing", out.Bytes())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("still waiting on the server")
	}
}

func TestEmptyAnswerIsAnError(t *testing.T) {
	var out bytes.Buffer
	if err := copySecret(&out, serve(t, nil, false), time.Second); err == nil {
		t.Errorf("got no error, wrote %q", out.Bytes())
	}
}
