# Vigenere Cipher over Network Sockets

This is a client-server application that uses network sockets for communication.

The client can send a message along with a key to the server.

The server then responds with the Vigenere cipher applied to the message.

Case is kept and special characters are ignored

---

## Setup
1. Clone repo
```sh
git clone https://github.com/kvnbanunu/vigenere-cipher
```
2. Build using make
```sh
make build-all
```

or

Build with Go

```sh
go build cmd/server/main.go -o bin/server
go build cmd/client/main.go -o bin/client
cp config.json bin/
```

---

## Run
1. Start Server
```sh
./bin/server <host ip> <port>
```
2. Send request with Client
```sh
./bin/client <message> <key> <host ip> <port>
```
Both programs can be run with a -h flag to display a help message

Both programs can also be run without arguments with the following config file

---

## Config
config.json includes two fields that can be changed (You do not need to rebuild)

- bufferSize sets the size of the buffer for read/write
- content sets the default message
- key sets the default encryption key
- ip sets the default host ip address
- port sets the default port

