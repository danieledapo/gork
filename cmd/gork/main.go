package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/d-dorazio/gork/gork"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a game")
		return
	}

	story := os.Args[1]

	buf, err := ioutil.ReadFile(story)
	if err != nil {
		panic(err)
	}

	logfile, err := os.Create(storyLogFilename(story))
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	mem := gork.NewZMemory(buf)

	header := gork.NewZHeader(mem)

	gork.NewZMachine(mem, header, gork.ZTerminal{}).InterpretAll()
}

func storyLogFilename(story string) string {
	name := path.Base(story)
	tmp := strings.Split(name, ".")
	if len(tmp) > 1 {
		name = tmp[0]
	}
	return name + ".log"
}
