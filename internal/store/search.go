package store

import (
	"encoding/json"
	"gocene/internal/utils"
	"log"
	"sort"
	"sync"
)

type RankedResultDoc struct {
	Score int             `json:"score"`
	Data  json.RawMessage `json:"data"`
}

type RankedDocData struct {
	Score     int
	ParentSeg *Segment

	DocID int
}

// Searches all the segments in the index concurrently.
// Todo: Limit goroutine spawning
func (idx *Index) SearchFullText(terms []Term) (results []RankedResultDoc, err error) {

	var res []RankedDocData

	log.Println("inside store SearchFullText()")

	resChan := make(chan []RankedDocData, len(idx.Segments)+1)
	resErrs := make(chan error, len(idx.Segments)+1)
	var wg sync.WaitGroup

	// search segments concurrently
	for _, seg := range idx.Segments {
		// log.Println("searching immutable segment")
		wg.Add(1)
		go func(s *Segment) {
			defer wg.Done()
			sRes, err := s.SearchFullText(terms)

			var resTemp []RankedDocData
			for _, r := range sRes {
				resTemp = append(resTemp, RankedDocData{
					Score:     r.Score,
					ParentSeg: s,
					DocID:     r.DocID,
				})
			}

			resChan <- resTemp
			resErrs <- err

		}(seg)
	}

	// search active segment too, needs a lock
	wg.Add(1)
	go func(as *ActiveSegment) {
		// log.Println("searching active segment")
		defer wg.Done()

		as.Mutex.RLock()
		defer as.Mutex.RUnlock()

		asRes, err := as.Seg.SearchFullText(terms)

		var resTemp []RankedDocData
		for _, r := range asRes {
			resTemp = append(resTemp, RankedDocData{
				Score:     r.Score,
				ParentSeg: as.Seg,
				DocID:     r.DocID,
			})
		}

		resChan <- resTemp
		resErrs <- err
	}(&idx.As)

	go func() {
		wg.Wait()
		close(resChan)
		close(resErrs)
	}()

	// do merge algo for top k later, since all segments return sorted
	// aggregate
	for temp := range resChan {
		res = append(res, temp...)
	}

	for tempErr := range resErrs {
		if tempErr != nil {
			log.Println("error searching an index")
		}
	}

	// score results
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].Score > res[j].Score
	})

	// get the json data for each scored and ranked doc
	for _, iter := range res {
		jsonStr, err := utils.GetDocumentFromMinio(idx.mc, iter.DocID, idx.Name)
		// jsonStr, err := iter.ParentSeg.GetDocument(iter.DocID)
		if err != nil {
			log.Println("error getting document: ", err.Error())
		}
		// fmt.Println("json str: ", jsonStr)
		results = append(results, RankedResultDoc{
			Score: iter.Score,
			Data:  json.RawMessage(jsonStr),
		})
	}

	// fmt.Println("RESULT OF SEARCH FINAL - ", results)

	return
}
