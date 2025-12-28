package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

/*
Segment Metadata structure
{
	name: <name>,
	term_dict: {map}
}

*/

// Flush segment metadata to disk.
// func FlushSegmentMetadata(s *Segment) (err error) {

// 	if os.MkdirAll(config.IndexMetadataDirectory+"/"+s.ParentIdxName, os.ModePerm) != nil {
// 		log.Fatalln("error creating index metadata directory, err: ", err.Error())
// 		return err
// 	}

// 	f, err := os.Create(config.IndexMetadataDirectory + "/" + s.ParentIdxName + "/" + s.Name)
// 	if err != nil {
// 		log.Fatalln("could not create segment metadata file, err: ", err)
// 		return err
// 	}

// 	defer f.Close()

// 	enc := gob.NewEncoder(f)
// 	enc.Encode(s)

// 	log.Println("segment metadata saved for ", s.Name)

// 	return nil
// }

// // Load segment metadata from disk into a Segment struct.
// func LoadSegmentMetadata(idxName, segName string) (s *Segment, err error) {

// 	log.Printf("loading segment metadata for idx: %s and seg: %s\n", idxName, segName)

// 	f, err := os.Open(config.IndexMetadataDirectory + "/" + idxName + "/" + segName)
// 	if err != nil {
// 		log.Fatalf("could not open %s idx, %s seg, err: %s", idxName, segName, err.Error())
// 		return nil, err
// 	}

// 	defer f.Close()

// 	dec := gob.NewDecoder(f)
// 	var tempSeg Segment
// 	dec.Decode(&tempSeg)

// 	fmt.Println("load segment data stuff", tempSeg.Name, tempSeg.PostingsMap)

// 	return &tempSeg, nil
// }

// Only parses JSONs of single level and type string at the moment, see how lucene does it for any JSON
func CreateDocumentFromJSON(jsonString string) (doc *Document, err error) {
	// log.Println("inside CreateDocumentFromJSON()")

	// fmt.Println("json doc in CREATEDOCJSON:", jsonString)

	var obj map[string]interface{}

	err = json.Unmarshal([]byte(jsonString), &obj)
	if err != nil {
		// log.Println("ERR IN UNMARSHALLING: ", jsonString)
		return nil, err
	}

	// fmt.Println("obj: ", obj)

	var id int = 0

	doc = NewDocument()

	for key := range obj {
		field := Field{
			ID:              id,
			Name:            key,
			Type:            StringField,
			TokenizerString: " ",
			Value:           fmt.Sprint(obj[key]),
		}

		doc.AddField(field)
		id++
	}

	doc.DocMap = obj

	// fmt.Println("doc after creating:", doc.DocMap["book_name"], doc.DocMap["content"])

	return doc, nil
}

// Parse the segment file and return the list of docs
func ParseSegmentDataFile(segFilePath string) (docStrs []string, err error) {

	file, err := os.Open(segFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var temp string
	var buff []byte = make([]byte, 1024)

	for {
		n, err := file.Read(buff)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n == 0 {
			break
		}
		temp = temp + string(buff[:n])
	}

	var singleDocStr string

	for _, c := range temp {
		singleDocStr = singleDocStr + string(c)
		if c == '}' {
			docStrs = append(docStrs, singleDocStr)
			singleDocStr = ""
		}
	}

	// fmt.Println("DOCS PARSED: ", docStrs)

	return
}
