package main

import (
	"flag"
	"fmt"
	"gork/gork"
	"io/ioutil"
)

func main() {
	i := flag.Bool("i", false, "show game information in header")
	o := flag.Bool("o", false, "show objects")
	t := flag.Bool("t", false, "show object tree")
	a := flag.Bool("a", false, "show abbreviations")
	d := flag.Bool("d", false, "show dictionary")
	flag.Parse()

	// test only Zork :)
	buf, err := ioutil.ReadFile("zork1.z5")
	if err != nil {
		panic(err)
	}
	mem := gork.NewZMemory(buf)

	fmt.Println("\nStory file is zork1.z5")

	header := gork.NewZHeader(mem)

	if *i {
		fmt.Println(header)
	}

	if *o {
		gork.DumpAllZObjects(mem, header)
	}

	if *t {
		gork.DumpZObjectsTree(mem, header)
	}

	if *a {
		gork.DumpAbbreviations(mem, header)
	}

	if *d {
		fmt.Println(gork.NewZDictionary(mem, header))
	}

	gork.NewZMachine(mem, header).InterpretAll()

	fmt.Println("")
}
