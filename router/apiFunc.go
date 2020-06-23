package router

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"time"
	"visitw_backend/db"
	. "visitw_backend/misc"
	. "visitw_backend/sms"
	//"time"
)

var (
	smsVerify    map[string]VerifyInfo
	cookieDomain = fmt.Sprintf("." + Config.Hostname)
	cookieTime   = 60 * 60 * 24 * 28
)

func init() {
	fmt.Println("apiFunc loading...")
	smsVerify = make(map[string]VerifyInfo)
}

//接收電話號碼並祭出驗證碼
func PhoneLoginStepOne(c *gin.Context) {

	//資料綁定
	type pr struct {
		Phone string `json:"phone"`
	}
	var ps pr
	err := c.ShouldBindJSON(&ps)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "error",
			"msg":   "error",
		})
		return
	}

	//驗證正規表達式
	matched, err := regexp.Match(`^09\d{8}$`, []byte(ps.Phone))
	//輸入錯誤的反應
	if !matched {
		c.JSON(http.StatusOK, gin.H{
			"state": "fail",
			"msg":   "input error",
		})
		return

	}

	if time.Now().Before(smsVerify[ps.Phone].ExprTime.Add(time.Minute * 4 * -1)) {
		c.JSON(http.StatusOK, gin.H{
			"state": "wait",
			"msg":   "wait",
		})
		return
	}

	code := RandCode(6)
	fmt.Println(ps.Phone, code)

	smsVerify[ps.Phone] = VerifyInfo{
		Code: code, ExprTime: time.Now().Add(time.Minute * 5),
	}

	//實際寄出簡訊
	//SendVerifyCode(ps.Phone, code)

	c.JSON(http.StatusOK, gin.H{
		"state":     "ok",
		"expr_time": smsVerify[ps.Phone].ExprTime.Unix() * 1000,
	})
	return
}

func PhoneLoginStepTwo(c *gin.Context) {
	//資料綁定
	type pr struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var ps pr
	err := c.ShouldBindJSON(&ps)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "error",
			"msg":   "error",
		})
		return
	}

	m1, _ := regexp.Match(`^09\d{8}$`, []byte(ps.Phone))
	m2, _ := regexp.Match(`^\d{6}$`, []byte(ps.Code))
	//輸入錯誤的反應
	if !m1 || !m2 {
		c.JSON(http.StatusOK, gin.H{
			"state": "fail",
			"msg":   "input error",
		})
		return
	}

	if time.Now().After(smsVerify[ps.Phone].ExprTime) {
		c.JSON(http.StatusOK, gin.H{
			"state": "time_out",
			"msg":   "time_out",
		})
		return
	}

	if smsVerify[ps.Phone].Code != ps.Code {

		code := RandCode(6)
		fmt.Println(ps.Phone, code)

		smsVerify[ps.Phone] = VerifyInfo{
			Code: code, ExprTime: time.Now().Add(time.Minute * 5),
		}

		SendVerifyCode(ps.Phone, code)

		c.JSON(http.StatusOK, gin.H{
			"state": "wrong",
			"msg":   "input error",
		})
		return
	} else {
		smsVerify[ps.Phone] = VerifyInfo{}
		delete(smsVerify, ps.Phone)

		uu := db.UserTable{}
		db.DbQuerybyKey(db.TableUser, "user_id", GetHash(ps.Phone), &uu)
		if uu.UserID == nil {
			uu = db.UserTable{
				UserID:    aws.String(GetHash(ps.Phone)),
				UserPhone: aws.String(MyEncryper(ps.Phone)),
				UserName:  aws.String(GetPhoneMask(ps.Phone)),
				ShopList:  aws.StringSlice([]string{""}),
			}
			db.DbAdd(db.TableUser, uu)
		}
		token, err := NewJwtClaims(*uu.UserID, *uu.UserName).GenerateToken()
		if err != nil {
			fmt.Println(err)
			return
		}
		c.SetCookie("jwtAccess", token, cookieTime, "/", cookieDomain, false, false)

		c.JSON(http.StatusOK, gin.H{
			"state": "ok",
			"msg":   "success",
		})
		return
	}

}

