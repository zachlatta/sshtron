package main

// Implement ssh.Channel, so we pretend to be an ssh connection.

import "os"

type TermChannel struct{}

func NewTermChannel() *TermChannel {
	return &TermChannel{}
}

func (tc TermChannel) Read(data []byte) (int, error) {
	return os.Stdin.Read(data)
}

func (tc TermChannel) Write(data []byte) (int, error) {
	return os.Stdout.Write(data)
}

func (tc TermChannel) Close() error {
	panic("exiting application")
}
