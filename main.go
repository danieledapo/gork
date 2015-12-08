package main

import (
	"fmt"
	"gork/gork"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	logfile, err := os.Create("gork.log")
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	// test only Zork :)
	buf, err := ioutil.ReadFile("zork1.z5")
	if err != nil {
		panic(err)
	}
	mem := gork.NewZMemory(buf)

	fmt.Println("\nStory file is zork1.z5")

	header := gork.NewZHeader(mem)

	gork.NewZMachine(mem, header).InterpretAll()

	fmt.Println("")
}
