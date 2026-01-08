package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"gocene/config"
	"gocene/internal/store"
	"gocene/internal/utils"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/raft"
	"github.com/minio/minio-go/v7"
)

// all service functions here

type Service struct {
	st          *store.Store
	minioClient *minio.Client
}

func NewService() *Service {

	s := &Service{}
	var err error
	s.minioClient, err = utils.CreateMinioClient()
	if err != nil {
		log.Fatalln("could not connect to minio server")
	}

	s.st = store.New(s.minioClient)
	return s
}

// Creates a new index.
func (s *Service) CreateIndex(inp CreateIndexInput) (res *CreateIndexResult, err error) {

	err = s.st.CreateIndex(inp.Name, inp.CaseSensitivity)
	if err != nil {
		if err == store.ErrNotLeader {
			return s.ForwardCreateIndexToLeader(inp)
		}
		return nil, err
	}

	return &CreateIndexResult{
		Success: err == nil,
	}, err
}

// Gets the list of all active indices on the service.
func (s *Service) GetIndices() (res *GetIndicesResult, err error) {

	res = &GetIndicesResult{}
	for idxName := range s.st.ActiveIndices {
		res.IndicesList = append(res.IndicesList, idxName)
	}

	return
}

// Adds Document to specified index if node is leader. Else, forwards it to the leader.
func (s *Service) AddDocument(idxName string, inp AddDocumentInput) (res *AddDocumentResult, err error) {
	log.Println("inside service AddDocument()")

	docId, err := s.st.AddDocument(idxName, inp.Data)
	if err != nil {
		if err == store.ErrNotLeader {
			// forward request to leader.
			return s.ForwardAddDocumentToLeader(idxName, inp)
		}
	}

	return &AddDocumentResult{
		DocID:   docId,
		Success: err == nil,
	}, err

}

// does nothing
func (s *Service) GetDocument(idxName string, inp GetDocumentInput) (res *GetDocumentResult, err error) {

	log.Println("inside service GetDocument()")
	return
}

// Performs full text search on specified index with given terms.
func (s *Service) SearchFullText(idxName string, inp SearchInput) (res *SearchResult, err error) {

	log.Println("inside service SearchFullText()")

	var idx *store.Index
	var ok bool

	if idx, ok = s.st.ActiveIndices[idxName]; !ok {
		log.Println(store.ErrIdxDoesNotExist.Error())
		// index does not exist
		return nil, store.ErrIdxDoesNotExist
	}

	terms := store.GetTermsFromPhrase(inp.SearchField, inp.SearchPhrase)

	rankedDocs, err := idx.SearchFullText(terms)
	if err != nil {
		return nil, err
	}

	res = &SearchResult{
		Results: rankedDocs,
		Count:   len(rankedDocs),
	}

	return
}

// Add the requesting node to the cluster if leader, and forwards the request to the leader if not.
func (s *Service) Join(inp JoinInput) (err error) {
	if s.st.Raft.State() != raft.Leader {
		return s.ForwardJoinToLeader(inp)
	}

	return s.st.Join(inp.NodeID, inp.Address)
}

// Returns the Raft status of the node.
func (s *Service) Status() (res StatusResult, err error) {
	t, err := s.st.Status()
	return StatusResult(t), err
}

// service functions that do the forwarding to leader

// Forwards the add doc request to the leader.
func (s *Service) ForwardAddDocumentToLeader(idxName string, inp AddDocumentInput) (res *AddDocumentResult, err error) {
	leaderAddr, _ := s.st.Raft.LeaderWithID()

	// leader does exist, forward request to it
	if leaderAddr != "" {
		payload := AddDocumentInput{
			Data: inp.Data,
		}
		data, lerr := json.Marshal(payload)
		if lerr != nil {
			log.Println("could not marshal json in fwd to leader add doc, err: ", lerr.Error())
			return nil, lerr
		}

		resp, lerr := http.Post(
			string(leaderAddr)+config.EndpointsMap[config.AddDocumentAPI],
			"application/json",
			bytes.NewBuffer(data),
		)

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var finalRes AddDocumentResult
		lerr = json.Unmarshal(body, &finalRes)
		if lerr != nil {
			log.Println("could not unmarshal add doc result from leader, err: ", lerr.Error())
			return nil, lerr
		}

		return &finalRes, nil
	}

	return nil, errors.New("no leader detected")
}

func (s *Service) ForwardCreateIndexToLeader(inp CreateIndexInput) (res *CreateIndexResult, err error) {

	leaderAddr, _ := s.st.Raft.LeaderWithID()

	// leader does exist, forward request to it
	if leaderAddr != "" {

		data, lerr := json.Marshal(inp)
		if lerr != nil {
			log.Println("could not marshal json in fwd to leader create index, err: ", lerr.Error())
			return nil, lerr
		}

		resp, lerr := http.Post(
			string(leaderAddr)+config.EndpointsMap[config.CreateIndexAPI],
			"application/json",
			bytes.NewBuffer(data),
		)

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var finalRes CreateIndexResult
		lerr = json.Unmarshal(body, &finalRes)
		if lerr != nil {
			log.Println("could not unmarshal create index result from leader, err: ", lerr.Error())
			return nil, lerr
		}

		return &finalRes, nil
	}

	return nil, errors.New("no leader detected")
}

func (s *Service) ForwardJoinToLeader(inp JoinInput) (err error) {

	return
}
