package store

// Raft Commands

const (
	CmdCreateIndex = iota
	CmdAddDocument
	CmdAddNode
	// CmdRemoveNode
)

// Param stores case sensitivity if the command is CreateIndex,
// or the Document ID if the command is AddDocument
type Command struct {
	CmdId   int
	IdxName string
	Param   int

	// peer info when new node joins
	NodeAddress     string
	NodeHTTPAddress string
}
