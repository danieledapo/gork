package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/d-dorazio/gork/gork"
	"github.com/gorilla/websocket"
)

type WSServer struct {
	story  string
	mem    *gork.ZMemory
	header *gork.ZHeader
}

func (server *WSServer) run(addr string) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	wsHandler := func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Failed to upgrade %s\n", err)
			return
		}

		remoteAddr := conn.RemoteAddr().String()
		logFilename := storyLogFilename(server.story)
		logfile, err := os.Create(fmt.Sprintf("wsserver_%s_%s", remoteAddr, logFilename))
		if err != nil {
			panic(err)
		}
		defer logfile.Close()
		logger := log.New(logfile, "", log.LstdFlags)

		wsdev := &gork.ZWSDev{Conn: conn}

		zm, err := gork.NewZMachine(server.mem, server.header, wsdev, logger)
		if err != nil {
			panic(err)
		}
		zm.InterpretAll()
	}

	http.HandleFunc("/play", wsHandler)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
