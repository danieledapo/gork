package main

import (
	"flag"
	"fmt"
	"gork/gork"
	"io/ioutil"
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

	header := gork.NewZHeader(mem)

	if conf.showHeader {
		fmt.Println(header)
	}

	if conf.showObjects {
		gork.DumpAllZObjects(mem, header)
	}

	if conf.showObjectTree {
		gork.DumpZObjectsTree(mem, header)
	}

	if conf.showAbbreviations {
		gork.DumpAbbreviations(mem, header)
	}

	if conf.showDictionary {
		fmt.Println(gork.NewZDictionary(mem, header))
	}

	fmt.Println("")
}