func QueryShopPhoneBySnippet(c *gin.Context) {
	type Recv struct {
		PhoneSnippet string `json:"phone_snippet"`
	}
	r := Recv{}
	err := c.ShouldBindJSON(&r)
	if err != nil {
		fmt.Println(err)

		c.JSON(http.StatusBadRequest, gin.H{"state": "input error"})
		return
	}

	var resp db.ShopTable
	var respa []db.ShopTable

	filt := expression.Name("shop_phone").BeginsWith(r.PhoneSnippet)
	proj := expression.NamesList(expression.Name("shop_id"), expression.Name("shop_phone"))

	ls, err := db.DbQueryMany(db.TableShop, filt, proj, resp)
	if err != nil {
		fmt.Println("query Error")
		c.JSON(http.StatusBadRequest, gin.H{"state": "query error"})
		return
	}

	if ls == nil {

		c.JSON(http.StatusBadRequest, gin.H{"state": "not found"})
		return
	}

	for _, i2 := range ls {
		var ts db.ShopTable
		err := json.Unmarshal(i2.([]byte), &ts)
		if err != nil {

			fmt.Println(err)

			c.JSON(http.StatusBadRequest, gin.H{"state": "query error"})
			return
		}
		//fmt.Println(ts)
		respa = append(respa, ts)
	}

	c.JSON(http.StatusOK, gin.H{"state": "ok", "phone_list": respa})
}

func ConfirmVisit(c *gin.Context) {
	shopID := c.Param("sid")

	if shopID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": "emptyQuery"})
		return
	}

	s := db.ShopTable{}
	db.DbQuerybyKey(db.TableShop, "shop_id", shopID, &s)

	if s.ShopID == nil {
		c.JSON(http.StatusNotFound, gin.H{"state": "notFound"})
		return
	}
	userID := c.MustGet("tokenClaim").(*Claims).UserID
	nV := db.NewVisited(aws.String(userID), aws.String(shopID))

	db.DbAdd(db.TableVisited, nV)

	c.JSON(http.StatusOK, gin.H{"state": "ok"})

}

func GetShopInfo(c *gin.Context) {
	shopID := c.Param("sid")
	if shopID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": "emptyQuery"})
		return
	}

	s := db.ShopTable{}
	db.DbQuerybyKey(db.TableShop, "shop_id", shopID, &s)

	if s.ShopID == nil {
		c.JSON(http.StatusNotFound, gin.H{"state": "notFound"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"state": "ok", "shop_info": s})

}

func NewShop(c *gin.Context) {
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
		Sls struct {
			ShopId    string `json:"shop_id"`
			ShopName  string `json:"shop_name"`
			ShopPhone string `json:"shop_phone"`
		}
	)
	rv := Recv{}
	err := c.ShouldBindJSON(&rv)
	if err != nil {
		fmt.Println(err)
	}

	if rv.ShopAddress.City == "" && rv.ShopAddress.Street == "" && rv.ShopAddress.Town == "" &&
		rv.ShopName == "" && rv.ShopPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": "can't validate", "recv": rv})
		return
	}
	nn := db.CheckShopPhone(rv.ShopPhone)
	if nn.ShopID != nil {
		c.JSON(http.StatusOK, gin.H{"state": "repeat", "shop_info": Sls{*nn.ShopID, *nn.ShopName, *nn.ShopPhone}})
		return
	}
	nid := db.GetNewestID()
	db.DbAdd(db.TableShop, db.ShopTable{ShopID: nid,
		ShopName:  aws.String(rv.ShopName),
		ShopPhone: aws.String(rv.ShopPhone),
		ShopAddress: &db.Address{
			City:   aws.String(rv.ShopAddress.City),
			Town:   aws.String(rv.ShopAddress.Town),
			Street: aws.String(rv.ShopAddress.Street),
		}})

	userID := c.MustGet("tokenClaim").(*Claims).UserID

	kn := map[string]*string{"#n": aws.String("shop_list")}

	type Updv struct {
		NewID     []string `json:":n"`
		EmptyList []string `json:":e"`
	}
	u := Updv{[]string{*nid}, []string{""}}
	db.DbUpdate(db.TableUser, "user_id", userID,
		"set #n = list_append( :n , if_not_exists(#n, :e))", kn, u)

	uu := db.UserTable{}
	db.DbQuerybyKey(db.TableUser, "user_id", userID, &uu)
	//fmt.Println(rv)
	//sl, err := c.Cookie("myShopList")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//if sl == "" {
	//	sl = *nid
	//} else {
	//	sl += ", " + *nid
	//}
	//c.SetCookie("myShopList", sl, cookieTime*12, "/", cookieDomain, false, false)

	c.JSON(http.StatusOK, gin.H{"state": "ok", "new_shop": Sls{*nid, rv.ShopName, rv.ShopPhone}})
}

