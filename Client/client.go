package main

import (
	"Client/protocol"
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"net"
	"os"
)

func checkError(err error, msg string){
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error em " + msg + "\n", err.Error())
		os.Exit(1)
	}
}

func checkParams(args []string) (string, string) {
	if len(args) != 4 {
		fmt.Fprintf(os.Stderr, "Error: Argumentos esperados: <hostname/ip> <porta> <arquivo>")
		os.Exit(1)
	}

	addr, err := net.ResolveIPAddr("ip", args[1]+args[2])
	if err == nil {
		fmt.Fprintf(os.Stderr, "Error: %s não é um hostname/ip válido.\n", args[1])
		os.Exit(1)
	}

	ip := addr.String()
	port := args[2]
	return ip, port
}

func readFile() []byte {
	file, err := os.Open(os.Args[3])
	checkError(err, "Open file")
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	fmt.Println(fileSize)

	bytes := make([]byte, fileSize)

	bufferReader := bufio.NewReader(file)
	bufferReader.Read(bytes)

	return bytes
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

func decodeDataInContent(buffer []byte) *protocol.DataLayer {
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

func sendPacket(packet gopacket.Packet, conn *net.UDPConn) *protocol.DataLayer {

	// VERIFICAR O TIPO DO PACOTE DE ACORDO COM A FLAG:
	// CLIENTE TEM 3 CASOS DE ENVIO:
	// se SYN: Envio da Solicita de criação da conexão, Sem payload;
	// se ACK: Só acontece na primeira vez que vai começar a enviar pacote de arquivo, Com payload;
	// se Nenhum: Apenas envio de pacotes de arquivo, Com payload
	// se FIN: Solicita o encerramento da conexão, Sem payload;

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

func recvPacket(conn *net.UDPConn) *protocol.DataLayer {
	// VERIFICAR O TIPO DO PACOTE DE ACORDO COM A FLAG:
	// CLIENTE TEM 3 CASOS DE RECEBIMENTO
	// se SYN e ACK: servidor criou a conexão e definiu IdConnection, Sem payload;
	// se só ACK: Confirmação que o pacote foi recebido, Sem payload;
	// se FIN e ACK: O pacote de encerramento de conexão chegou! Sem payload, encerrar aplicação os.Exit(0)!

	var result [524]byte
	_, err := conn.Read(result[:])
	checkError(err, "Read")

	// DECODIFICAÇÃO DO PACOTE QUE CHEGOU ...
	content := decodeDataInContent(result[:])

	return content
}

func sendPayload(conn *net.UDPConn) {
	binary := readFile()
	encode := base64.StdEncoding.EncodeToString(binary)
	for {
		//sendPacket(packet, conn)
		conn.Write([]byte(encode))
	}
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

	fmt.Println(packet)
	return packet
}

func handleServer(conn *net.UDPConn) {
	packet := createFirstPacket()

	content := sendPacket(packet, conn) // INIT
	printPacket("SEND", content)

	content = recvPacket(conn)
	printPacket("RECV", content)

	// COMEÇA A ENVIAR O PAYLOAD

	conn.Close()
}

func main() {
	_, port := checkParams(os.Args)

	udpAddr, err := net.ResolveUDPAddr("udp", port)
	checkError(err, "ResolveUDPAddr")

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err, "ListenUDP")

	handleServer(conn)

	os.Exit(0)
}