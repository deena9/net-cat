package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const defaultPort = "8989"
const maxConnections = 10

var (
	clients      = make(map[net.Conn]string)
	clientsMutex sync.Mutex
	chatHistory  []string
)

func main() {
	port := getPort()
	fmt.Printf("Listening on port :%s\n", port)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		if len(clients) >= maxConnections {
			fmt.Fprintln(conn, "Server is full. Try again later.")
			conn.Close()
			continue
		}
		go handleConnection(conn)
	}
}

func getPort() string {
	if len(os.Args) < 2 {
		return defaultPort
	}
	if len(os.Args) == 2 {
		return os.Args[1]
	}
	fmt.Println("[USAGE]: ./TCPChat $port")
	os.Exit(1)
	return ""
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Fprintln(conn, "Welcome to TCP-Chat!")
	fmt.Fprintln(conn, asciiArt())
	fmt.Fprint(conn, "[ENTER YOUR NAME]: ")

	name := readLine(conn)
	if name == "" {
		fmt.Fprintln(conn, "Name cannot be empty.")
		return
	}

	clientsMutex.Lock()
	clients[conn] = name
	clientsMutex.Unlock()

	broadcast(fmt.Sprintf("%s has joined our chat...", name), conn)

	// Send chat history to the new client
	clientsMutex.Lock()
	for _, msg := range chatHistory {
		fmt.Fprintln(conn, msg)
	}
	clientsMutex.Unlock()

	for {
		msg := readLine(conn)
		if msg == "" {
			continue
		}
		if msg == "exit" {
			break
		}
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		formattedMsg := fmt.Sprintf("[%s][%s]: %s", timestamp, name, msg)
		clientsMutex.Lock()
		chatHistory = append(chatHistory, formattedMsg)
		clientsMutex.Unlock()
		broadcast(formattedMsg, conn)
	}

	clientsMutex.Lock()
	delete(clients, conn)
	clientsMutex.Unlock()
	broadcast(fmt.Sprintf("%s has left our chat...", name), conn)
}

func broadcast(message string, sender net.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		if client != sender {
			fmt.Fprintln(client, message)
		}
	}
}

func readLine(conn net.Conn) string {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return line[:len(line)-1]
}

func asciiArt() string {
	return `         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    '.       | ' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     '-'       '--'
`
}
