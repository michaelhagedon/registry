package controllers

import (
	"github.com/APTrust/registry/helpers"
	"github.com/gin-gonic/gin"
)

func ErrorShow(c *gin.Context) {
	c.HTML(200, "error/show.html", helpers.TemplateVars(c))
}
