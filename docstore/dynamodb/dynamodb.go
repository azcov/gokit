package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	awsddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/azcov/gokit/docstore"
)

var _ docstore.Store = (*DynamoDB)(nil)

type DynamoDB struct {
	client *awsddb.Client
	pkName string
}

type Config struct {
	AWSConfig aws.Config `json:"-" yaml:"-"`
	// PartitionKey is the DynamoDB attribute name used as the document ID. Defaults to "id".
	PartitionKey string `config:"partition_key"`
}

func New(cfg Config) *DynamoDB {
	pk := cfg.PartitionKey
	if pk == "" {
		pk = "id"
	}
	return &DynamoDB{
		client: awsddb.NewFromConfig(cfg.AWSConfig),
		pkName: pk,
	}
}

func (d *DynamoDB) Get(ctx context.Context, collection, id string, dest any) error {
	key, err := attributevalue.MarshalMap(map[string]any{d.pkName: id})
	if err != nil {
		return err
	}
	out, err := d.client.GetItem(ctx, &awsddb.GetItemInput{
		TableName: aws.String(collection),
		Key:       key,
	})
	if err != nil {
		return err
	}
	if out.Item == nil {
		return fmt.Errorf("docstore/dynamodb: item not found: %s/%s", collection, id)
	}
	return attributevalue.UnmarshalMap(out.Item, dest)
}

func (d *DynamoDB) Set(ctx context.Context, collection, id string, doc any) error {
	item, err := attributevalue.MarshalMap(doc)
	if err != nil {
		return err
	}
	item[d.pkName] = &types.AttributeValueMemberS{Value: id}
	_, err = d.client.PutItem(ctx, &awsddb.PutItemInput{
		TableName: aws.String(collection),
		Item:      item,
	})
	return err
}

func (d *DynamoDB) Update(ctx context.Context, collection, id string, fields map[string]any) error {
	names := map[string]string{}
	values := map[string]types.AttributeValue{}
	setParts := make([]string, 0, len(fields))

	for i, k := range sortedKeys(fields) {
		nk := fmt.Sprintf("#k%d", i)
		vk := fmt.Sprintf(":v%d", i)
		names[nk] = k
		av, err := attributevalue.Marshal(fields[k])
		if err != nil {
			return err
		}
		values[vk] = av
		setParts = append(setParts, nk+" = "+vk)
	}

	key, err := attributevalue.MarshalMap(map[string]any{d.pkName: id})
	if err != nil {
		return err
	}
	expr := "SET " + strings.Join(setParts, ", ")
	_, err = d.client.UpdateItem(ctx, &awsddb.UpdateItemInput{
		TableName:                 aws.String(collection),
		Key:                       key,
		UpdateExpression:          aws.String(expr),
		ExpressionAttributeNames:  names,
		ExpressionAttributeValues: values,
	})
	return err
}

func (d *DynamoDB) Delete(ctx context.Context, collection, id string) error {
	key, err := attributevalue.MarshalMap(map[string]any{d.pkName: id})
	if err != nil {
		return err
	}
	_, err = d.client.DeleteItem(ctx, &awsddb.DeleteItemInput{
		TableName: aws.String(collection),
		Key:       key,
	})
	return err
}

func (d *DynamoDB) List(ctx context.Context, collection string, q docstore.Query) ([]map[string]any, error) {
	input := &awsddb.ScanInput{TableName: aws.String(collection)}

	if len(q.Filters) > 0 {
		names := map[string]string{}
		values := map[string]types.AttributeValue{}
		parts := make([]string, 0, len(q.Filters))
		for i, f := range q.Filters {
			nk := fmt.Sprintf("#fk%d", i)
			vk := fmt.Sprintf(":fv%d", i)
			names[nk] = f.Field
			av, err := attributevalue.Marshal(f.Value)
			if err != nil {
				return nil, err
			}
			values[vk] = av
			parts = append(parts, nk+" "+ddbOp(f.Op)+" "+vk)
		}
		expr := strings.Join(parts, " AND ")
		input.FilterExpression = aws.String(expr)
		input.ExpressionAttributeNames = names
		input.ExpressionAttributeValues = values
	}

	if q.Limit > 0 {
		input.Limit = aws.Int32(int32(q.Limit))
	}

	out, err := d.client.Scan(ctx, input)
	if err != nil {
		return nil, err
	}

	docs := make([]map[string]any, 0, len(out.Items))
	for _, item := range out.Items {
		m := map[string]any{}
		if err := attributevalue.UnmarshalMap(item, &m); err != nil {
			return nil, err
		}
		docs = append(docs, m)
	}
	return docs, nil
}

func (d *DynamoDB) Close() error { return nil }

func ddbOp(op string) string {
	switch op {
	case docstore.OpNe:
		return "<>"
	default:
		return op
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
