## Download da imagem
Baixe a imagem remota para a sua máquina:
```
docker pull douglasmoraiis/udp-reliable:latest
```
## Instalação
Execute a imagem baixada em um container:
```
docker run --name trabalho-container -it douglasmoraiis/udp-reliable
```
## Execução da aplicação
Quando o terminal da imagem iniciar, execute o arquivo `servidor.go`:
```
go run servidor/servidor.go <porta> <pasta-diretorio>
```
Agora abra um novo terminal no seu computador e execute o seguinte comando:
```
docker container exec -it trabalho-container bash
```
Ele vai abrir um novo terminal da mesma imagem que já está em execução.

Execute o arquivo `cliente.go`:
```
go run cliente/cliente.go <ip/hostname> <porta> <arquivo>
```
Pronto a comunicação foi efetuada e o arquivo foi enviado para a pasta destino definida no servidor.
O cliente finalizou, mas o servidor ainda aguarda novas conexões.
