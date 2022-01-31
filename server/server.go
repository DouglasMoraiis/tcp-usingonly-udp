package main

import (
	"../protocol"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

// GLOBAL
var QTD_CONNECTIONS uint16 = 0
var EXTENSION_FILE string
var FILE []byte

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func checkParams(args []string) (string, string) {
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Error: Argumentos esperados: <porta> <diretório>")
		os.Exit(1)
	}
	port := ":" + args[1]
	dir := args[2]
	return port, dir
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

func parseFlags(isAck bool, isSyn bool, isFin bool) uint16 {
	var flags uint16 = 0
	if isAck {
		flags = flags | (1 << 2)
	}
	if isSyn {
		flags = flags | (1 << 1)
	}
	if isFin {
		flags = flags | (1)
	}

	return flags
}

func printPacket(prefix string, content *protocol.DataLayer) {
	isAck, isSyn, isFin := readFlags(content.Flags)
	var strAck = ""
	var strSyn = ""
	var strFin = ""

	if isAck {
		strAck = "ACK"
	}
	if isSyn {
		strSyn = "SYN"
	}
	if isFin {
		strFin = "FIN"
	}

	fmt.Println(
		prefix,
		content.SequenceNumber,
		content.AckNumber,
		content.IdConnection,
		strAck,
		strSyn,
		strFin,
	)
}

func decodeBytesInContent(buffer []byte) *protocol.DataLayer {
	packet := gopacket.NewPacket(
		buffer[0:],
		protocol.DataLayerType,
		gopacket.Default,
	)
	decodePacket := packet.Layer(protocol.DataLayerType)
	if decodePacket == nil {
		fmt.Fprintf(os.Stderr, "decodePacket is nil!", error.Error)
	}
	content := decodePacket.(*protocol.DataLayer)
	return content
}

func encodeDataInPacket(seqNum uint32, ackNum uint32, idCon uint16, flags uint16, payload []byte) gopacket.Packet {
	var buffer bytes.Buffer

	var seqNumBytes = make([]byte, 4)
	var ackNumBytes = make([]byte, 4)
	var idConBytes = make([]byte, 2)
	var flagsBytes = make([]byte, 2)

	// PARSE DADOS PARA []BYTE
	binary.BigEndian.PutUint32(seqNumBytes, seqNum)
	binary.BigEndian.PutUint32(ackNumBytes, ackNum)
	binary.BigEndian.PutUint16(idConBytes, idCon)
	binary.BigEndian.PutUint16(flagsBytes, flags)

	// JUNTA TODOS EM UM UNICO []BYTE
	buffer.Write(seqNumBytes)
	buffer.Write(ackNumBytes)
	buffer.Write(idConBytes)
	buffer.Write(flagsBytes)
	buffer.Write(payload)

	var packet = gopacket.NewPacket(
		buffer.Bytes(),
		protocol.DataLayerType,
		gopacket.Default,
	)

	return packet
}

func sendPacket(packet gopacket.Packet, conn *net.UDPConn, receiver *net.UDPAddr) *protocol.DataLayer {

	decodePacket := packet.Layer(protocol.DataLayerType)
	if decodePacket == nil {
		fmt.Fprintf(os.Stderr, "decodePacket is nil!", error.Error)
	}
	content := decodePacket.(*protocol.DataLayer)

	//ENVIANDO DADO PARA A CONEXÃO
	_, err := conn.WriteToUDP(packet.Data(), receiver )
	checkError(err, "conn.Write")

	return content
}

func recvPacket(conn *net.UDPConn) (*protocol.DataLayer, *net.UDPAddr) {

	var result [524]byte
	size, address, err := conn.ReadFromUDP(result[0:])
	checkError(err, "Read")

	// DECODIFICAÇÃO DO PACOTE QUE CHEGOU ...
	content := decodeBytesInContent(result[0:size])

	return content, address
}

func createFirstAckPacket(previousContent *protocol.DataLayer, IDConn *uint16) gopacket.Packet {
	QTD_CONNECTIONS++
	*IDConn = QTD_CONNECTIONS

	var buffer bytes.Buffer

	// DADOS DO PRIMEIRO PACOTE
	var seqNum uint32 = 4321
	var ackNum = previousContent.SequenceNumber + 1
	var idCon = *IDConn
	var flags = parseFlags(true, true, false)
	var payload []byte = nil

	var seqNumBytes = make([]byte, 4)
	var ackNumBytes = make([]byte, 4)
	var idConBytes = make([]byte, 2)
	var flagsBytes = make([]byte, 2)

	// PARSE DADOS PARA []BYTE
	binary.BigEndian.PutUint32(seqNumBytes, seqNum)
	binary.BigEndian.PutUint32(ackNumBytes, ackNum)
	binary.BigEndian.PutUint16(idConBytes, idCon)
	binary.BigEndian.PutUint16(flagsBytes, flags)

	// JUNTA TODOS EM UM UNICO []BYTE
	buffer.Write(seqNumBytes)
	buffer.Write(ackNumBytes)
	buffer.Write(idConBytes)
	buffer.Write(flagsBytes)
	buffer.Write(payload)

	var packet = gopacket.NewPacket(
		buffer.Bytes(),
		protocol.DataLayerType,
		gopacket.Default,
	)

	return packet
}

func handleClient(conn *net.UDPConn, dir string)  {
	var IDConn uint16 = 0

	for {
		content, address := recvPacket(conn)
		printPacket("RECV", content)

		isAck, isSyn, isFin := readFlags(content.Flags)
		if isSyn {

			packet := createFirstAckPacket(content, &IDConn)
			fmt.Println("UM PACOTE SYN FOI RECEBIDO! CRIANDO CONEXÃO!")
			newContent := sendPacket(packet, conn, address)
			printPacket("SEND", newContent)

		} else if isAck || (!isAck && !isSyn && !isFin) {
			var seqNum uint32 = 4322
			var ackNum = content.SequenceNumber + uint32(len(content.Payload))
			var idCon = IDConn
			var flags = parseFlags(true, false, false)
			var payload []byte

			// PACOTE COM EXTENSÃO DO ARQUIVO RECEBIDO
			if isAck {
				EXTENSION_FILE = strings.TrimSpace(string(content.LayerPayload()))
			} else {
				saveInFILE(content.LayerPayload())
			}
			packet := encodeDataInPacket(seqNum, ackNum, idCon, flags, payload)

			newContent := sendPacket(packet, conn, address)
			printPacket("SEND", newContent)

		} else if isFin {
			var seqNum uint32 = 4322
			var ackNum uint32 = 0
			var idCon = IDConn
			var flags = parseFlags(true, false, true)
			var payload []byte

			packet := encodeDataInPacket(seqNum, ackNum, idCon, flags, payload)

			exportFileLocalStorage(IDConn, dir)

			newContent := sendPacket(packet, conn, address)
			printPacket("SEND", newContent)
		}
	}

	// se Nenhum: Cliente enviou pacote de arquivo, Com payload;
	// se ACK: Confirmação que o pacote de criação de conexão chegou, Já vem com payload ;
	// se FIN: Solicitação de encerramento de conexão! Sem payload, encerrar conexão (conn.close())!
}

func exportFileLocalStorage(IDConn uint16, dir string) {
	fileName := strconv.Itoa(int(IDConn))
	dirFile := strings.TrimSpace(dir) + fileName + EXTENSION_FILE

	err := ioutil.WriteFile(dirFile, FILE[0:], 0666)
	checkError(err, "WriteFile")
}

func saveInFILE(payload []byte) {
	FILE = append(FILE, payload...)
}

func main() {
	port, dir := checkParams(os.Args)

	udpAddr, _ := net.ResolveUDPAddr("udp", port)
	conn, _ := net.ListenUDP("udp", udpAddr)
	fmt.Println("SERVIDOR INICIADO")

	handleClient(conn, dir)

}