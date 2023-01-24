package main

import (
	"log"
	"net"
	"sync"
	"time"
)

type UnixClient struct {
	UsdPath string
	UAddr   *net.UnixAddr
	Conn    *net.UnixConn
	wmutex  sync.Mutex
	rmutex  sync.Mutex
}

func NewUnixClient(path string) UnixClientI {
	Uaddr, _ := net.ResolveUnixAddr("unix", path)
	return &UnixClient{Conn: nil,
		UAddr:   Uaddr,
		UsdPath: path,
		wmutex:  sync.Mutex{},
		rmutex:  sync.Mutex{},
	}
}

func (u *UnixClient) CreateUnixConnection() {
	for {
		Conn, err := net.DialUnix("unix", nil, u.UAddr)
		u.Conn = Conn
		if err == nil {
			log.Printf("Unix connected!\n")
			return
		}
		log.Printf("DialUnix error %v\n", err)
		time.Sleep(50 * time.Millisecond)
	}
}

func (u *UnixClient) CloseConnection() {
	err := u.Conn.Close()
	go func() {
		time.Sleep(5 * time.Second)
		u.Conn = nil
		u.UAddr = nil
		u.UsdPath = ""
		log.Printf("Uds: clean the unix socket\n")
	}()
	if err != nil {
		log.Printf("Uds: CloseConnection got error :%v\n", err)
		return
	}
}

func (u *UnixClient) Read(p []byte) (n int, err error) {
	return u.Conn.Read(p)
}

func (u *UnixClient) Write(p []byte) (n int, err error) {
	return u.Conn.Write(p)
}
