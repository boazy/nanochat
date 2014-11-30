package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	writeTimeout      = 30 * time.Second
	loginTimeout      = 30 * time.Second
	inactivityTimeout = 5 * time.Minute
)

// message Represents a chat message
type message struct {
	user string // The user who sent the message
	text string
}

// client Represents a client connection
type client struct {
	conn net.Conn
	user string
}

// newClient Creates a new client object
func newClient(conn net.Conn, user string) *client {
	return &client{conn, user}
}

// readLine Reads a line without the trailing LF character
func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	return line[:len(line)-1], err
}

// runServer runs the main server event loop
func runServer(bindTo string) {
	// Create a listener
	listener, err := net.Listen("tcp", bindTo)
	if err != nil {
		log.Fatalf("FATAL: Cannot open %v for listening.\n%v\n", bindTo, err)
	}
	defer listener.Close()

	// Initialize channels
	clientLogin := make(chan *client)
	clientQuit := make(chan *client)
	connAccepted := make(chan net.Conn)
	messages := make(chan message)

	// The read loop
	readLoop := func(conn net.Conn) {
		// Read first line which contains the username chosen by the client
		conn.SetReadDeadline(time.Now().Add(loginTimeout))
		reader := bufio.NewReader(conn)
		user, err := readLine(reader)
		if err != nil {
			log.Printf("ERROR: Cannot read user from new connection.\n%v\n", err)
			conn.Close()
			return
		}

		// Create the client object and signal a login
		cl := newClient(conn, user)
		clientLogin <- cl

		// Signal the quit channel when this function ends
		defer func() { clientQuit <- cl }()

		for {
			// Read next chat message
			conn.SetReadDeadline(time.Now().Add(inactivityTimeout))
			cmd, err := readLine(reader)
			if err != nil {
				log.Printf("ERROR: Connection from '%v' closed prematurely.\n%v\n",
					user, err)
				return // Quit on error
			}

			if strings.HasPrefix(cmd, "*") {
				// Normal (text) chat messages starts with a '*'
				msg := message{user, cmd[1:]}
				messages <- msg
			} else {
				// Handle special commands here
				switch cmd {
				case "QUIT":
					// Client gracefully requested to quit
					return
				default:
					log.Printf("WARNING: Unknown command message '%v' ignored.\n", cmd)
				}
			}
		}
	}

	// Accept loop (accepts new connections and sends them to a channel)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("ERROR: Cannot accept connection.\n%v\n", err)
			}
			connAccepted <- conn
		}
	}()

	// Maps username to its respective client struct
	clients := map[string]*client{}

	// Event loop
	for {
		select {
		// A new connection was accepted
		case conn := <-connAccepted:
			go readLoop(conn)
		// Client connection error or client gracefuly quit
		case cl := <-clientQuit:
			log.Printf("[%v] Quits\n", cl.user)
			delete(clients, cl.user)
			cl.conn.Close()
		// New client connection successfully passed the login stage
		case cl := <-clientLogin:
			if clients[cl.user] != nil {
				log.Printf("ERROR: Username already exist: %v", cl.user)
				cl.conn.Close()
			} else {
				clients[cl.user] = cl
			}
		// One of the clients sent a message to everyody
		case msg := <-messages:
			// Format message to include sender's username.
			msgText := fmt.Sprintf("[%v]: %v\n", msg.user, msg.text)
			log.Printf(msgText)

			// Broadcast message to all other clients...
			for user, cl := range clients {
				if user != msg.user { // ...except the sender, of course.
					// We spawn a separate goroutine for each client, to avoid
					// sequentially blocking on write for every single client.
					go func(cl *client) {
						cl.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
						if _, err := cl.conn.Write([]byte(msgText)); err != nil {
							clientQuit <- cl // Quit on error
						}
					}(cl)
				}
			}
		}
	}
}
