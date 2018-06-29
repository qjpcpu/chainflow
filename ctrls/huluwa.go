package ctrls

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Huluwa struct {
	Name   string                 `json:"name"`
	Image  string                 `json:"image"`
	Des    string                 `json:"description"`
	Attr   map[string]interface{} `json:"attributes"`
	ExtUrl string                 `json:"external_url"`
	Bg     string                 `json:"background_color"`
}

var hulubrothers = map[string]Huluwa{
	"1": Huluwa{
		Name:  "大娃",
		Image: "https://ss3.bdstatic.com/70cFv8Sh_Q1YnxGkpoWK1HF6hhy/it/u=1780963879,2183154338&fm=27&gp=0.jpg",
		Des:   "大娃",
		Bg:    "FBEAEE",
	},
}

func GetHuluwa(c *gin.Context) {
	hlw := hulubrothers[c.Param(":id")]
	c.JSON(http.StatusOK, hlw)
}
