// Copyright (c) 2022 Tobias Briones. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// This file is part of https://github.com/tobiasbriones/ep-file-system-server

package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
)

func listen(server net.Listener) {
	for {
		conn, err := server.Accept()
		requireNoError(err)
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	msg, err := readMessage(conn)

	if err != nil {
		log.Println("Skipped: ", conn)
		return
	}
	log.Println(msg.Status)
	switch msg.Status {
	case Start:
		handleStatusStart(conn, msg)
		break
	default:
		writeStatus(Error, conn)
		break
	}
}

func handleStatusStart(conn net.Conn, msg Message) {
	payload, err := msg.StartPayload()
	requireNoError(err)

	switch payload.Action {
	case ActionUpload:
		handleUpload(conn, payload)
		break
	case ActionDownload:
		handleDownload(conn, payload)
		break
	default:
		writeStatus(Error, conn)
		break
	}
}

func handleDownload(conn net.Conn, payload StartPayload) {
	log.Println(payload.FileInfo)
	// TODO
	writeStatus(Ok, conn)
}

func handleUpload(conn net.Conn, payload StartPayload) {
	info := payload.FileInfo

	writeStatus(Ok, conn)

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
			writeStatus(Error, conn)
			conn.Close()
			return
		}
	}
	if count != info.Size {
		log.Println(
			"Fail to finish writing the file:",
			"More bytes were written",
		)
		writeStatus(Error, conn)
		conn.Close()
		return
	}

	// Wait for EOF signal = empty chunk
	chunk := readChunk(conn)

	if len(chunk) != 0 {
		log.Println("Fail to read EOF signal")
		writeStatus(Error, conn)
		return
	}
	writeStatus(Ok, conn)
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

func writeStatus(status Status, conn net.Conn) {
	msg := Message{
		Status: status,
	}
	enc := json.NewEncoder(conn)
	err := enc.Encode(msg)
	requireNoError(err)
}

func readMessage(conn net.Conn) (Message, error) {
	var msg Message
	dec := json.NewDecoder(conn)
	err := dec.Decode(&msg)
	return msg, err
}
