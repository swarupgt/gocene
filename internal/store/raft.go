package store

import (
	"encoding/json"
	"fmt"
	cfg "gocene/config"
	"gocene/internal/utils"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// This file contains all the raft related functionality for Store.

const SnapshotCount = 2

// fsmSnapshot stores the entire Store metadata in one object for Raft snapshotting.
type fsmSnapshot struct {
	ActiveIndices []IndexMetadata `json:"active_indices"`
}

// for Raft index snapshot loading
type IndexMetadata struct {
	Name              string            `json:"name"`
	SegmentList       []SegmentMetadata `json:"segment_list"`
	NextDocID         int               `json:"next_doc_id"`
	SegCount          int               `json:"seg_count"`
	CaseSensitivity   bool              `json:"case_sensitivity"`
	ActiveSegmentName string            `json:"active_segment_name"`
}

// for Raft Segment snapshot loading
type SegmentMetadata struct {
	IsActive      bool                `json:"is_active"`
	Name          string              `json:"name"`
	TermDict      TermDictionary      `json:"term_dict"`
	ParentIdxName string              `json:"parent_idx_name"`
	PostingsMap   map[int]docPosition `json:"postingsMap"`
	DocCount      int                 `json:"doc_count"`
	ByteSize      int                 `json:"byte_size"`
}

// Instantiates the raft configs for the node, and bootstraps if it's the first node to start
func (s *Store) Open() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(cfg.RaftId)

	addr, err := net.ResolveTCPAddr("tcp", s.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	os.MkdirAll(cfg.RaftDirectory, os.ModePerm)
	dirs, err := os.ReadDir(cfg.RaftDirectory)
	if err != nil {
		log.Fatalln("could not read raft dir, err: ", err.Error())
	}

	isEmptyRaftDirAtStartup := len(dirs) == 0

	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, SnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	var logStore raft.LogStore
	var stableStore raft.StableStore

	boltDB, err := raftboltdb.New(raftboltdb.Options{
		Path: filepath.Join(s.RaftDir, "raft.db"),
	})
	if err != nil {
		return fmt.Errorf("new bbolt store: %s", err)
	}
	logStore = boltDB
	stableStore = boltDB

	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.Raft = ra

	// check if no raft data on disk, only then bootstrap
	if cfg.RaftBootstrap && isEmptyRaftDirAtStartup {
		log.Println("leader, bootstrapping")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)

		// wait till you become leader and add yourself as a peer

		go func() {
			timeout := time.After(8 * time.Second)
			tick := time.NewTicker(200 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-timeout:
					log.Println("timed out waiting to become leader to add self to peer list")
					return
				case <-tick.C:
					if s.Raft.State() == raft.Leader {
						if err := s.AddNode(cfg.RaftAddress, cfg.RaftSelfHTTPAddress); err != nil {
							log.Println("could not add self into PeerHTTP via AddNode:", err)
						} else {
							log.Println("added self to PeerHTTP via AddNode")
						}
						return
					}
				}
			}
		}()

	} else if cfg.RaftJoinAddress != "" {
		// join cluster at join address
		err = JoinLeaderAsFollower(cfg.RaftSelfHTTPAddress, cfg.RaftJoinAddress, cfg.RaftAddress, cfg.RaftId)
		if err != nil {
			log.Fatalln("could not join specified leader, err: ", err)
		}
	}

	return nil
}

// Raft Apply implementation
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c Command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.CmdId {
	case CmdAddDocument:
		return f.ApplyAddDocument(c.IdxName, c.Param)
	case CmdCreateIndex:
		return f.ApplyCreateIndex(c.IdxName, c.Param)
	case CmdAddNode:
		return f.ApplyAddNode(c.NodeAddress, c.NodeHTTPAddress)
	default:
		panic(fmt.Sprintf("unrecognized command op ID: %d", c.CmdId))
	}
}

// Apply adding document to the FSM store
func (f *fsm) ApplyAddDocument(idxName string, docID int) interface{} {

	if _, ok := f.ActiveIndices[idxName]; !ok {
		return ErrIdxDoesNotExist
	}

	// fetch doc from S3 bucket
	docStr, err := utils.GetDocumentFromMinio(f.mc, docID, idxName)
	if err != nil {
		log.Println("could not fetch doc from minio")
		return err
	}

	doc, err := CreateDocumentFromJSON(docStr)
	if err != nil {
		return err
	}

	id, err := f.ActiveIndices[idxName].AddDocument(doc)
	if err != nil {
		return err
	}

	return id
}

