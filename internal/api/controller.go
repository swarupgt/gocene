package api

import (
	"gocene/internal/store"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// all HTTP API controller functions here

type Controller struct {
	serv *Service
}

func NewController() (cont *Controller) {
	return &Controller{
		serv: NewService(),
	}
}

// Create Index HTTP func
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
		if err == store.ErrIdxNameExists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "index name already exists"})
			return http.StatusBadRequest
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return http.StatusInternalServerError
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK

}

// Get Indices HTTP
func (c *Controller) GetIndices(ctx *gin.Context) (status int) {

	res, err := c.serv.GetIndices()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return http.StatusInternalServerError
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK
}

// Add Document HTTP
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
		if err == store.ErrIdxDoesNotExist {
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

// Get Document HTTP
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
		if err == store.ErrIdxDoesNotExist {
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

// Search Full Text HTTP
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
		if err == store.ErrIdxDoesNotExist {
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

func (c *Controller) Join(ctx *gin.Context) (status int) {

	var inp JoinInput
	if err := ctx.BindJSON(&inp); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect join structure"})
		return http.StatusBadRequest
	}

	res, err := c.serv.Join(inp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return http.StatusInternalServerError
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK
}

func (c *Controller) Status(ctx *gin.Context) (status int) {

	res, err := c.serv.Status()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return http.StatusInternalServerError
	}

	ctx.JSON(http.StatusOK, res)
	return http.StatusOK
}
