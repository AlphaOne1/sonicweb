package main

import (
	"net/http"
	"os"
)

const (
	FILE_SERVER = iota
	METRICS_SERVER
)

type serverEntry struct {
	Server  *http.Server
	SigChan *chan os.Signal
}

var serverRegister = make(map[int]serverEntry, 2)

func registerServer(kind int, server *http.Server, sigChan *chan os.Signal) {
	serverRegister[kind] = serverEntry{
		Server:  server,
		SigChan: sigChan,
	}
}

func signalizeAll(s os.Signal) {
	for _, v := range serverRegister {
		*v.SigChan <- s
	}
}
