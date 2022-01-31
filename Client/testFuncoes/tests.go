package main

import (
	"fmt"
)

type DataLayer struct {
	SequenceNumber uint32
	AckNumber uint32
	IdConnection uint16
	Flags uint16
	Payload []byte
}


func readFlags(flags uint16) (bool, bool, bool) {
	var ack, syn, fin uint16
	var isAck, isSyn, isFin bool
	ack = flags & (1 << 2)
	syn = flags & (1 << 1)
	fin = flags & (1)

	if ack != 0 {
		isAck = true
	}
	if syn != 0 {
		isSyn = true
	}
	if fin != 0 {
		isFin = true
	}
	return isAck, isSyn, isFin
}

func printPacketRecv(content DataLayer) {
	isAck, isSyn, isFin := readFlags(content.Flags)
	var strAck = ""
	var strSyn = ""
	var strFin = ""

	if isAck {
		strAck = " ACK"
	}
	if isSyn {
		strSyn = " SYN"
	}
	if isFin {
		strFin = " FIN"
	}

	fmt.Println("RECV ",
		content.SequenceNumber,
		content.AckNumber,
		content.IdConnection,
		strAck,
		strSyn,
		strFin,
	)
}

func main() {
	var content DataLayer
	content.SequenceNumber = 123
	content.AckNumber = 321
	content.IdConnection = 1
	content.Flags = 0

	printPacketRecv(content)
}