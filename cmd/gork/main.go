package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/d-dorazio/gork/gork"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a game")
	}

	story := os.Args[1]

	buf, err := ioutil.ReadFile(story)
	if err != nil {
		panic(err)
	}

	logfile, err := os.Create(strings.Split(story, ".")[0] + ".log")
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	mem := gork.NewZMemory(buf)

	header := gork.NewZHeader(mem)

	gork.NewZMachine(mem, header).InterpretAll()

	fmt.Println("")
}
