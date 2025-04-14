package routes

import (
	"github.com/gin-gonic/gin"
)

type RoutesImpl struct {
	engine *gin.Engine
}

func NewRoute(r *gin.Engine) RoutesImpl {
	routes := new(RoutesImpl)
	routes.engine = r
	return *routes
}

func (i *RoutesImpl) Setup() {

}
