package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/danieledapo/gork/gork"
)

func main() {
	identity := flag.String("identity", "", "ssh key to use to start server")
	addr := flag.String("address", "0.0.0.0:4273", "address to listen on for ssh connections")
	ws := flag.Bool("ws", false, "start the web socket server on addr")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Please provide a game")
		return
	}

	story := flag.Args()[0]

	buf, err := ioutil.ReadFile(story)
	if err != nil {
		panic(err)
	}
	mem := gork.NewZMemory(buf)

	header, err := gork.NewZHeader(mem)
	if err != nil {
		panic(err)
	}

	if *identity != "" {
		server := &SshServer{
			id_rsa: *identity,
			story:  story,
			mem:    mem,
			header: header,
		}
		server.run(*addr)
	} else if *ws {
		server := &WSServer{
			story:  story,
			mem:    mem,
			header: header,
		}
		server.run(*addr)
	} else {
		terminalUI(story, mem, header)
	}
}

func terminalUI(story string, mem *gork.ZMemory, header *gork.ZHeader) {
	logfile, err := os.Create(storyLogFilename(story))
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	logger := log.New(logfile, "", log.LstdFlags)

	zm, err := gork.NewZMachine(mem, header, gork.ZTerminal{}, logger)
	if err != nil {
		panic(err)
	}

	if err := zm.InterpretAll(); err != nil {
		panic(err)
	}
}

func storyLogFilename(story string) string {
	name := path.Base(story)
	tmp := strings.Split(name, ".")
	if len(tmp) > 1 {
		name = tmp[0]
	}
	return name + ".log"
}
