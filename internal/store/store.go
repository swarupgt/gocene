package store

import (
	"encoding/gob"
	"gocene/config"
	"log"
	"os"
)

// Make it concurrent
func SaveIndexToPersistentMemory(idx *Index) (err error) {
	f, err := os.Create(config.IndexFileDirectory + idx.Name + config.IndexFileExtension)
	if err != nil {
		log.Fatalln("could not create index file, err: ", err)
		return err
	}

	defer f.Close()

	// err = binary.Write(f, binary.LittleEndian, idx)

	enc := gob.NewEncoder(f)
	enc.Encode(idx)

	// if err != nil {
	// 	log.Fatalln("could not write index to disk, err: ", err.Error())
	// }

	log.Println("index saved")

	return nil
}

func LoadIndexFromPersistentMemory(idxName string) (idx *Index, err error) {

	f, err := os.Open(config.IndexFileDirectory + idx.Name + config.IndexFileExtension)

	if err != nil {
		log.Fatalln("could not open index file, err: ", err)
		return nil, err
	}
	defer f.Close()

	idx = &Index{}

	dec := gob.NewDecoder(f)
	dec.Decode(&idx)

	return idx, nil
}
