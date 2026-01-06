package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gocene/config"
	"gocene/internal/utils"
	"io"
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
}

type fsm Store

type Node struct {
	NodeID  string `json:"node_id"`
	Address string `json:"address"`
}

type StatusResult struct {
	Me        Node   `json:"me"`
	Leader    Node   `json:"leader"`
	Followers []Node `json:"followers"`
}

func New(mc *minio.Client) *Store {
	s := &Store{
		mc: mc,
	}
	s.Init()

	return s
}

func (s *Store) Init() {
	// if bootstrap, become leader. else join using join address
}

func (s *Store) AddDocument(idxName string, docData map[string]any) (docId int, err error) {

	if s.Raft.State() != raft.Leader {
		return 0, ErrNotLeader
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

	return c.Param, f.Error()
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

	return f.Error()
}

func (idx *Index) LoadDocumentsIntoIndex() (err error) {
	reader := bufio.NewReader(idx.DocList)

	for {
		docStr, err := reader.ReadString('\n')
		if err == io.EOF {
			if len(docStr) > 0 {
				doc, err := CreateDocumentFromJSON(docStr)
				if err != nil {
					return err
				}
				idx.LoadDocument(doc)
			}
			break
		}

		if err != nil {
			return err
		}

		doc, err := CreateDocumentFromJSON(docStr)
		if err != nil {
			return err
		}
		idx.LoadDocument(doc)

	}

	log.Println("added docs to idx", idx.Name)

	return nil
}

// Join request
// Assumes the leader request redirection if not leader has been handled in the service.
func (s *Store) Join(nodeID, addr string) (err error) {

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
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
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
	return nil
}

// Status returns information about the Store.
func (s *Store) Status() (StatusResult, error) {
	leaderServerAddr, leaderId := s.Raft.LeaderWithID()
	leader := Node{
		NodeID:  string(leaderId),
		Address: string(leaderServerAddr),
	}

	servers := s.Raft.GetConfiguration().Configuration().Servers
	followers := []Node{}
	me := Node{
		Address: s.RaftBind,
	}
	for _, server := range servers {
		if server.ID != leaderId {
			followers = append(followers, Node{
				NodeID:  string(server.ID),
				Address: string(server.Address),
			})
		}

		if string(server.Address) == s.RaftBind {
			me = Node{
				NodeID:  string(server.ID),
				Address: string(server.Address),
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

// func (s *Store) Init() {

// 	// if bootstrap set and raftdir empty, then become leader
// 	//
// 	// if bootstrap is false, join

// 	// create directories for index files and segments if not present

// 	s.ActiveIndices = make(map[string]*Index)

// 	os.MkdirAll(config.IndexDataDirectory, os.ModePerm)
// 	os.MkdirAll(config.IndexDocListDirectory, os.ModePerm)

// 	// load all files from doc list folder into index
// 	idxFiles, err := os.ReadDir(config.IndexDocListDirectory)
// 	if err != nil {
// 		log.Fatalln("could not read index dir, err: ", err.Error())
// 	}

// 	if len(idxFiles) == 0 {
// 		log.Println("no idxs to load, initialised")
// 		return
// 	}

// 	var wg sync.WaitGroup

// 	// make loading concurrent
// 	for _, idxDocFile := range idxFiles {
// 		tempIdx := NewIndex(idxDocFile.Name(), config.CaseSensitivity)

// 		if err != nil {
// 			log.Fatalln("could not load index from doc list, err: ", err.Error())
// 		}

// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			err = tempIdx.LoadDocumentsIntoIndex()

// 			if err != nil {
// 				log.Fatalln("error loading documents into index ", tempIdx.Name)
// 			}
// 		}()
// 		tempIdx.mc = s.mc
// 		s.ActiveIndices[idxDocFile.Name()] = tempIdx
// 	}

// 	log.Println("idxs initialised :)")

// }
