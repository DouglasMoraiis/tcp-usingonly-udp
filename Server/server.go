package main

import (
	"Client/protocol"
	"fmt"
	"github.com/google/gopacket"
	"net"
	"os"
)

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
	port := os.Args[1]
	dir := os.Args[2]
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
	// SERVIDOR TEM 3 CASOS DE ENVIO
	// se SYN e ACK: Servidor criou a conexão e definiu IdConnection, Sem payload;
	// se só ACK: Confirmação que o pacote foi recebido, Sem payload;
	// se FIN e ACK: O pacote de encerramento de conexão chegou! Sem payload, encerrar aplicação!

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
	// SERVIDOR TEM 3 CASOS DE RECEBIMENTO
	// se SYN: Cliente solicitou uma nova conexão, Sem payload;
	// se Nenhum: Cliente enviou pacote de arquivo, Com payload;
	// se ACK: Confirmação que o pacote de criação de conexão chegou, Já vem com payload ;
	// se FIN: Solicitação de encerramento de conexão! Sem payload, encerrar conexão (conn.close())!

	var result [524]byte
	fmt.Println("Recebendo dados...")
	_, err := conn.Read(result[0:])
	checkError(err, "Read")

	// DECODIFICAÇÃO DO PACOTE QUE CHEGOU ...
	content := decodeDataInContent(result[0:])

	return content
}

func handleClient(conn *net.UDPConn, dir string)  {
	content := recvPacket(conn)
	printPacket("RECV", content)

	// AINDA FALTA CRIAR O PACKET DO SEND ACK DE CONEXAO
	//content = sendPacket(packet, conn)

/*	var netBuffer [524]byte
	var fileBuffer [524]byte

	size, err := conn.Read(netBuffer[0:])
	fmt.Println(size)
	checkError(err, "Read")

	_, err = base64.StdEncoding.Decode(fileBuffer[0:size], netBuffer[0:size])
	checkError(err, "Decode")

	dirFile := strings.TrimSpace(dir) + "file.png"
	err = ioutil.WriteFile(dirFile, fileBuffer[0:size], 0666)
	checkError(err, "WriteFile")
	os.Exit(0)*/
}

func main() {
	port, dir := checkParams(os.Args)

	udpAddr, _ := net.ResolveUDPAddr("udp", port)
	conn, _ := net.ListenUDP("udp", udpAddr)

	handleClient(conn, dir)
}
