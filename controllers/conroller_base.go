package controllers

import "github.com/gin-gonic/gin"

type ControllerBase struct {
}

func (basic *ControllerBase) JsonSuccess(c *gin.Context, status int, obj interface{}) {
	c.JSON(status, obj)
}

func (basic *ControllerBase) JsonFail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"message": message,
		},
	})
}
