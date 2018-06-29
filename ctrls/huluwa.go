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
		Image: "https://gss0.bdstatic.com/94o3dSag_xI4khGkpoWK1HF6hhy/baike/c0%3Dbaike80%2C5%2C5%2C80%2C26/sign=cc5042bc5ddf8db1a8237436684ab631/728da9773912b31b20618e0d8518367adab4e181.jpg",
		Des:   "翻天掀地，力大无穷。他是七兄弟中的老大哥，生来就是一个大力士，身体可以任意变大或缩小",
		Attr:  map[string]interface{}{"property": "力量"},
		Bg:    "FBEAEE",
	},
	"2": Huluwa{
		Name:  "二娃",
		Image: "https://gss2.bdstatic.com/-fo3dSag_xI4khGkpoWK1HF6hhy/baike/c0%3Dbaike80%2C5%2C5%2C80%2C26/sign=636da23e2b34349b600b66d7a8837eab/7e3e6709c93d70cf7d87e929fbdcd100baa12b90.jpg",
		Des:   "慧眼千里，耳闻八方。橙娃天生便拥有一双千里眼和一对顺风耳，妖怪的一切秘密都瞒不住他",
		Attr:  map[string]interface{}{"property": "千里眼"},
		Bg:    "FBEAEE",
	},
	"3": Huluwa{
		Name:  "三娃",
		Image: "https://gss1.bdstatic.com/-vo3dSag_xI4khGkpoWK1HF6hhy/baike/c0%3Dbaike80%2C5%2C5%2C80%2C26/sign=fac3cc06cbfcc3cea0cdc161f32cbded/279759ee3d6d55fb0f2836026e224f4a20a4dd2c.jpg",
		Des:   "铜头铁臂，刀枪不入。黄娃三弟是个拥有钢筋铁骨的神娃，刀枪箭炮对他丝毫无伤",
		Attr:  map[string]interface{}{"property": "铜头铁臂"},
		Bg:    "FBEAEE",
	},
	"4": Huluwa{
		Name:  "四娃",
		Image: "https://gss2.bdstatic.com/9fo3dSag_xI4khGkpoWK1HF6hhy/baike/c0%3Dbaike80%2C5%2C5%2C80%2C26/sign=1e2e850b39292df583cea447dd583705/8326cffc1e178a8233c13f84f503738da977e811.jpg",
		Des:   "炉火纯青，刚阳烈焰。绿娃四弟乃天界火神下凡，可任意吞吐烈火",
		Attr:  map[string]interface{}{"property": "火"},
		Bg:    "FBEAEE",
	},
	"5": Huluwa{
		Name:  "五娃",
		Image: "https://gss3.bdstatic.com/-Po3dSag_xI4khGkpoWK1HF6hhy/baike/c0%3Dbaike80%2C5%2C5%2C80%2C26/sign=7be24f3d2ef5e0fefa1581533d095fcd/cefc1e178a82b901c16d8317708da9773812efd6.jpg",
		Des:   "惊涛骇浪，气吞山河。青娃五弟乃江河水神转世，可任意吞吐江河之水，难怪吞饮百坛烈酒而不醉",
		Attr:  map[string]interface{}{"property": "水"},
		Bg:    "FBEAEE",
	},
	"6": Huluwa{
		Name:  "六娃",
		Image: "https://gss1.bdstatic.com/-vo3dSag_xI4khGkpoWK1HF6hhy/baike/c0%3Dbaike80%2C5%2C5%2C80%2C26/sign=4c2abaab6f81800a7ae8815cd05c589f/8601a18b87d6277f035d711c2b381f30e924fc43.jpg",
		Des:   "来无影，去无踪。蓝娃六弟是七兄弟中最灵敏，最聪明的。他的隐身法令妖怪束手无策，并大闹妖洞，智救五哥，夺走了宝贝如意",
		Attr:  map[string]interface{}{"property": "隐身"},
		Bg:    "FBEAEE",
	},
	"7": Huluwa{
		Name:  "七娃",
		Image: "https://gss2.bdstatic.com/9fo3dSag_xI4khGkpoWK1HF6hhy/baike/c0%3Dbaike80%2C5%2C5%2C80%2C26/sign=35bba3768635e5dd8421ad8d17afcc8a/5bafa40f4bfbfbedd11d755b7bf0f736afc31f49.jpg",
		Des:   "镇妖之宝，本领无穷。紫娃七弟自身没多大本事，但他有个宝葫芦，是太上老君修炼仙丹用的紫金神葫",
		Attr:  map[string]interface{}{"property": "葫芦"},
		Bg:    "FBEAEE",
	},
}

func GetHuluwa(c *gin.Context) {
	hlw := hulubrothers[c.Param("id")]
	c.JSON(http.StatusOK, hlw)
}
