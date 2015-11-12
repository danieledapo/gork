package main

import (
	"fmt"
	"gork/gork"
	"io/ioutil"
)

func main() {
	// test only Zork :)
	story, err := ioutil.ReadFile("zork1.z5")
	if err != nil {
		panic(err)
	}

	fmt.Print("\n**** Header ****\n\n")

	zmem := gork.NewZHeader(story)
	zmem.Dump()

	// trust me :)
	const objTblPos = 0x02B0
	const abbrTblPos = 0x01F0

	fmt.Print("\n\n**** Objects ****\n\n")

	gork.DumpAllZObjects(story, objTblPos, abbrTblPos)

	fmt.Print("\n\n**** Object Tree ****\n\n")

	gork.DumpZObjectsTree(story, objTblPos, abbrTblPos)

	fmt.Println("")
}
