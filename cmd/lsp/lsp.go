package lsp

import (
	"context"
	"github.com/MineGame159/protocol"
	"github.com/spf13/cobra"
	"go.lsp.dev/jsonrpc2"
	"go.uber.org/zap"
	"log"
	"net"
	"os"
	"strconv"
)

var port uint16

func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Language server for Fireball",
		Run:   lspCmd,
	}

	cmd.Flags().Uint16VarP(&port, "port", "p", 0, "Port to start the LSP server on. If not specified the LSP server will use STDOUT / STDIN.")

	return cmd
}

func lspCmd(_ *cobra.Command, _ []string) {
	server := &handler{}

	stream, logger := getStream()
	_, conn, client := protocol.NewServer(context.Background(), server, stream, logger)

	server.logger = logger
	server.client = client

	server.docs = &Documents{
		client: client,
		docs:   make(map[protocol.URI]*Document),
	}

	logger.Info("Listening")
	<-conn.Done()
}

func getStream() (jsonrpc2.Stream, *zap.Logger) {
	// STDIO
	if port == 0 {
		return jsonrpc2.NewStream(os.Stdout), zap.NewNop()
	}

	// TCP
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln(err.Error())
	}
	logger.Info("Listening on :" + strconv.Itoa(int(port)))

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
