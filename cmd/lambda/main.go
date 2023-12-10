package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/paralleltree/heartbeatmon/persistence"
	"github.com/paralleltree/heartbeatmon/task"
)

type message struct {
	Region     string `json:"region"`
	BucketName string `json:"bucketName"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, msg message) error {
	clientCredentialStore, err := persistence.NewS3Store(msg.Region, msg.BucketName, "clientCredential.json")
	if err != nil {
		return err
	}
	accessTokenStore, err := persistence.NewS3Store(msg.Region, msg.BucketName, "accessToken.json")
	if err != nil {
		return err
	}
	recordStore, err := persistence.NewS3Store(msg.Region, msg.BucketName, "latest.json")
	if err != nil {
		return err
	}
	if err := task.RefreshHeartrate(ctx, clientCredentialStore, accessTokenStore, recordStore); err != nil {
		return err
	}
	return nil
}
