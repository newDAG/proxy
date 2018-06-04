package Dummy

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type SocketDummyProxy struct {
	nodeAddress string
	bindAddress string

	client *SocketDummyProxyClient
	server *SocketDummyProxyServer
}

func NewSocketDummyProxy(nodeAddr string,
	bindAddr string,
	timeout time.Duration,
	logger *logrus.Logger) (*SocketDummyProxy, error) {

	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	client := NewSocketDummyProxyClient(nodeAddr, timeout)
	server, err := NewSocketDummyProxyServer(bindAddr, timeout, logger)
	if err != nil {
		return nil, err
	}

	proxy := &SocketDummyProxy{
		nodeAddress: nodeAddr,
		bindAddress: bindAddr,
		client:      client,
		server:      server,
	}
	go proxy.server.listen()

	return proxy, nil
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//Implement DummyProxy interface

func (p *SocketDummyProxy) CommitCh() chan Commit {
	return p.server.commitCh
}

func (p *SocketDummyProxy) SubmitTx(tx []byte) error {
	ack, err := p.client.SubmitTx(tx)
	if err != nil {
		return err
	}
	if !*ack {
		return fmt.Errorf("Failed to deliver transaction to Dummy")
	}
	return nil
}
