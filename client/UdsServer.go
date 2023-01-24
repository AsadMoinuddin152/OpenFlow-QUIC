package main

import (
	"log"
	"net"
	"syscall"
	"time"
)

var _ UnixServerI = &UnixServer{}

type UnixServer struct {
	UsdPath  string
	UAddr    *net.UnixAddr
	Listener *net.UnixListener
	Conn     *net.UnixConn
}

// UdsServer is for OVS to connect the EPM to switch OF messages.
// when QuicServerConnection Opened a new QUIC connection, here, Uds Open a Server
// to be connected by OVS.
// logs should be deleted.

func NewUnixServer(path string) UnixServerI {
	// delete the Uds first if it existed.
	syscall.Unlink(path)
	return &UnixServer{
		UsdPath:  path,
		UAddr:    nil,
		Listener: nil,
		Conn:     nil,
	}
}

func (u *UnixServer) CreateUnixServer() {
	log.Printf("CreateUnixServer: unix server path %v\n", u.UsdPath)
	laddr, _ := net.ResolveUnixAddr("unix", u.UsdPath)
	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		panic(err)
	}
	log.Printf("CreateUnixServer: waiting for conn from unix socks\n")
	conn, err := listener.AcceptUnix()
	log.Printf("CreateUnixServer: got conn from unix socks\n")
	if err != nil {
		log.Printf("CreateUnixServer: got error: %v, or close socket\n", err)
		panic(err)

	}
	u.Conn = conn
	u.Listener = listener
	u.UAddr = laddr
}

func (u *UnixServer) CloseConnection() {
	_ = u.Conn.Close()
	_ = u.Listener.Close()
	log.Printf("CloseConnection: CloseConnection got error!!!\n")
	go func() {
		time.Sleep(5 * time.Second)
		u.Conn = nil
		u.UAddr = nil
		u.Listener = nil
		u.UsdPath = ""
		log.Printf("CloseConnection: clean the unix socket\n")
	}()
}

func (u *UnixServer) Read(p []byte) (n int, err error) {
	return u.Conn.Read(p)
}

func (u *UnixServer) Write(p []byte) (n int, err error) {
	return u.Conn.Write(p)
}
