package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// all api controller functions here

type Controller struct {
	serv Service
}

func NewController() (cont *Controller) {
	return &Controller{}
}

func (c *Controller) CreateIndex(ctx *gin.Context) (status int) {
	//bind query
	var inp CreateIndexInput
	if err := ctx.BindJSON(&inp); err != nil {
		//bind failed, return 400
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect input structure"})
		return http.StatusBadRequest
	}

	res, err := c.serv.CreateIndex(inp)
	if err != nil {
		//server error, log
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return http.StatusInternalServerError
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK

}

func (c *Controller) GetIndices(ctx *gin.Context) (status int) {

	res, err := c.serv.GetIndices()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return http.StatusInternalServerError
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK
}

// add doc
// func (c *Controller) AddDocument(ctx *gin.Context) (status int) {

// }

// modify doc

// get doc (all docs too)

//search full text
