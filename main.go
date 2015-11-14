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
	flag.Parse()

	// test only Zork :)
	story, err := ioutil.ReadFile("zork1.z5")
	if err != nil {
		panic(err)
	}
	// trust me :)
	const objTblPos = 0x02B0
	const abbrTblPos = 0x01F0

	fmt.Println("\nStory file is zork1.z5")

	if *i {
		fmt.Printf("%s\n", gork.NewZHeader(story))
	}

	if *o {
		gork.DumpAllZObjects(story, objTblPos, abbrTblPos)
	}

	if *t {
		gork.DumpZObjectsTree(story, objTblPos, abbrTblPos)
	}

	fmt.Println("")
}
