package proxy

import (
	"fmt"
	"os"

	"time"

	"github.com/newDAG/crypto"
	"github.com/newDAG/ledger"
	bproxy "github.com/newDAG/proxy/dummy"
	"github.com/sirupsen/logrus"
)

type State struct {
	stateHash []byte
	logger    *logrus.Logger
}

//CommitBlock
func (a *State) CommitBlock(block ledger.Block) ([]byte, error) {
	a.logger.WithField("block", block).Debug("CommitBlock")
	err := a.writeBlock(block)
	if err != nil {
		return a.stateHash, err
	}
	return a.stateHash, nil
}

func (a *State) writeBlock(block ledger.Block) error {
	file, err := a.getFile()
	if err != nil {
		a.logger.Error(err)
		return err
	}
	defer file.Close()

	// write some text to file
	//and update state hash
	hash := a.stateHash
	for _, tx := range block.Transactions() {
		_, err = file.WriteString(fmt.Sprintf("%s\n", string(tx)))
		if err != nil {
			a.logger.Error(err)
			return err
		}
		hash = crypto.SimpleHashFromTwoHashes(hash, crypto.SHA256(tx))
	}

	err = file.Sync()
	if err != nil {
		a.logger.Error(err)
		return err
	}

	a.stateHash = hash

	return nil
}

func (a *State) writeMessage(tx []byte) {
	file, err := a.getFile()
	if err != nil {
		a.logger.Error(err)
		return
	}
	defer file.Close()

	// write some text to file
	_, err = file.WriteString(fmt.Sprintf("%s\n", string(tx)))
	if err != nil {
		a.logger.Error(err)
	}
	err = file.Sync()
	if err != nil {
		a.logger.Error(err)
	}
}

func (a *State) getFile() (*os.File, error) {
	path := "messages.txt"
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
}

//------------------------------------------------------

type DummySocketClient struct {
	state      *State
	DummyProxy *bproxy.SocketDummyProxy
	logger     *logrus.Logger
}

func NewDummySocketClient(clientAddr string, nodeAddr string, logger *logrus.Logger) (*DummySocketClient, error) {

	DummyProxy, err := bproxy.NewSocketDummyProxy(nodeAddr, clientAddr, 1*time.Second, logger)
	if err != nil {
		return nil, err
	}

	state := State{
		stateHash: []byte{},
		logger:    logger,
	}
	state.writeMessage([]byte(clientAddr))

	client := &DummySocketClient{
		state:      &state,
		DummyProxy: DummyProxy,
		logger:     logger,
	}

	go client.Run()

	return client, nil
}

func (c *DummySocketClient) Run() {
	for {
		select {
		case commit := <-c.DummyProxy.CommitCh():
			c.logger.Debug("CommitBlock")
			stateHash, err := c.state.CommitBlock(commit.Block)
			commit.Respond(stateHash, err)
		}
	}
}

func (c *DummySocketClient) SubmitTx(tx []byte) error {
	return c.DummyProxy.SubmitTx(tx)
}
