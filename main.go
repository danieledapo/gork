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
	// buf, err := ioutil.ReadFile("zork2.z5")
	// buf, err := ioutil.ReadFile("zork3.z5")
	// buf, err := ioutil.ReadFile("hhgg.z3")
	if err != nil {
		panic(err)
	}
	mem := gork.NewZMemory(buf)

	header := gork.NewZHeader(mem)

	gork.NewZMachine(mem, header).InterpretAll()

	fmt.Println("")
}