// Applying creating an index to the FSM Store
func (f *fsm) ApplyCreateIndex(idxName string, cs int) error {

	if _, ok := f.ActiveIndices[idxName]; ok {
		return ErrIdxNameExists
	}

	f.ActiveIndices[idxName] = NewIndex(idxName, cs == 1, f.mc)
	return nil
}

func (f *fsm) ApplyAddNode(nodeAddr, nodeHTTPAddr string) error {
	log.Printf("-----applying add node of %s with http addr: %s\n", nodeAddr, nodeHTTPAddr)
	f.PeerHTTP[nodeAddr] = nodeHTTPAddr
	return nil
}

// Raft FSM Snapshot implementation
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {

	f.mu.Lock()
	defer f.mu.Unlock()

	var fSnap fsmSnapshot

	// get metadata for every segment of every index
	for _, idx := range f.ActiveIndices {
		idxMd := IndexMetadata{
			Name:              idx.Name,
			NextDocID:         idx.NextDocID,
			SegCount:          idx.SegCount,
			CaseSensitivity:   idx.CaseSensitivity,
			ActiveSegmentName: idx.As.Seg.Name,
		}

		// add immutable segments metadata
		for _, seg := range idx.Segments {
			segMd := SegmentMetadata{
				IsActive:      false,
				Name:          seg.Name,
				TermDict:      seg.TermDict,
				ParentIdxName: idx.Name,
				// PostingsMap:   seg.PostingsMap,
				DocCount: seg.DocCount,
				ByteSize: seg.ByteSize,
			}

			idxMd.SegmentList = append(idxMd.SegmentList, segMd)
		}

		idxMd.SegmentList = append(idxMd.SegmentList, SegmentMetadata{
			IsActive:      true,
			Name:          idx.As.Seg.Name,
			TermDict:      idx.As.Seg.TermDict,
			ParentIdxName: idx.Name,
			// PostingsMap:   idx.As.Seg.PostingsMap,
			DocCount: idx.As.DocCount,
			ByteSize: idx.As.Seg.ByteSize,
		})

		fSnap.ActiveIndices = append(fSnap.ActiveIndices, idxMd)
	}

	return fSnap, nil
}

// Raft FSM Restore implementation,
// restore FSM Store to a previous state
func (f *fsm) Restore(rc io.ReadCloser) error {

	var fSnap fsmSnapshot
	if err := json.NewDecoder(rc).Decode(&fSnap); err != nil {
		return err
	}

	for _, idxMd := range fSnap.ActiveIndices {
		tempIdx := NewIndex(idxMd.Name, idxMd.CaseSensitivity, f.mc)
		for _, segMd := range idxMd.SegmentList {

			// if immutable segments, append to seg list
			if !segMd.IsActive {
				tempSeg, err := NewSegment(segMd.Name, tempIdx)
				if err != nil {
					log.Println("could not create new segment while restoring snapshot, err: ", err.Error())
					return err
				}
				tempSeg.TermDict = segMd.TermDict
				// tempSeg.PostingsMap = segMd.PostingsMap
				tempSeg.DocCount = segMd.DocCount
				tempSeg.ByteSize = segMd.ByteSize
			} else {
				tempAsSeg, err := NewActiveSegment(segMd.Name, tempIdx)
				if err != nil {
					log.Println("could not create active segment while restoring snapshot, err: ", err.Error())
					return err
				}
				tempAsSeg.Seg.TermDict = segMd.TermDict
				// tempAsSeg.Seg.PostingsMap = segMd.PostingsMap
				tempAsSeg.Seg.DocCount = segMd.DocCount
				tempAsSeg.DocCount = segMd.DocCount
				tempAsSeg.Seg.ByteSize = segMd.ByteSize
			}
		}
	}

	var err error
	f.mc, err = utils.CreateMinioClient()
	if err != nil {
		log.Fatalln("could not create minio client, err: ", err.Error())
		return err
	}

	return nil
}

func (fs fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(fs)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (fs fsmSnapshot) Release() {}
