package lsp

import (
	"context"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
	"log"
	"net"
	"os"
	"strconv"
)

func Start(port uint16) {
	server := &handler{
		files: make(map[protocol.URI]string),
	}

	stream, logger := getStream(port)

	_, conn, client := protocol.NewServer(context.Background(), server, stream, logger)

	server.logger = logger
	server.client = client

	logger.Info("Listening")
	<-conn.Done()
}

func getStream(port uint16) (jsonrpc2.Stream, *zap.Logger) {
	// STDIO
	if port == 0 {
		return jsonrpc2.NewStream(os.Stdout), zap.NewNop()
	}

	// TCP
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln(err.Error())
	}

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		logger.Fatal(err.Error())
	}

	conn, err := listener.Accept()
	if err != nil {
		logger.Fatal(err.Error())
	}

	return jsonrpc2.NewStream(conn), logger
}
