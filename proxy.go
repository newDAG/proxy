package proxy

import "github.com/newDAG/ledger"

type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block ledger.Block) ([]byte, error)
}

type DummyProxy interface {
	CommitCh() chan ledger.Block
	SubmitTx(tx []byte) error
}
