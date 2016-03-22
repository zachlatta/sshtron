package main

// Implement ssh.Channel, so we pretend to be an ssh connection.

import "github.com/pkg/term"

type TermChannel struct {
	T *term.Term
}

func NewTermChannel() *TermChannel {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)

	return &TermChannel{t}
}

func (tc TermChannel) Restore() {
	tc.T.Restore()
}

func (tc TermChannel) Read(data []byte) (int, error) {
	return tc.T.Read(data)
}

func (tc TermChannel) Write(data []byte) (int, error) {
	return tc.T.Write(data)
}

func (tc TermChannel) Close() error {
	return tc.T.Close()
}
