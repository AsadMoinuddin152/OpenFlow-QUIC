package main

import (
	"C"
	"fmt"
	"log"
	"os"

	quic "github.com/quic-go/quic-go"
)

var (
	config = quic.Config{
		KeepAlive: true,
	}
)

type QuicServerManager struct {
	Server QuicServerI
}

func (m *QuicServerManager) CreateQuicServer(UnixPath string) C.int {
	if m.Server.CheckListenerAlive() == 0 { // 0 -> still alive 1 -> listener is nil
		return C.int(0)
	}
	listener, err := quic.ListenAddr(Addr, generateTLSConfig(), &config)
	if err != nil {
		panic(err)
	}
	logFile, logErr := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		fmt.Println("Fail to log", *logFile, "Server start Failed")
		os.Exit(1)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
	LOG = NewLogger()
	go LOG.LogRoutine()
	log.Printf("init RYU-QUIC %v \n", UnixPath)
	go func() {
		err := m.Server.ServerRun(listener, UnixPath)
		if err != nil {
			panic("Server run over, needs to change panic to C.int retrieval")
		}
	}() //unblock
	return C.int(0)
}

func (m *QuicServerManager) CloseServer(AddrC *C.char) C.int {
	Addr := C.GoString(AddrC)
	if Addr == "0.0.0.0:6653" {
		m.Server.CloseAllConnection()
		m.Server.CloseListener()
		m.Server = nil
		return C.int(2) /* close listener -> server */
	}
	err := m.Server.CloseClientConnection(Addr)
	if err != nil {
		return C.int(1)
	}
	return C.int(0)

}
