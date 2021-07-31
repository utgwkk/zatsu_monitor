package main

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// StatusStore represents current status cache store
type StatusStore struct {
	dynamoDB *dynamodb.DynamoDB
}

var awsSession = session.Must(session.NewSession())
var tableName = aws.String("ZatsuMonitor")

const (
	// NotFoundKey represents value if key is not found
	NotFoundKey = -1
)

// NewStatusStore create new StatusStore instance
func NewStatusStore(databaseFile string) *StatusStore {
	s := new(StatusStore)
	s.dynamoDB = dynamodb.New(awsSession)
	return s
}

// GetDbStatus returns status code for specified key
func (s *StatusStore) GetDbStatus(key string) (int, error) {
	result, err := s.dynamoDB.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
		},
		TableName: tableName,
	})
	if err != nil {
		return 0, err
	}

	if result == nil {
		return NotFoundKey, nil
	}

	statusCode, ok := result.Item["status_code"]
	if !ok {
		return NotFoundKey, nil
	}

	return strconv.Atoi(*statusCode.N)
}

// SaveDbStatus saves status code for specified key
func (s *StatusStore) SaveDbStatus(key string, statusCode int) error {
	statusCodeStr := strconv.Itoa(statusCode)
	_, err := s.dynamoDB.UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#S": aws.String("StatusCode"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				N: aws.String(statusCodeStr),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
		},
		UpdateExpression: aws.String("SET #S = :s"),
		TableName:        tableName,
	})
	return err
}
