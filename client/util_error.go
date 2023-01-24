package main

import "errors"

var StreamRecvError = errors.New("QUIC Stream Recv Error")

var StreamRecv8Error = errors.New("QUIC Stream Recv 8 Error")

var StreamRecvOver8Error = errors.New("QUIC Stream Recv Over 8 Error")

var TryingToGetDataFromClosedBuffer = errors.New("some functions are trying to get data from a closed buffer, check the code by debugging.")

var UsingBrokenConnection = errors.New("Using Broken Connection ")