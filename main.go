package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Client struct {
	conn     net.Conn
	nickname string
}

var (
	clients   []*Client
	clientsMu sync.Mutex
)

func handleClient(client *Client) {
	defer client.conn.Close()

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		message := scanner.Text()
		broadcastMessage(fmt.Sprintf("%s: %s", client.nickname, message))
	}

	// Remove client when they disconnect
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for i, c := range clients {
		if c == client {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	broadcastMessage(fmt.Sprintf("%s has left the chat", client.nickname))
}

func broadcastMessage(message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for _, client := range clients {
		_, err := fmt.Fprintln(client.conn, message)
		if err != nil {
			fmt.Println("Error broadcasting message:", err)
			return
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Prompt the client for a nickname
		_, err = fmt.Fprintf(conn, "Enter your nickname: ")
		if err != nil {
			fmt.Println("Error sending prompt:", err)
			conn.Close()
			continue
		}

		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			nickname := scanner.Text()

			// Create a new client and add it to the list
			client := &Client{conn: conn, nickname: nickname}
			clientsMu.Lock()
			clients = append(clients, client)
			clientsMu.Unlock()

			// Broadcast a message to other clients about new connection
			go broadcastMessage(fmt.Sprintf("%s has joined the chat", nickname))

			// Handle client messages concurrently
			go handleClient(client)
		}
	}
}
