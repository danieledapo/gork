package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/danieledapo/gork/gork"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type SshServer struct {
	id_rsa string
	story  string
	mem    *gork.ZMemory
	header *gork.ZHeader
}

func (server *SshServer) run(addr string) {
	config := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, pwd []byte) (*ssh.Permissions, error) {
			// it's just a demo, in a real server this should not be done :)
			if string(pwd) == "<3txt" {
				return nil, nil
			}
			return nil, errors.New("pwd error, please enter <3txt to enter")
		},
	}

	// You can generate a keypair with 'ssh-keygen -t rsa'
	privateBytes, err := ioutil.ReadFile(server.id_rsa)
	if err != nil {
		fmt.Printf("Failed to load private key %s\n", server.id_rsa)
		return
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		fmt.Println("Failed to parse private key")
		return
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to listen on %s (%s)\n", addr, err)
		return
	}
	fmt.Printf("Listening on %s...", addr)

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept incoming connection (%s)\n", err)
			continue
		}

		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			fmt.Printf("Failed to handshake (%s)\n", err)
			continue
		}

		fmt.Printf("New SSH connection from %s (%s)\n", sshConn.RemoteAddr(), sshConn.ClientVersion())

		go ssh.DiscardRequests(reqs)
		go server.handleChannels(sshConn.User(), chans)
	}
}

func (server *SshServer) handleChannels(user string, chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go server.handleChannel(user, newChannel)
	}
}

func (server *SshServer) handleChannel(user string, newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		fmt.Printf("Could not accept channel (%s)", err)
		return
	}
	defer connection.Close()

	logfile, err := os.Create(fmt.Sprintf("%s_%s", user, storyLogFilename(server.story)))
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	logger := log.New(logfile, "", log.LstdFlags)

	terminal := terminal.NewTerminal(connection, "")
	zsshterm := &gork.ZSshTerminal{Term: terminal}

	zm, err := gork.NewZMachine(server.mem, server.header, zsshterm, logger)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				termLen := req.Payload[3]
				w, h := parseDims(req.Payload[termLen+4:])
				terminal.SetSize(w, h)
			case "window-change":
				w, h := parseDims(req.Payload)
				terminal.SetSize(w, h)
			}
		}
	}()

	defer func() {
		recover()
	}()

	zm.InterpretAll()

}

func parseDims(b []byte) (int, int) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return int(w), int(h)
}
