package db

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"os"
)

/*

important file :
https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/expression/#pkg-examples


*/

const (
	TableUser    = "USER"
	TableShop    = "SHOP"
	TableVisited = "VISITED"

	awsTokyo = "ap-northeast-1"

	DebugMode = false
)

var svc *dynamodb.DynamoDB

func init() {

	fmt.Println("Database loading...")

	sess, err := session.NewSession(&aws.Config{Region: aws.String(awsTokyo)})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	svc = dynamodb.New(sess)
	fmt.Println("Datebase connect SUCCESS")
}

// to create and init a DataTable with a key attribute
func CreateDBtable(tableName, keyName string) {

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(keyName),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(keyName),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String(tableName),
	}
	_, err := svc.CreateTable(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if DebugMode {
		fmt.Println("Table " + tableName + " creating Succeeded")
	}
}

// to push a new object into Table
func DbAdd(tableName string, object interface{}) {

	av, err := dynamodbattribute.MarshalMap(object)

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = svc.PutItem(input)

	if err != nil {

		fmt.Println("P2")
		fmt.Println(err)
		//os.Exit(1)
	}

	if DebugMode {
		fmt.Println(object)
		fmt.Println("Table item adding Succeeded")
	}
}

// to query an object by key name and key value
func DbQuerybyKey(tableName, keyName, keyValue string, ri interface{}) {

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			keyName: {
				S: aws.String(keyValue),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, ri)

	if DebugMode {
		fmt.Println(ri)
	}

}

// to query objects by filter
func DbQueryMany(tableName string, filt expression.ConditionBuilder, proj expression.ProjectionBuilder, structCont interface{}) ([]interface{}, error) {

	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		fmt.Println("Got error building expression:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		//os.Exit(1)
		return nil, err
	}

	var iif []interface{}
	for _, i := range result.Items {

		err = dynamodbattribute.UnmarshalMap(i, &structCont)

		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		j, err := json.Marshal(&structCont)
		if err != nil {
			fmt.Println("Json error marshalling:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		iif = append(iif, j)
	}

	if DebugMode {
		fmt.Println(iif)
	}
	return iif, nil

}

func DbUpdate(tableName, keyName, keyValue, updCmd string, updateKeyname map[string]*string, updateData interface{}) bool {

	upd, err := dynamodbattribute.MarshalMap(updateData)
	if err != nil {
		fmt.Println(err.Error())
		return true
	}

	input := &dynamodb.UpdateItemInput{

		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			keyName: {
				S: aws.String(keyValue),
			},
		},
		ExpressionAttributeNames:  updateKeyname,
		ExpressionAttributeValues: upd,
		ReturnValues:              aws.String("UPDATED_NEW"),
		UpdateExpression:          aws.String(updCmd),
	}

	_, err = svc.UpdateItem(input)
	if err != nil {
		fmt.Println(err.Error())
		return true
	}
	if DebugMode {
		fmt.Println("update Succeeded")
	}

	return false
}

func updateExample() {

	kn := map[string]*string{"#o": aws.String("owner")}

	type updv struct {
		Owner string `json:":o"`
	}

	u := updv{"jp"}
	DbUpdate("ITEMS", "itemID", "0002",
		"set #o = :o", kn, u)
}
