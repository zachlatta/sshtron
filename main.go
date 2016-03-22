package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"golang.org/x/crypto/ssh"
)

const (
	sshPortEnv  = "SSH_PORT"
	httpPortEnv = "PORT"

	defaultSshPort  = "2022"
	defaultHttpPort = "3000"
)

func handler(conn net.Conn, gm *GameManager, config *ssh.ServerConfig) {
	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		fmt.Println("Failed to handshake with new client")
		return
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
			fmt.Println("could not accept channel.")
			return
		}

		// TODO: Remove this -- only temporary while we launch on HN
		//
		// To see how many concurrent users are online
		fmt.Printf("Player joined. Current stats: %d users, %d games\n",
			gm.SessionCount(), gm.GameCount())

		// Reject all out of band requests accept for the unix defaults, pty-req and
		// shell.
		go func(in <-chan *ssh.Request) {
			for req := range in {
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

		gm.HandleChannel(channel)
	}
}

func port(opt, env, def string) string {
	var port string
	if opt != "" {
		port = opt
	} else {
		port = os.Getenv(env)
		if port == "" {
			port = def
		}
	}

	return fmt.Sprintf(":%s", port)
}

func main() {

	var sshPortOpt string
	var httpPortOpt string
	flag.StringVar(&sshPortOpt, "ssh-port", "", "set the ssh port to use")
	flag.StringVar(&httpPortOpt, "http-port", "", "set the http port to use")
	flag.Parse()

	sshPort := port(sshPortOpt, sshPortEnv, defaultSshPort)
	httpPort := port(httpPortOpt, httpPortEnv, defaultHttpPort)

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

	fmt.Printf(
		"Listening on port %s for SSH and port %s for HTTP...\n",
		sshPort,
		httpPort,
	)

	go func() {
		panic(http.ListenAndServe(httpPort, http.FileServer(http.Dir("./static/"))))
	}()

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0%s", sshPort))
	if err != nil {
		panic("failed to listen for connection")
	}
	gm := NewGameManager()
	for {
		nConn, err := listener.Accept()
		if err != nil {
			panic("failed to accept incoming connection")
		}

		go handler(nConn, gm, config)
	}
}
