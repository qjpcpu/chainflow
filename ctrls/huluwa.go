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
		Image: "http://p1.qhimgs4.com/t01fdf4690dd8404ecb.webp",
		Des:   "翻天掀地，力大无穷。他是七兄弟中的老大哥，生来就是一个大力士，身体可以任意变大或缩小",
		Attr:  map[string]interface{}{"property": "力量"},
		Bg:    "FBEAEE",
	},
	"2": Huluwa{
		Name:  "二娃",
		Image: "http://p2.qhimgs4.com/t01b9d2a9f09eb73485.webp",
		Des:   "慧眼千里，耳闻八方。橙娃天生便拥有一双千里眼和一对顺风耳，妖怪的一切秘密都瞒不住他",
		Attr:  map[string]interface{}{"property": "千里眼"},
		Bg:    "FBEAEE",
	},
	"3": Huluwa{
		Name:  "三娃",
		Image: "http://p1.qhimgs4.com/t010a677f6a10f70283.webp",
		Des:   "铜头铁臂，刀枪不入。黄娃三弟是个拥有钢筋铁骨的神娃，刀枪箭炮对他丝毫无伤",
		Attr:  map[string]interface{}{"property": "铜头铁臂"},
		Bg:    "FBEAEE",
	},
	"4": Huluwa{
		Name:  "四娃",
		Image: "http://p1.qhimgs4.com/t0174bbd1b130f95d65.webp",
		Des:   "炉火纯青，刚阳烈焰。绿娃四弟乃天界火神下凡，可任意吞吐烈火",
		Attr:  map[string]interface{}{"property": "火"},
		Bg:    "FBEAEE",
	},
	"5": Huluwa{
		Name:  "五娃",
		Image: "http://p2.qhimgs4.com/t01dc166043854262e5.webp",
		Des:   "惊涛骇浪，气吞山河。青娃五弟乃江河水神转世，可任意吞吐江河之水，难怪吞饮百坛烈酒而不醉",
		Attr:  map[string]interface{}{"property": "水"},
		Bg:    "FBEAEE",
	},
	"6": Huluwa{
		Name:  "六娃",
		Image: "http://p1.qhimgs4.com/t012c2f570d03e22935.webp",
		Des:   "来无影，去无踪。蓝娃六弟是七兄弟中最灵敏，最聪明的。他的隐身法令妖怪束手无策，并大闹妖洞，智救五哥，夺走了宝贝如意",
		Attr:  map[string]interface{}{"property": "隐身"},
		Bg:    "FBEAEE",
	},
	"7": Huluwa{
		Name:  "七娃",
		Image: "http://p0.qhimgs4.com/t01a60b2bda30d1b00b.webp",
		Des:   "镇妖之宝，本领无穷。紫娃七弟自身没多大本事，但他有个宝葫芦，是太上老君修炼仙丹用的紫金神葫",
		Attr:  map[string]interface{}{"property": "葫芦"},
		Bg:    "FBEAEE",
	},
}

func GetHuluwa(c *gin.Context) {
	hlw := hulubrothers[c.Param("id")]
	c.JSON(http.StatusOK, hlw)
}
