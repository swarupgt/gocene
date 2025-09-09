package api

import (
	"gocene/internal/store"
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
		if err == ErrIdxNameExists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "index name already exists"})
			return http.StatusBadRequest
		}
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
func (c *Controller) AddDocument(ctx *gin.Context) (status int) {

	log.Println("inside cont AddDocument()")

	idx := ctx.Param("idx_name")
	if idx == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "index not specified"})
		return http.StatusBadRequest
	}

	var inp AddDocumentInput
	if err := ctx.BindJSON(&inp); err != nil {
		//bind failed, return 400
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect input structure"})
		return http.StatusBadRequest
	}

	res, err := c.serv.AddDocument(idx, inp)
	if err != nil {
		log.Println("Error adding document: ", err.Error())
		if err == ErrIdxDoesNotExist {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "index specified does not exist"})
			return http.StatusBadRequest
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return http.StatusInternalServerError
		}
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK
}

// modify doc

// get doc (all docs too)
func (c *Controller) GetDocument(ctx *gin.Context) (status int) {
	log.Println("inside cont GetDocument()")

	idx := ctx.Param("idx_name")
	if idx == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "index not specified"})
		return http.StatusBadRequest
	}

	var inp GetDocumentInput
	if err := ctx.BindJSON(&inp); err != nil {
		//bind failed, return 400
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect input structure"})
		return http.StatusBadRequest
	}

	res, err := c.serv.GetDocument(idx, inp)
	if err != nil {
		if err == ErrIdxDoesNotExist {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "index specified does not exist"})
			return http.StatusBadRequest
		} else if err == store.ErrDocumentNotFound {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "document specified does not exist"})
			return http.StatusBadRequest
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return http.StatusInternalServerError
		}
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK
}

// search full text
func (c *Controller) SearchFullText(ctx *gin.Context) (status int) {
	log.Println("inside cont SearchFullText()")

	idx := ctx.Param("idx_name")
	if idx == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "index not specified"})
		return http.StatusBadRequest
	}

	var inp SearchInput
	if err := ctx.BindJSON(&inp); err != nil {
		//bind failed, return 400
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect input structure"})
		return http.StatusBadRequest
	}

	res, err := c.serv.SearchFullText(idx, inp)
	if err != nil {
		if err == ErrIdxDoesNotExist {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "index specified does not exist"})
			return http.StatusBadRequest
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return http.StatusInternalServerError
		}
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK
}
