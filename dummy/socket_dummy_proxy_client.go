package Dummy

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

type SocketDummyProxyClient struct {
	nodeAddr string
	timeout  time.Duration
}

func NewSocketDummyProxyClient(nodeAddr string, timeout time.Duration) *SocketDummyProxyClient {
	return &SocketDummyProxyClient{
		nodeAddr: nodeAddr,
		timeout:  timeout,
	}
}

func (p *SocketDummyProxyClient) getConnection() (*rpc.Client, error) {
	conn, err := net.DialTimeout("tcp", p.nodeAddr, p.timeout)
	if err != nil {
		return nil, err
	}
	return jsonrpc.NewClient(conn), nil
}

func (p *SocketDummyProxyClient) SubmitTx(tx []byte) (*bool, error) {
	rpcConn, err := p.getConnection()
	if err != nil {
		return nil, err
	}
	var ack bool
	err = rpcConn.Call("Dummy.SubmitTx", tx, &ack)
	if err != nil {
		return nil, err
	}
	return &ack, nil
}
