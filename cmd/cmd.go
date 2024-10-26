package cmd

import (
	"gocene/internal/api"
	"gocene/internal/store"
	"log"
)

func Begin() {
	log.Println("beginning service")

	store.Init()

	router := api.GetRouter()
	router.SetEndpoints()
	router.StartRouter()

	// idx := store.NewIndex("test", false)

	// doc, err := utils.CreateDocumentFromJSON(`
	// {
	// 	"title":"Rise of the Beast",
	// 	"article":"The sun set over the horizon, painting the sky with vibrant hues of orange and pink."
	// }`)

	// if err != nil {
	// 	log.Println("err: ", err)
	// }

	// err = idx.AddDocument(doc)
	// if err != nil {
	// 	log.Println("err: ", err)
	// }

	// doc2, err := utils.CreateDocumentFromJSON(`
	// {
	// 	"title":"Taken",
	// 	"article":"A gentle breeze rustled the leaves, carrying the sweet scent of blooming flowers across the meadow, filling the air with fragrance."
	// }`)

	// if err != nil {
	// 	log.Println("err: ", err)
	// }

	// err = idx.AddDocument(doc2)
	// if err != nil {
	// 	log.Println("err: ", err)
	// }

	// docs := idx.GetAllDocuments()
	// // count := idx.GetDocumentCount()

	// fmt.Println("Docs: ", docs)
	// fmt.Println("Doc count: ", idx.GetDocumentCount())

	// // terms, counts1 := idx.GetTermsAndFreqFromDocNo(0)

	// // fmt.Println("terms and their counts: ")

	// // for i, term := range terms {
	// // 	fmt.Println(term.Value, counts1[i])
	// // }

	// // res, err := idx.SearchTerm(store.NewTerm("article", "the"))

	// res, err := idx.SearchFullText([]store.Term{
	// 	store.NewTerm("article", "the"),
	// 	store.NewTerm("article", "sweet"),
	// })

	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	for _, iter := range res {
	// 		log.Println(iter.Score, iter.Document.ID)
	// 	}
	// }

}
