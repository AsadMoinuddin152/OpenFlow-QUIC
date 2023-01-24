package main

import "errors"

var CanNotFindClientByMap = errors.New("CanNotFindClientByMap")

var StreamRecvError = errors.New("QUIC Stream Recv Error")

var StreamRecv8Error = errors.New("QUIC Stream Recv 8 Error")

var StreamRecvOver8Error = errors.New("QUIC Stream Recv Over 8 Error")

var ClientHasBeenClose = errors.New("Client Has Been Close, check the invoking stack.")

var QuicListenerUnavailable = errors.New("quic listener is unavailable")