func GetShopListOfUser(c *gin.Context) {
	userID := c.MustGet("tokenClaim").(*Claims).UserID
	uu := db.UserTable{}
	db.DbQuerybyKey(db.TableUser, "user_id", userID, &uu)

	type Sls struct {
		ShopId    string `json:"shop_id"`
		ShopName  string `json:"shop_name"`
		ShopPhone string `json:"shop_phone"`
	}

	sl := []Sls{}

	for i, i2 := range uu.ShopList {
		if i2 != nil {
			fmt.Println(i, *i2)
			tmp := Sls{}
			db.DbQuerybyKey(db.TableShop, "shop_id", *i2, &tmp)
			if tmp.ShopId != "" {
				sl = append(sl, tmp)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"state": "ok", "shop_list": sl})

}

func GetFullUserData(c *gin.Context) {

	claim := c.MustGet("tokenClaim").(*Claims)

	u := db.UserTable{}

	db.DbQuerybyKey(db.TableUser, "user_id", claim.UserID, &u)
	u.UserPhone = nil
	c.JSON(http.StatusOK, gin.H{"state": "ok", "user_data": u})

}

func SaveFCMToken(c *gin.Context) {
	type Recv struct {
		FCMToken string `json:"fcm_token"`
	}
	r := Recv{}
	err := c.ShouldBindJSON(&r)
	if err != nil {
		fmt.Println(err)
	}

	claim := c.MustGet("tokenClaim").(*Claims)

	kn := map[string]*string{"#n": aws.String("fcm_token")}
	type Updv struct {
		NewToken *string `json:":n"`
	}
	u := Updv{aws.String(r.FCMToken)}
	db.DbUpdate(db.TableUser, "user_id", claim.UserID,
		"set #n = :n", kn, u)

	c.JSON(http.StatusOK, gin.H{"state": "goooooood"})

}

func UpdateUserName(c *gin.Context) {
	type Recv struct {
		UserName string `json:"r_user_name"`
	}
	r := Recv{}
	err := c.ShouldBindJSON(&r)
	if err != nil {
		fmt.Println(err)
	}

	claim := c.MustGet("tokenClaim").(Claims)

	kn := map[string]*string{"#n": aws.String("user_name")}
	type Updv struct {
		NewName *string `json:":n"`
	}
	u := Updv{aws.String(r.UserName)}
	db.DbUpdate(db.TableUser, "user_id", claim.UserID,
		"set #n = :n", kn, u)

	c.JSON(http.StatusOK, gin.H{"state": "goooooood"})
}

func CheckJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("jwtAccess")
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(token)

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"state": "not Login"})
			c.Abort()
			return
		} else {
			tokenState, claims, err := ParseJwtToken(token)
			if err != nil {
				fmt.Println(err)
			}
			switch tokenState {

			case TokenAccecpt:
				c.Set("tokenClaim", claims)
				c.Next()
			case TokenTimeout:
				c.JSON(http.StatusUnauthorized, gin.H{"state": "token timeout"})
				c.Abort()
			case TokenValidationFailed:
				c.JSON(http.StatusUnauthorized, gin.H{"state": "can't validate"})
				c.Abort()
			}
		}

	}
}
