package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"math"
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
	_, port := checkParams(os.Args)

	file, err := os.Open(os.Args[3])
	checkError(err, "Open file")
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	//INICIO DA MODIFICAÇÃO PARA DIVIDIR O ARQUIVO EM PARTES E SALVAR AS PARTES NA PASTA LOCAL DO CLIENTE;
	//MODIFICAÇÕES A SEREM FEITAS(FUNÇÃO SER COLOCADA NA CAMADA DE PROTOCOLOS || ENVIAR A PARTE DO ARQUIVO PARA O SERVIDOR POR CONEXÃO NO LUGAR DE SALVAR NA PASTA CLIENTE)
	const filePartsSize = 200
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(filePartsSize)))

	fmt.Printf("Dividindo o arquivo em %d partes\n", totalPartsNum)

	bytes := make([]byte, fileSize)

	bufferReader := bufio.NewReader(file)
	bufferReader.Read(bytes)


	for i := uint64(0); i < totalPartsNum; i++{
		partSize := int(math.Min(filePartsSize, float64(fileSize-int64(i*filePartsSize))))
		partBuffer := make([]byte, partSize)

		file.Read(partBuffer)

		encode := base64.StdEncoding.EncodeToString(partBuffer)

		udpAddr, err := net.ResolveUDPAddr("udp", port)
		checkError(err, "ResolveUDPAddr")

		conn, err := net.DialUDP("udp", nil, udpAddr)
		checkError(err, "ListenUDP")

		conn.Write([]byte(encode))

		conn.Close()

/*
		// write to disk
		fileName := "Parte_" + strconv.FormatUint(i, 10)
		_, err := os.Create(fileName)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// write/save buffer to disk
		ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)

		fmt.Println("Split to : ", fileName)

 */
	}

	//FIM DA MODIFICAÇÃO
/*
	fmt.Println(fileSize)


	bytes := make([]byte, fileSize)

	bufferReader := bufio.NewReader(file)
	bufferReader.Read(bytes)

	return bytes

 */
	return bytes
}

func main() {
	readFile()
	/*
	_, port := checkParams(os.Args)

	binary := readFile()

	encode := base64.StdEncoding.EncodeToString(binary)

	udpAddr, err := net.ResolveUDPAddr("udp", port)
	checkError(err, "ResolveUDPAddr")

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err, "ListenUDP")

	conn.Write([]byte(encode))

	os.Exit(0)

	 */
}