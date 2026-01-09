package store

import (
	"encoding/json"
	"fmt"
	"gocene/config"
	"gocene/internal/utils"
	"log"
	"sync"

	"github.com/hashicorp/raft"
	"github.com/minio/minio-go/v7"
)

// Store implements the Raft functionality on top of the indexes.

type Store struct {
	ActiveIndices map[string]*Index
	mc            *minio.Client

	// lock should be used only used while snapshotting
	mu *sync.Mutex

	// raft stuff
	Raft     *raft.Raft
	RaftDir  string
	RaftBind string

	// for forwarding if not leader
	PeerHTTP map[string]string
}

type fsm Store

type Node struct {
	NodeID      string `json:"node_id"`
	Address     string `json:"address"`
	HTTPAddress string `json:"http_address"`
}

type StatusResult struct {
	Me        Node   `json:"me"`
	Leader    Node   `json:"leader"`
	Followers []Node `json:"followers"`
}

func New(mc *minio.Client) *Store {
	s := &Store{
		mc:       mc,
		PeerHTTP: make(map[string]string),
	}
	s.Init()

	return s
}

func (s *Store) Init() {
	// if bootstrap, become leader. else join using join address
	s.RaftBind = config.RaftAddress
	s.RaftDir = config.RaftDirectory

	s.ActiveIndices = make(map[string]*Index)

	err := s.Open()
	if err != nil {
		log.Fatalln("could not create a new store, err: ", err.Error())
	}

}

func (s *Store) AddDocument(idxName string, docData map[string]any) (docId int, err error) {

	if s.Raft.State() != raft.Leader {
		return 0, ErrNotLeader
	}

	if _, ok := s.ActiveIndices[idxName]; !ok {
		return 0, ErrIdxDoesNotExist
	}

	// store docto S3
	err = utils.StoreDocumentToMinio(s.mc, s.ActiveIndices[idxName].NextDocID, docData, idxName)
	if err != nil {
		return 0, err
	}

	// use Command.Param to store Document ID

	// raft apply
	c := Command{
		CmdId:   CmdAddDocument,
		IdxName: idxName,
		Param:   s.ActiveIndices[idxName].NextDocID,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return 0, err
	}

	f := s.Raft.Apply(b, config.RaftTimeout)

	if f.Error() != nil {
		return 0, f.Error()
	}

	if resp := f.Response(); resp != nil {
		if ferr, ok := resp.(error); ok {
			return 0, ferr
		}

		if docid, ok := resp.(int); ok {
			return docid, nil
		}
	}

	return 0, fmt.Errorf("nil response from FSM")
}

func (s *Store) CreateIndex(idxName string, cs bool) (err error) {

	if s.Raft.State() != raft.Leader {
		return ErrNotLeader
	}

	param := 0

	// raft apply
	// use Command.Param to store case sensitivity
	if cs {
		param = 1
	}

	c := Command{
		CmdId:   CmdCreateIndex,
		IdxName: idxName,
		Param:   param,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.Raft.Apply(b, config.RaftTimeout)

	if f.Error() != nil {
		return f.Error()
	}

	if resp := f.Response(); resp != nil {
		if ferr, ok := resp.(error); ok {
			return ferr
		}
		return fmt.Errorf("unexpected response from FSM: %T", resp)
	}

	return nil
}

func (s *Store) AddNode(addr, httpAddr string) (err error) {

	if s.Raft.State() != raft.Leader {
		return ErrNotLeader
	}

	// raft apply
	// use Command.NodeAddress and Command.NodeHTTPAddress to store the new node's info into Raft logs.

	c := Command{
		CmdId:           CmdAddNode,
		NodeAddress:     addr,
		NodeHTTPAddress: httpAddr,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.Raft.Apply(b, config.RaftTimeout)

	if f.Error() != nil {
		return f.Error()
	}

	if resp := f.Response(); resp != nil {
		if ferr, ok := resp.(error); ok {
			return ferr
		}
		return fmt.Errorf("unexpected response from FSM: %T", resp)
	}

	return nil
}

// Join request
// Assumes the leader request redirection if not leader has been handled in the service.
func (s *Store) Join(nodeID, addr, httpAddr string) (err error) {

	log.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := s.Raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {

			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				log.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := s.Raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := s.Raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	log.Printf("node %s at %s joined successfully", nodeID, addr)

	// commit the join into the Raft logs
	return s.AddNode(addr, httpAddr)
}

// Status returns Raft information (id, address, and HTTP reachable address)
// about the Store, who I am, who the leader is, and who the followers are.
func (s *Store) Status() (StatusResult, error) {
	leaderServerAddr, leaderId := s.Raft.LeaderWithID()
	leader := Node{
		NodeID:      string(leaderId),
		Address:     string(leaderServerAddr),
		HTTPAddress: s.PeerHTTP[string(leaderServerAddr)],
	}

	servers := s.Raft.GetConfiguration().Configuration().Servers
	followers := []Node{}
	me := Node{
		Address: s.RaftBind,
	}
	for _, server := range servers {
		if server.ID != leaderId {
			followers = append(followers, Node{
				NodeID:      string(server.ID),
				Address:     string(server.Address),
				HTTPAddress: s.PeerHTTP[string(server.Address)],
			})
		}

		if string(server.Address) == s.RaftBind {
			me = Node{
				NodeID:      string(server.ID),
				Address:     string(server.Address),
				HTTPAddress: s.PeerHTTP[string(server.Address)],
			}
		}
	}

	status := StatusResult{
		Me:        me,
		Leader:    leader,
		Followers: followers,
	}

	return status, nil
}
