// Iris - Decentralized Messaging Framework
// Copyright 2013 Peter Szilagyi. All rights reserved.
//
// Iris is dual licensed: you can redistribute it and/or modify it under the
// terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// The framework is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// Alternatively, the Iris framework may be used in accordance with the terms
// and conditions contained in a signed written agreement between you and the
// author(s).
//
// Author: peterke@gmail.com (Peter Szilagyi)

// Package stream wraps a TCP/IP network connection with the Go gob en/decoder.
//
// Note, in case of a serialization error (encoding or decoding failure), it is
// assumed that there is either a protocol mismatch between the parties, or an
// implementation bug; but in any case, the connection is deemed failed and is
// terminated.
package stream

import (
	"bufio"
	"encoding/gob"
	"log"
	"net"
	"time"
)

// Constants for the protocol TCP/IP layer
const acceptBlockTimeout = 250 * time.Millisecond

// Stream listener to accept inbound connections.
type Listener struct {
	Sink chan *Stream // Channel receiving the accepted connections

	socket *net.TCPListener // Network socket to accept connections on
	quit   chan chan error  // Termination synchronization channel
}

// TCP/IP based stream with a gob encoder on top.
type Stream struct {
	socket  *net.TCPConn      // Network connection to the remote endpoint
	buffers *bufio.ReadWriter // Buffered access to the network socket
	encoder *gob.Encoder      // Gob encoder for data serialization
	decoder *gob.Decoder      // Gob decoder for data deserialization
}

// Opens a TCP server socket and returns a stream listener, ready to accept. If
// an auto-port (0) is requested, the port is updated in the argument.
func Listen(addr *net.TCPAddr) (*Listener, error) {
	// Open the server socket
	sock, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	addr.Port = sock.Addr().(*net.TCPAddr).Port

	// Initialize and return the listener
	return &Listener{
		socket: sock,
		Sink:   make(chan *Stream),
		quit:   make(chan chan error),
	}, nil
}

// Starts the stream connection accepter, with a maximum timeout to wait for an
// established connection to be handled.
func (l *Listener) Accept(timeout time.Duration) {
	go l.accepter(timeout)
}

// Terminates the acceptor and returns any encountered errors.
func (l *Listener) Close() error {
	errc := make(chan error)
	l.quit <- errc
	return <-errc
}

// Accepts incoming connection requests, converts them info a TCP/IP gob stream
// and send them back on the sink channel.
func (l *Listener) accepter(timeout time.Duration) {
	var errc chan error
	var errv error

	// Loop until an error occurs or quit is requested
	for errv == nil && errc == nil {
		select {
		case errc = <-l.quit:
			continue
		default:
			// Accept an incoming connection but without blocking for too long
			l.socket.SetDeadline(time.Now().Add(acceptBlockTimeout))
			if conn, err := l.socket.AcceptTCP(); err == nil {
				strm := newStream(conn)
				select {
				case l.Sink <- strm:
					// Ok, connection was handled
				case <-time.After(timeout):
					log.Printf("stream: failed to handle accepted connection in %v, dropping.", timeout)
					strm.Close()
				}
			} else if !err.(net.Error).Timeout() {
				log.Printf("stream: failed to accept connection: %v.", err)
				errv = err
			}
		}
	}
	// Close upstream stream sink and socket (keep initial error, if any)
	close(l.Sink)
	if err := l.socket.Close(); errv == nil {
		errv = err
	}
	// Wait for termination sync and return
	if errc == nil {
		errc = <-l.quit
	}
	errc <- errv
}

// Creates a new, gob backed network stream based on a live TCP/IP connection.
func newStream(sock *net.TCPConn) *Stream {
	reader := bufio.NewReader(sock)
	writer := bufio.NewWriter(sock)

	return &Stream{
		socket:  sock,
		buffers: bufio.NewReadWriter(reader, writer),
		encoder: gob.NewEncoder(writer),
		decoder: gob.NewDecoder(reader),
	}
}

// Connects to a remote host and returns the connection stream.
func Dial(address string, timeout time.Duration) (*Stream, error) {
	if sock, err := net.DialTimeout("tcp", address, timeout); err != nil {
		return nil, err
	} else {
		return newStream(sock.(*net.TCPConn)), nil
	}
}

// Retrieves the raw connection object if special manipulations are needed.
func (s *Stream) Sock() *net.TCPConn {
	return s.socket
}

// Serializes an object and sends it over the wire. In case of an error, the
// connection is torn down.
func (s *Stream) Send(data interface{}) error {
	if err := s.encoder.Encode(data); err != nil {
		s.socket.Close()
		return err
	}
	return nil
}

// Flushes the outbound socket. In case of an error, the  network stream is torn
// down.
func (s *Stream) Flush() error {
	if err := s.buffers.Flush(); err != nil {
		s.socket.Close()
		return err
	}
	return nil
}

// Receives a gob of the given type and returns it. If an  error occurs, the
// network stream is torn down.
func (s *Stream) Recv(data interface{}) error {
	if err := s.decoder.Decode(data); err != nil {
		s.socket.Close()
		return err
	}
	return nil
}

// Closes the underlying network connection of a stream.
func (s *Stream) Close() error {
	return s.socket.Close()
}
