package main

import "flag"

// Very simple main function: Just parse the arguments and run the server
func main() {
	bindTo := flag.String(
		"l", "0.0.0.0:999", "interface and port to listen at")
	flag.Parse()
	runServer(*bindTo)
}
