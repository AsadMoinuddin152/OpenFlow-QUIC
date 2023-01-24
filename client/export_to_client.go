package main

import (
	"C"
	_ "crypto/rand"
	_ "crypto/rsa"
	"crypto/tls"
	_ "crypto/x509"
	_ "encoding/pem"
	_ "math/big"
	_ "os"

	"github.com/lucas-clemente/quic-go"
)
import "log"

// main file to trigger by Open VSwitch
var (
	qcm     = NewQuicClientManager()
	ip, _   = externalIP() // get the local ip for logging
	logaddr = "/tmp/" + ip.String() + "ovs-go_module.log"
	tlsConf = &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	config = &quic.Config{
		KeepAlive: true,
	}
	LOG *Logger
)

//export InitClient
/* InitClient return the fd to make a Unix domain socket connection */
/* The file format is "/tmp/mininet-int(fd).sock"*/
func InitClient(addr *C.char) C.int {
	qcm.InitClient()                                     // initiate QUIC connection manager
	fd := qcm.Connect(C.GoString(addr), tlsConf, config) // make a QUIC connection
	return C.int(fd)
}

//export Close
/* Close, close all connection, including Unix domain socket*/
func Close(empty C.int) C.int {
	qcm.CloseCleanClient()
	log.Printf("OVS Close QUIC \n")
	return C.int(0)
}
