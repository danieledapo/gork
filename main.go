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
	story, err := ioutil.ReadFile("zork1.z5")
	if err != nil {
		panic(err)
	}

	// trust me :)
	const (
		objTblPos  = 0x02B0
		abbrTblPos = 0x01F0
		dictPos    = 0x3B21
	)

	fmt.Println("\nStory file is zork1.z5")

	if *i {
		fmt.Println(gork.NewZHeader(story))
	}

	if *o {
		gork.DumpAllZObjects(story, objTblPos, abbrTblPos)
	}

	if *t {
		gork.DumpZObjectsTree(story, objTblPos, abbrTblPos)
	}

	if *a {
		gork.DumpAbbreviations(story, abbrTblPos)
	}

	if *d {
		fmt.Println(gork.NewZDictionary(story, dictPos, abbrTblPos))
	}

	fmt.Println("")
}
