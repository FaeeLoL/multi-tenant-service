package controllers

import (
	"fmt"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TenantsController struct {
	ControllerBase
}

func (t TenantsController) CreateTenant(c *gin.Context) {
	var newTenant models.Tenant
	err := c.Bind(&newTenant)
	if err != nil {
		t.JsonFail(c, http.StatusBadRequest, "Bad request")
	}
	fmt.Printf("new tenant data: %+v\n", newTenant)
	
}

