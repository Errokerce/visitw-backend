package router

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

var r *gin.Engine

func init() {

	fmt.Println("router loading...")

	rr := gin.Default()
	r = rr

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://visitw.nctu.me", "https://api.visitw.nctu.me",
		"https://localhost:4000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"X-Total-Count"}

	r.Use(cors.New(config))
}
func routeMap() {

	apiG := r.Group("/api")

	apiG.POST("/login/phone", PhoneLoginStepOne)
	apiG.POST("/login/code", PhoneLoginStepTwo)

	apiG.PATCH("/ping", func(c *gin.Context) {
		type (
			RecvAddr struct {
				City   string `json:"r_city"`
				Town   string `json:"r_town"`
				Street string `json:"r_street"`
			}
			Recv struct {
				ShopID      string   `json:"r_shop_id"`
				ShopName    string   `json:"r_shop_name"`
				ShopPhone   string   `json:"r_shop_phone"`
				ShopAddress RecvAddr `json:"r_shop_address"`
			}
		)
		rv := Recv{}
		err := c.ShouldBindJSON(&rv)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"a": rv})
	})

	apiG.Use(CheckJWT())

	apiG.GET("/user", GetFullUserData)
	apiG.GET("/user/shop_list", GetShopListOfUser)

	apiG.PATCH("/user/fcm_token", SaveFCMToken)
	apiG.PATCH("/user/name", UpdateUserName)

	apiG.GET("/shop/:sid", GetShopInfo)

	apiG.POST("/shop/reg", NewShop)
	apiG.POST("/shop/query", QueryShopPhoneBySnippet)

	apiG.GET("/visit/:sid", ConfirmVisit)

	apiG.GET("/visited_list")
}
func Start(port int) {
	routeMap()
	portS := strconv.Itoa(port)
	log.Fatal(r.Run(":" + portS))
}
