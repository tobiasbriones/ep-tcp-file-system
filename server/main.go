// Copyright (c) 2022 Tobias Briones. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// This file is part of https://github.com/tobiasbriones/ep-file-system-server

// Entry point for the file system server.
//
// Author: Tobias Briones
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	port    = 8080
	network = "tcp"
	bufSize = 1024
)

type Status string

const (
	START Status = "start"
	OK    Status = "ok"
	DATA  Status = "data"
	EOF   Status = "eof"
	ERROR Status = "error"
)

type Message struct {
	Status
	Action  string
	Payload string
	Data    []byte
}

func main() {
	server, err := net.Listen(network, getServerAddress())

	defer server.Close()
	requireNoError(err)
	listen(server)
}

func listen(server net.Listener) {
	for {
		conn, err := server.Accept()
		requireNoError(err)
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	dec := json.NewDecoder(conn)

	var msg Message

	err := dec.Decode(&msg)

	if err != nil {
		log.Println("Skipped: ", conn)
		return
	}
	log.Println(msg.Status)
	switch msg.Status {
	case START:
		handleStatusStart(conn, msg)
		break
	default:
		writeStatus(ERROR, conn)
		break
	}
}

func handleStatusStart(conn net.Conn, msg Message) {
	switch msg.Action {
	case "upload":
		handleUpload(conn, msg)
		break
	case "download":
		handleDownload(conn, msg)
		break
	default:
		writeStatus(ERROR, conn)
		break
	}
}

func handleDownload(conn net.Conn, msg Message) {
	info := getFileInfo(msg)

	log.Println(info)
	// TODO
	writeStatus(OK, conn)
}

func handleUpload(conn net.Conn, msg Message) {
	info := getFileInfo(msg)

	writeStatus(OK, conn)

	_, err := os.Create(info.getPath())
	requireNoError(err)

	log.Println("Writing file:", info.RelPath, "Size:", info.Size)

	// Status = DATA, wait for chunks only
	count := int64(0)
	for {
		chunk := readChunk(conn)

		WriteBuf(info.RelPath, chunk)

		count += int64(len(chunk))
		if count >= info.Size {
			break
		}
		if len(chunk) == 0 {
			log.Print(
				"Fail to read file chunk: ",
				"The EOF was before the right position",
			)
			writeStatus(ERROR, conn)
			conn.Close()
			return
		}
	}
	if count != info.Size {
		log.Println(
			"Fail to finish writing the file:",
			"More bytes were written",
		)
		writeStatus(ERROR, conn)
		conn.Close()
		return
	}

	// Wait for EOF signal = empty chunk
	chunk := readChunk(conn)

	if len(chunk) != 0 {
		log.Println("Fail to read EOF signal")
		writeStatus(ERROR, conn)
		return
	}
	writeStatus(OK, conn)
	log.Println("File successfully written")
}

func readChunk(conn net.Conn) []byte {
	b := make([]byte, bufSize)
	n, err := conn.Read(b)

	if err != nil {
		if err.Error() != "EOF" {
			log.Println("Error reading chunk:", err)
			requireNoError(err)
		}
	}
	return b[:n]
}

func getFileInfo(msg Message) FileInfo {
	info := FileInfo{}
	err := json.Unmarshal([]byte(msg.Payload), &info)

	requireNoError(err)
	return info
}

func writeStatus(status Status, conn net.Conn) {
	msg := Message{
		Status: status,
	}
	b, err := json.Marshal(msg)

	requireNoError(err)
	_, err = conn.Write(b)

	requireNoError(err)
}

func getServerAddress() string {
	return fmt.Sprintf("localhost:%v", port)
}
