FROM golang:1.6

MAINTAINER Douglas

WORKDIR /TRABALHO

COPY .. /TRABALHO

RUN apt-get install git
RUN apt-get update
RUN apt install libpcap-dev -y
RUN go get github.com/google/gopacket

EXPOSE 3200