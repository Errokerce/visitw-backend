package db

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"time"
)

type (
	UserTable struct {
		UserID    *string   `json:"user_id"` //hash of phone
		UserMail  *string   `json:"user_mail"`
		UserPhone *string   `json:"user_phone"` //需要加密
		UserName  *string   `json:"user_name"`
		FCMToken  *string   `json:"fcm_token"`
		ShopList  []*string `json:"shop_list"`
	}
	ShopTable struct {
		ShopID      *string  `json:"shop_id"`
		ShopName    *string  `json:"shop_name"`
		ShopPhone   *string  `json:"shop_phone"`
		ShopAddress *Address `json:"shop_address"`
	}
	Address struct {
		City   *string `json:"city"`
		Town   *string `json:"town"`
		Street *string `json:"street"`
	}
	Visited struct {
		UserID *string `json:"user_id"`
		ShopID *string `json:"shop_id"`
		Date   *string `json:"date"`
		Time   *int64  `json:"time"`
	}
	NewShopID struct {
		ShopID *string `json:"shop_id"`
		NewID  *int    `json:"newest_id"`
	}
)

func NewVisited(userId, shopId *string) *Visited {
	nt := time.Now()
	return &Visited{
		UserID: userId,
		ShopID: shopId,
		Date:   aws.String(fmt.Sprintf("%d/%d/%d", nt.Year(), nt.Month(), nt.Day())),
		Time:   aws.Int64(nt.UnixNano()),
	}
}

func GetNewestID() *string {
	resl := NewShopID{}
	DbQuerybyKey(TableShop, "shop_id", "00000000", &resl)

	resp := aws.String(fmt.Sprintf("%08d", *resl.NewID))

	kn := map[string]*string{"#n": aws.String("newest_id")}

	type Updv struct {
		NewID *int `json:":n"`
	}
	*resl.NewID++
	u := Updv{resl.NewID}
	DbUpdate(TableShop, "shop_id", "00000000",
		"set #n = :n", kn, u)

	return resp
}

func CheckShopPhone(phone string) *ShopTable {
	resp := ShopTable{}

	filt := expression.Name("shop_phone").Equal(expression.Value(aws.String(phone)))
	proj := expression.NamesList(expression.Name("shop_id"), expression.Name("shop_phone"))

	_, _ = DbQueryMany(TableShop, filt, proj, &resp)

	return &resp
}
