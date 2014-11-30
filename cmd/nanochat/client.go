package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

// Some timeout consts
const (
	connectTimeout    = 30 * time.Second
	writeTimeout      = 15 * time.Second
	inactivityTimeout = 5 * time.Minute
)

// connect Connects and logs in to the specified server and username.
// If it fails, the entire process is shut down.
func connect(server string, username string) net.Conn {
	fmt.Printf("Connecting to %v...", server)
	conn, err := net.DialTimeout("tcp", server, connectTimeout)
	if err != nil {
		fmt.Println(" Error!")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(" Connected.")

	// First we need to tell the server our username
	writeLine(conn, username)

	return conn
}

// writeLine tries to write a line (with timeout) to a net.Conn.
// If it fails, the entire process is shut down.
func writeLine(conn net.Conn, line string) {
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if _, err := conn.Write([]byte(line + "\n")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// readLine Reads a line without the trailing LF character
func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	return line[:len(line)-1], err
}

// readMsgs reads all chat server messages and pushes them into a channel
func readMsgs(conn net.Conn, msgs chan string) {
	reader := bufio.NewReader(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(inactivityTimeout))
		line, err := readLine(reader)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		msgs <- line
	}
}

// The main function
func main() {
	// Parse args
	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments!")
		fmt.Println("Usage: nanochat <server host:port> <username>")
		os.Exit(1)
	}
	server := os.Args[1]
	username := os.Args[2]

	// Connect to server (could fail and end process)
	conn := connect(server, username)
	defer conn.Close()

	// This channel is signaled with true when the user triggers a quit command
	// from the UI.
	quitChan := make(chan bool)

	// Create the UI and run it
	cui := newChatUI()
	cui.run(quitChan)

	// Spawn a goroutine to read messages
	go readMsgs(conn, cui.chatMsgs)

	// Main event loop. Handles quits and writes to server.
	for {
		select {
		case hasQuit := <-quitChan:
			if hasQuit {
				writeLine(conn, "QUIT")
				os.Exit(0)
			}
		case inputMsg := <-cui.inputMsgs:
			writeLine(conn, "*"+inputMsg)
		}
	}
}
