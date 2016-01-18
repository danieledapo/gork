package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/d-dorazio/gork/gork"
)

type config struct {
	showHeader        bool
	showObjects       bool
	showObjectTree    bool
	showAbbreviations bool
	showDictionary    bool
}

func main() {
	i := flag.Bool("i", true, "show game information in header")
	o := flag.Bool("o", false, "show objects")
	t := flag.Bool("t", false, "show object tree")
	a := flag.Bool("a", false, "show abbreviations")
	d := flag.Bool("d", false, "show dictionary")
	flag.Parse()

	conf := &config{
		showHeader:        *i,
		showObjects:       *o,
		showObjectTree:    *t,
		showAbbreviations: *a,
		showDictionary:    *d,
	}

	for _, story := range flag.Args() {
		dumpStoryInfo(story, conf)
	}

}

func dumpStoryInfo(story string, conf *config) {
	buf, err := ioutil.ReadFile(story)
	if err != nil {
		fmt.Println("\nUnable to open story", story, "Error:", err)
		return
	}
	mem := gork.NewZMemory(buf)

	fmt.Println("\nStory file is", story)

	header, err := gork.NewZHeader(mem)
	if err != nil {
		panic(err)
	}

	if conf.showHeader {
		fmt.Println(header)
	}

	if conf.showObjects {
		DumpAllZObjects(mem, header)
	}

	if conf.showObjectTree {
		DumpZObjectsTree(mem, header)
	}

	if conf.showAbbreviations {
		DumpAbbreviations(mem, header)
	}

	if conf.showDictionary {
		fmt.Println(gork.NewZDictionary(mem, header))
	}

	fmt.Println("")
}

func DumpAbbreviations(mem *gork.ZMemory, header *gork.ZHeader) {
	fmt.Print("\n    **** Abbreviations ****\n\n")

	abbrs := gork.GetAbbreviations(mem, header)

	if len(abbrs) == 0 {
		fmt.Printf("  No abbreviation information.\n")
		return
	}

	for i, abbr := range abbrs {
		fmt.Printf("  [%2d] \"%s\"\n", i, abbr)
	}
}

func DumpAllZObjects(mem *gork.ZMemory, header *gork.ZHeader) {
	total, err := gork.ZObjectsCount(mem, header)
	if err != nil {
		panic(err)
	}

	fmt.Print("\n    **** Objects ****\n\n")
	fmt.Printf("  Object count = %d\n\n", total)

	for i := uint8(1); i <= total; i++ {
		obj, err := gork.NewZObject(mem, i, header)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%3d. %s", i, obj)
	}
}

func DumpZObjectsTree(mem *gork.ZMemory, header *gork.ZHeader) {

	fmt.Print("\n    **** Object tree ****\n\n")

	total, err := gork.ZObjectsCount(mem, header)
	if err != nil {
		panic(err)
	}

	var printObject func(obj *gork.ZObject, depth int)
	printObject = func(obj *gork.ZObject, depth int) {
		for {

			for j := 0; j < depth; j++ {
				fmt.Print(" . ")
			}
			fmt.Printf("[%3d] ", obj.Id())
			fmt.Printf("\"%s\"\n", obj.Name())

			if obj.ChildId() != 0 {
				childobj, err := gork.NewZObject(mem, obj.ChildId(), header)
				if err != nil {
					panic(err)
				}
				printObject(childobj, depth+1)
			}

			if obj.SiblingId() == 0 {
				break
			}
			obj, err = gork.NewZObject(mem, obj.SiblingId(), header)
			if err != nil {
				panic(err)
			}
		}
	}

	for i := uint8(1); i <= total; i++ {
		zobj, err := gork.NewZObject(mem, i, header)
		if err != nil {
			panic(err)
		}

		// root
		if zobj.ParentId() == 0 {
			printObject(zobj, 0)
			break
		}
	}
}
