package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// StatusStore represents current status cache store
type StatusStore struct {
	dynamoDB *dynamodb.Client
}

var tableName = aws.String("ZatsuMonitor")

const (
	// NotFoundKey represents value if key is not found
	NotFoundKey = -1
)

// NewStatusStore create new StatusStore instance
func NewStatusStore(_ string) *StatusStore {
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	s := new(StatusStore)
	s.dynamoDB = dynamodb.NewFromConfig(awsConfig)
	return s
}

// GetDbStatus returns status code for specified key
func (s *StatusStore) GetDbStatus(key string) (int, error) {
	result, err := s.dynamoDB.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"Key": &types.AttributeValueMemberS{
				Value: key,
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

	statusCode, ok := result.Item["StatusCode"]
	if !ok {
		return NotFoundKey, nil
	}
	statusCodeN, ok := statusCode.(*types.AttributeValueMemberN)
	if !ok {
		return NotFoundKey, fmt.Errorf("invalid attribule value type (%T)", statusCode)
	}

	return strconv.Atoi(statusCodeN.Value)
}

// SaveDbStatus saves status code for specified key
func (s *StatusStore) SaveDbStatus(key string, statusCode int) error {
	statusCodeStr := strconv.Itoa(statusCode)
	_, err := s.dynamoDB.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]string{
			"#S": "StatusCode",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberN{
				Value: statusCodeStr,
			},
		},
		Key: map[string]types.AttributeValue{
			"Key": &types.AttributeValueMemberS{
				Value: key,
			},
		},
		UpdateExpression: aws.String("SET #S = :s"),
		TableName:        tableName,
	})
	return err
}
