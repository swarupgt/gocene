package api

import (
	"gocene/config"

	"github.com/gin-gonic/gin"
)

type Router struct {
	R    *gin.Engine
	Cont *Controller
}

// gin router and endpoint init here

func GetRouter() *Router {

	// add configs and log options later
	return &Router{
		R:    gin.Default(),
		Cont: NewController(),
	}
}

func (r *Router) StartRouter() error {
	return r.R.Run(config.Port)
}

// set endpoints()
func (router *Router) SetEndpoints() {

	// add the other APIs later
	for apiId, endpoint := range config.EndpointsMap {

		if apiId == config.CreateIndexAPI {
			router.R.POST(endpoint, func(ctx *gin.Context) {
				// use return types later for logs
				router.Cont.CreateIndex(ctx)
			})
		} else if apiId == config.GetIndicesAPI {
			router.R.GET(endpoint, func(ctx *gin.Context) {
				router.Cont.GetIndices(ctx)
			})
		} else if apiId == config.AddDocumentAPI {
			router.R.POST(endpoint, func(ctx *gin.Context) {
				router.Cont.AddDocument(ctx)
			})
		}
	}
}
