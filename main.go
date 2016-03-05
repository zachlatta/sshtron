package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
)

const (
	keyUp    = 'w'
	keyLeft  = 'a'
	keyDown  = 's'
	keyRight = 'd'
)

func handler(conn net.Conn, game *Game, config *ssh.ServerConfig) {
	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		panic("failed to handshake")
	}
	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			panic("could not accept channel.")
		}

		// Reject all out of band requests accept for the unix defaults, pty-req and
		// shell.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				fmt.Println("New request:", req.Type)
				switch req.Type {
				case "pty-req":
					req.Reply(true, nil)
					continue
				case "shell":
					req.Reply(true, nil)
					continue
				}
				req.Reply(false, nil)
			}
		}(requests)

		fmt.Println("Received new connection")
		session := NewSession(channel)
		game.AddSession(session)

		reader := bufio.NewReader(channel)
		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				fmt.Println(err)
				break
			}

			switch r {
			case keyUp:
				fmt.Println("Up")
				session.Player.Pos.Y -= 1
			case keyLeft:
				fmt.Println("Left")
				session.Player.Pos.X -= 1
			case keyDown:
				fmt.Println("Down")
				session.Player.Pos.Y += 1
			case keyRight:
				fmt.Println("Right")
				session.Player.Pos.X += 1
			}

			game.hub.Rerender <- struct{}{}
		}
	}
}

func main() {
	// Everyone can login!
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		panic("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Create the game itself
	game := NewGame(50, 25)
	go game.Run()

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		panic("failed to listen for connection")
	}
	for {
		nConn, err := listener.Accept()
		if err != nil {
			panic("failed to accept incoming connection")
		}

		go handler(nConn, game, config)
	}
}
