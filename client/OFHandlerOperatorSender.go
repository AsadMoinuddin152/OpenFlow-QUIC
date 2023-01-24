package main

// OFHandlerOperatorSender is the executer. Every OpenFlow message will match the function if needs.
// Here is the EPM's extra operations to improve the OpenFlow protocol performance.

// OFHandlerMapperSend
// every functions add into mapping for executing.
var OFHandlerMapperSend = make(map[byte]func([]byte, *OFHandler))

// LoadOFHandlerMapperSend initiates the mapping
func LoadOFHandlerMapperSend() {
	OFHandlerMapperSend[10] = PacketInOperatorSend
	OFHandlerMapperSend[13] = PacketOutOperatorSend
	OFHandlerMapperSend[14] = FlowModOperatorSend
	OFHandlerMapperSend[18] = MultipartRequestOperatorSend
	OFHandlerMapperSend[19] = MultipartReplyOperatorSend
}

func MultipartReplyOperatorSend(message []byte, ofh *OFHandler) {
}

func MultipartRequestOperatorSend(message []byte, ofh *OFHandler) {
}

func FlowModOperatorSend(message []byte, ofh *OFHandler) {
}

func PacketOutOperatorSend(message []byte, ofh *OFHandler) {
}

func PacketInOperatorSend(message []byte, ofh *OFHandler) {
	ofh.PacketIn = true
}

func DefaultOperatorSend(message []byte, ofh *OFHandler) {
	ofh.PacketIn = false
}
