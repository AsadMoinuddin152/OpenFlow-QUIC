package main

import "C"

var (
	qsm         = &QuicServerManager{}
	logFileName = "/tmp/ryu-go_module.log"
	ServerInit  = 0
	LOG         *Logger
	Addr        = ":6633" // here, the addr needs to be changed, it should be provided by RYU
)

//export InitQuicServerManager
// InitQuicServerManager active the QuicServer and initiate it.
func InitQuicServerManager(UnixPath *C.char) C.int {
	if ServerInit == 0 {
		qsm.Server = NewQuicServer()
		ServerInit = 1
	}
	return qsm.CreateQuicServer(C.GoString(UnixPath))
}

//export CloseClient
func CloseClient(UnixPath *C.char) C.int {
	ret := qsm.CloseServer(UnixPath)
	if ret == C.int(2) {
		ServerInit = 0
	}
	return C.int(0)
}
