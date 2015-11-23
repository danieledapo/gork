package main

import (
	"flag"
	"fmt"
	"gork/gork"
	"io/ioutil"
)

func main() {
	i := flag.Bool("i", true, "show game information in header")
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
	story := gork.NewZStory(buf)

	fmt.Println("\nStory file is zork1.z5")

	header := gork.NewZHeader(story)

	if *i {
		fmt.Println(header)
	}

	if *o {
		gork.DumpAllZObjects(story, header)
	}

	if *t {
		gork.DumpZObjectsTree(story, header)
	}

	if *a {
		gork.DumpAbbreviations(story, header)
	}

	if *d {
		fmt.Println(gork.NewZDictionary(story, header))
	}

	gork.NewZMachine(story, header).InterpretAll()

	fmt.Println("")
}
