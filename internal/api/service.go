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

	// add yourself as a peer if alone
	// err = s.st.AddNode(config.RaftAddress, config.RaftSelfHTTPAddress)
	// if err != nil {
	// 	log.Fatalln("could not add myself to raft log, err: ", err.Error())
	// }

	return s
}

// Creates a new index.
func (s *Service) CreateIndex(inp CreateIndexInput) (res *CreateIndexResult, err error) {

	err = s.st.CreateIndex(inp.Name, inp.CaseSensitivity)
	if err != nil {
		if err == store.ErrNotLeader {
			return s.ForwardCreateIndexToLeader(inp)
		}
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
func (s *Service) Join(inp JoinInput) (res *JoinResult, err error) {
	if s.st.Raft.State() != raft.Leader {
		log.Println("not leader, forwarding...")
		return s.ForwardJoinToLeader(inp)
	}

	err = s.st.Join(inp.NodeID, inp.Address, inp.HTTPAddress)
	if err != nil {
		return nil, err
	}

	return &JoinResult{
		LeaderHTTPAddress: config.RaftSelfHTTPAddress,
	}, nil
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
	leaderHTTPAddr := s.st.PeerHTTP[string(leaderAddr)]

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
			"http://"+string(leaderHTTPAddr)+"/"+idxName+"/add_document",
			"application/json",
			bytes.NewBuffer(data),
		)

		if lerr != nil {
			log.Println("could not forward add doc to leader, err: ", lerr.Error())
			return nil, lerr
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var finalRes AddDocumentResult
		lerr = json.Unmarshal(body, &finalRes)
		if lerr != nil {
			log.Println("could not unmarshal add doc result from leader, err: ", lerr.Error())
			return nil, lerr
		}

		if resp.StatusCode == http.StatusBadRequest && finalRes.Error == "index specified does not exist" {
			return &finalRes, store.ErrIdxDoesNotExist
		} else if resp.StatusCode == http.StatusInternalServerError {
			return &finalRes, errors.New("something went wrong")
		}

		return &finalRes, nil
	}

	return nil, errors.New("no leader detected")
}

func (s *Service) ForwardCreateIndexToLeader(inp CreateIndexInput) (res *CreateIndexResult, err error) {

	leaderAddr, _ := s.st.Raft.LeaderWithID()
	leaderHTTPAddr := s.st.PeerHTTP[string(leaderAddr)]

	// leader does exist, forward request to it
	if leaderAddr != "" {

		data, lerr := json.Marshal(inp)
		if lerr != nil {
			log.Println("could not marshal json in fwd to leader create index, err: ", lerr.Error())
			return nil, lerr
		}

		resp, lerr := http.Post(
			"http://"+string(leaderHTTPAddr)+config.EndpointsMap[config.CreateIndexAPI],
			"application/json",
			bytes.NewBuffer(data),
		)

		if lerr != nil {
			log.Println("could not forward create index to leader, err: ", lerr.Error())
			return nil, lerr
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var finalRes CreateIndexResult
		lerr = json.Unmarshal(body, &finalRes)
		if lerr != nil {
			log.Println("could not unmarshal create index result from leader, err: ", lerr.Error())
			return nil, lerr
		}

		if resp.StatusCode == http.StatusBadRequest && finalRes.Error == "index name already exists" {
			return &finalRes, store.ErrIdxNameExists
		} else if resp.StatusCode == http.StatusInternalServerError {
			return &finalRes, errors.New("something went wrong")
		}

		return &finalRes, nil
	}

	return nil, errors.New("no leader detected")
}

func (s *Service) ForwardJoinToLeader(inp JoinInput) (res *JoinResult, err error) {

	leaderAddr, _ := s.st.Raft.LeaderWithID()
	leaderHTTPAddr := s.st.PeerHTTP[string(leaderAddr)]

	// leader does exist, forward request to it
	if leaderAddr != "" {

		data, lerr := json.Marshal(inp)
		if lerr != nil {
			log.Println("could not marshal json in fwd to leader create index, err: ", lerr.Error())
			return nil, lerr
		}

		resp, lerr := http.Post(
			"http://"+string(leaderHTTPAddr)+config.EndpointsMap[config.JoinAPI],
			"application/json",
			bytes.NewBuffer(data),
		)

		if lerr != nil {
			log.Println("could not forward join to leader, err: ", lerr.Error())
			return nil, lerr
		}

		defer resp.Body.Close()
		var finalRes JoinResult
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(data, &finalRes); err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("could not join leader")
		}

		return &finalRes, nil
	}

	return nil, errors.New("no leader detected")
}
