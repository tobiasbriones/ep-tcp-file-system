// Copyright (c) 2022 Tobias Briones. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// This file is part of https://github.com/tobiasbriones/ep-file-system-server

package main

import (
	"encoding/json"
	"log"
	"net"
	"testing"
)

const (
	testFile = "file.pdf"
)

// Side effect test. Requires a file "file.pdf" into the server's file system
// directory. It tests the server file system for write and read.
func TestReceiveSend(t *testing.T) {
	serverFileInfo, err := newTestFileInfo()
	size := serverFileInfo.Size

	requirePassedTest(t, err, "Fail to load test file info")
	downloaded := make([]byte, 0, size)
	ds := newDataStream(testFile, bufSize, func(buf []byte) {
		downloaded = append(downloaded, buf...)
	})

	StreamFile(&ds) // blocking

	// Upload the file back
	newPath := "new-file.pdf"
	CreateFile(newPath)
	for i := 0; i < cap(downloaded); i += bufSize {
		end := i + bufSize

		if end >= cap(downloaded) {
			end = cap(downloaded) - 1
		}
		chunk := downloaded[i:end]

		// Mimic sending to remote server
		WriteBuf(newPath, chunk)
	}
}

// Makes a request to the server. It can be either upload or download. After the
// initial request (status START), the server will respond with status OK.
func TestTcpConn(t *testing.T) {
	conn := initiateConn("upload")
	res := readInitialOkMsg(conn)

	if res.Status != OK {
		t.Fatal("Fail to establish the TCP connection to the server")
	}
	conn.Close()
}

func initiateConn(action string) *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr(network, getServerAddress())

	requireNoError(err)
	conn, err := net.DialTCP(network, nil, tcpAddr)
	requireNoError(err)

	info, err := newTestFileInfo()
	infoStr, err := json.Marshal(info)
	msg := Message{
		Status:  "start",
		Action:  action,
		Payload: string(infoStr),
	}
	b, err := json.Marshal(msg)
	requireNoError(err)
	_, err = conn.Write(b)
	requireNoError(err)
	return conn
}

func readInitialOkMsg(conn net.Conn) Message {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	reply := buf[:n]

	requireNoError(err)
	log.Println("Reply from server: ", string(reply))

	res := Message{}
	err = json.Unmarshal(reply, &res)

	requireNoError(err)
	return res
}

func newTestFileInfo() (FileInfo, error) {
	i := FileInfo{
		RelPath: testFile,
		Size:    0,
	}
	size, err := i.readFileSize()
	i.Size = size
	return i, err
}
