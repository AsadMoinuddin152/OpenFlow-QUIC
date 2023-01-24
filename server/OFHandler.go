package main

var _ OFHandlerI = &OFHandler{}

// OFHandler is the analyzer of EPM. It contains the status if the specific OpenFlow messages comes or leaves.
type OFHandler struct {
	PacketIn         bool
	MultipartRequest bool
	MultipartReply   bool
}

func NewOFHandler() OFHandlerI {
	LoadOFHandlerMapperRecv()
	LoadOFHandlerMapperSend()
	return &OFHandler{
		MultipartReply:   false,
		MultipartRequest: false,
		PacketIn:         false,
	}
}

// OFMessageRecvOperation run strategies in receiver
func (o *OFHandler) OFMessageRecvOperation(message []byte, Uds UnixClientI) (int, error) {
	switch message[1] {
	case 18: // if multipartrequest, then it should wait for 1 RTT if got packet-in
		buf := make([]byte, len(message))
		copy(buf, message)
		go func() {
			_, _ = OFHandlerMapperRecv[18](buf, o, Uds)
		}()
		return len(message), nil
	default:
		return DefaultOperatorRecv(message, o, Uds)
	}
}

// OFMessageSendOperation run strategies in sender
func (o *OFHandler) OFMessageSendOperation(message []byte) {
	switch message[1] {
	case 10: // if multipartrequest, then it should wait for 1 RTT if got packet-in
		OFHandlerMapperSend[10](message, o)
	default:
		DefaultOperatorSend(message, o)
	}
}
