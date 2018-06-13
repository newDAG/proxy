package proxy

import (
	"time"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/newdag/ledger"
	"github.com/sirupsen/logrus"
)

//-------client---------------------------------------------------------------------------------------------
type APP_StateHash struct {
	Hash []byte
}

type socketAppProxyClient struct {
	clientAddr string
	timeout    time.Duration
	logger     *logrus.Logger
}

//call the client rpc State.CommitBlock
func (p *socketAppProxyClient) commitBlock(block ledger.Block) ([]byte, error) {
	var stateHash APP_StateHash

	conn, err := net.DialTimeout("tcp", p.clientAddr, p.timeout)
	if err != nil {
		return nil, err
	}
	rpcConn := jsonrpc.NewClient(conn)

	err = rpcConn.Call("State.CommitBlock", block, &stateHash)

	p.logger.WithFields(logrus.Fields{
		"block":      block.Index(),
		"state_hash": stateHash.Hash,
	}).Debug("AppProxyClient.commitBlock")

	return stateHash.Hash, err
}

//-------server---------------------------------------------------------------------------------------------

type socketAppProxyServer struct {
	netListener *net.Listener
	rpcServer   *rpc.Server
	submitCh    chan []byte
	logger      *logrus.Logger
}

func newSocketAppProxyServer(bindAddress string, logger *logrus.Logger) *socketAppProxyServer {
	server := &socketAppProxyServer{ submitCh: make(chan []byte),	logger:   logger,	}
	rpcServer := rpc.NewServer()
	rpcServer.RegisterName("NewDAG", server)
	server.rpcServer = rpcServer

	l, err := net.Listen("tcp", bindAddress)
	if err != nil {
		logger.WithField("error", err).Error("Failed to listen")
	}
	server.netListener = &l

	return server
}

//call by rpc
func (p *socketAppProxyServer) SubmitTx(tx []byte, ack *bool) error {
	p.logger.Debug("SubmitTx")
	p.submitCh <- tx
	*ack = true
	return nil
}

func (p *socketAppProxyServer) listen() {
	for {
		conn, err := (*p.netListener).Accept()
		if err != nil { 
			p.logger.WithField("error", err).Error("Failed to accept")
		}

		go (*p.rpcServer).ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

//----------------------------------------------------------------------------------------------------
type SocketAppProxy struct {
	clientAddress string
	bindAddress   string

	client *socketAppProxyClient
	server *socketAppProxyServer

	logger *logrus.Logger
}

func NewSocketAppProxy(clientAddr string, bindAddr string, timeout time.Duration, logger *logrus.Logger) *SocketAppProxy {
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.DebugLevel
	}

	client := &socketAppProxyClient{ clientAddr: clientAddr, timeout: timeout, logger: logger, }
	server := newSocketAppProxyServer(bindAddr, logger)

	proxy := &SocketAppProxy{
		clientAddress: clientAddr,
		bindAddress:   bindAddr,
		client:        client,
		server:        server,
		logger:        logger,
	}
	go proxy.server.listen()

	return proxy
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//Implement AppProxy Interface
type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block ledger.Block) ([]byte, error)
}

func (p *SocketAppProxy) SubmitCh() chan []byte {
	return p.server.submitCh
}

func (p *SocketAppProxy) CommitBlock(block ledger.Block) ([]byte, error) {
	return p.client.commitBlock(block)
}
