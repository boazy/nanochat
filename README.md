Nanochat
========

A miniature, proof-of-concept, chat server written in Go.

Installation
============

1. Unpack the archive to an an src directory under your $GOPATH.
2. There are two different executables to build:
   a. The client (cmd/nanochat)
   b. The server (cmd/nanochatd)
3. Build and install both (to $GOPATH/bin) with build.sh.

Usage
=====

Client: nanochat <server IP/host[:port]> <username>
Server: nanochatd [IP/host to bind]:[port to listen]

Protocol
========

The chat porotocol is a simple line-based textual protocol running over TCP.
Lines MUST be separated by an LF (ASCII 0x10) character.

When the client connects to the server, the first line it sends is the
username the client will be using. If the username is not available (in use by
another user), the server MAY terminate the connection.

After choosing the username, each line sent by the client is interpreted as a
command. If the command starts with a '*' character, it is a normal message
(which will be boradcast to all other users). Otherwise it's a special
command.

The server will send back to the client all the messages it received from
other clients (but not by THIS client - i.e. there is no remote echo).

Special commands
----------------
QUIT   gracefully quit the chat (the chat server SHOULD disconnect the client
       after receiving this command)

Server Implementation
=====================
The server is concurrent and implemented using goroutines. Behind the scenes
(that is, on the OS level), I/O in Go is asynchronous, and when a goroutine
blocks on I/O, it yields to scheduelr which replaces it with another
goroutine. The go scheduler can run goroutines on multiple threads by setting
the GOMAXPROCS environment variable to be larger than 1, but the server code
itself is thread-agnostic, and probably does not have a lot to gain from
CPU/core parallelism (which is the purpose of using multiple threads in the go
scheduler), since it has very little CPU-bound operations going on.

No actual load testing and profiling was done, but a real server would
probably have to be properly tested. This is the only way to properly ensure
that it has no performance sinkholes or race conditions.
