package main

import (
	"../protocol"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"net"
	"os"
	"path/filepath"
	"time"
)

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func checkParams(args []string) string {
	if len(args) != 4 {
		fmt.Fprintf(os.Stderr, "Error: Argumentos esperados: <hostname/ip> <porta> <arquivo>")
		os.Exit(1)
	}
	addr := net.JoinHostPort(args[1], args[2])
	return addr
}

func readFile() []byte {
	file, err := os.Open(os.Args[3])
	checkError(err, "Open file")
	defer file.Close()

	fileInfo, err := file.Stat()
	checkError(err, "file.Stat")
	fileSize := fileInfo.Size()

	fileBuffer := make([]byte, fileSize)

	bufferReader := bufio.NewReader(file)
	bufferReader.Read(fileBuffer)

	return fileBuffer
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

func sendPacket(packet gopacket.Packet, conn *net.UDPConn) *protocol.DataLayer {

	// PARTE USADA APENAS PARA IMPRIMIR NA TELA
	decodePacket := packet.Layer(protocol.DataLayerType)
	if decodePacket == nil {
		fmt.Fprintf(os.Stderr, "decodePacket is nil!", error.Error)
	}
	content := decodePacket.(*protocol.DataLayer)

	//ENVIANDO DADO PARA A CONEXÃO
	_, err := conn.Write(packet.Data())
	checkError(err, "conn.Write")

	return content
}

func recvPacket(conn *net.UDPConn) (*protocol.DataLayer, *net.UDPAddr) {
	// VERIFICAR O TIPO DO PACOTE DE ACORDO COM A FLAG:
	// CLIENTE TEM 3 CASOS DE RECEBIMENTO
	// se SYN e ACK: servidor criou a conexão e definiu IdConnection, Sem payload;
	// se só ACK: Confirmação que o pacote foi recebido, Sem payload;
	// se FIN e ACK: O pacote de encerramento de conexão chegou! Sem payload, encerrar aplicação os.Exit(0)!

	var result [524]byte
	_, address, err := conn.ReadFromUDP(result[:])
	checkError(err, "Read")

	// DECODIFICAÇÃO DO PACOTE QUE CHEGOU ...
	content := decodeBytesInContent(result[:])

	return content, address
}

func createFirstPacket() gopacket.Packet {
	var buffer bytes.Buffer

	// DADOS DO PRIMEIRO PACOTE
	var seqNum uint32 = 12345
	var ackNum uint32 = 0
	var idCon uint16 = 0
	var flags = parseFlags(false, true, false)
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

func handleServer(conn *net.UDPConn) {
	var IDConn uint16 = 0

	fileBytes := readFile()
	var initIndexPacket = 0
	var endIndexPacket = 512

	if len(fileBytes) <= 512 {
		endIndexPacket = len(fileBytes)
	}

	// ENVIO DOS PACOTES
	packet := createFirstPacket()
	content := sendPacket(packet, conn) // INIT
	printPacket("SEND", content)

	for {
		content, _ = recvPacket(conn)
		printPacket("RECV", content)
		IDConn = content.IdConnection

		isAck, isSyn, isFin := readFlags(content.Flags)

		// se SYN e ACK: servidor criou a conexão e definiu IdConnection, Sem payload;
		if isAck && isSyn {
			// ENVIO A INFORMAÇÃO DA EXTENSAO DO ARQUIVO
			var seqNum = content.AckNumber
			var ackNum = content.SequenceNumber + 1
			var idCon = IDConn
			var flags = parseFlags(true, false, false)

			// ENVIA A EXTENSÃO DO ARQUIVO
			fileExtension := filepath.Ext(os.Args[3])
			var payload = []byte(fileExtension)

			newPacket := encodeDataInPacket(seqNum, ackNum, idCon, flags, payload)

			newContent := sendPacket(newPacket, conn)
			printPacket("SEND", newContent)

			// se só ACK: Confirmação que o pacote foi recebido, Sem payload;
		} else if isAck && !isFin {

			var seqNum = content.AckNumber
			var ackNum uint32 = 0
			var idCon = IDConn
			var flags uint16

			var payload []byte
			if endIndexPacket >= len(fileBytes) {
				endIndexPacket = len(fileBytes)
			}

			if initIndexPacket > len(fileBytes) {
				flags = parseFlags(false, false, true)
			} else {
				flags = parseFlags(false, false, false)
				payload = fileBytes[initIndexPacket:endIndexPacket]
			}

			initIndexPacket += 512
			endIndexPacket += 512

			newPacket := encodeDataInPacket(seqNum, ackNum, idCon, flags, payload)

			newContent := sendPacket(newPacket, conn)
			printPacket("SEND", newContent)

			// se FIN e ACK: O pacote de encerramento de conexão chegou! Sem payload, encerrar aplicação os.Exit(0)!
		} else if isAck && isFin {
			print("O arquivo foi enviado e o fechamento da conexão foi confirmado! Desligando...")
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}
	}

}

func main() {
	addr := checkParams(os.Args)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	checkError(err, "ResolveUDPAddr")

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err, "ListenUDP")

	handleServer(conn)

	os.Exit(0)
}