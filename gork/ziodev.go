package gork

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type ZIODev interface {
	Print(...interface{})
	ReadLine() string
}

type ZTerminal struct{}

func (_ ZTerminal) Print(s ...interface{}) {
	for _, si := range s {
		fmt.Print(si)
	}
}

func (_ ZTerminal) ReadLine() string {
	r := bufio.NewReader(os.Stdin)

	s, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}

	return s
}

type ZSshTerminal struct {
	term *terminal.Terminal
}

func (sshTerm ZSshTerminal) Print(s ...interface{}) {
	for _, si := range s {
		sshTerm.term.Write([]byte(fmt.Sprint(si)))
	}
}

func (sshTerm ZSshTerminal) ReadLine() string {
	l, err := sshTerm.term.ReadLine()

	if err != nil {
		panic(err)
	}

	return l
}
