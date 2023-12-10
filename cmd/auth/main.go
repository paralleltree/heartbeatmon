package main

import (
	"context"
	"fmt"
	"log"

	"github.com/paralleltree/heartbeatmon/metric"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	if err := do(ctx); err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
}

func do(ctx context.Context) error {
	fmt.Printf("Enter client id: ")
	clientID := readLine()
	fmt.Printf("Enter client secret: ")
	clientSecret := readLine()

	url, verifier, err := startAuthorization(clientID, clientSecret)
	if err != nil {
		return err
	}

	fmt.Println(url)

	token, err := verifier(ctx, readLine())

	if err != nil {
		return err
	}

	fmt.Printf("access token: %s\n", token.AccessToken)
	fmt.Printf("refresh token: %s\n", token.RefreshToken)
	return nil
}

func readLine() string {
	var item string
	if _, err := fmt.Scan(&item); err != nil {
		panic(err)
	}
	return item
}

type verifierFunc func(ctx context.Context, authCode string) (*oauth2.Token, error)

func startAuthorization(clientID, clientSecret string) (string, verifierFunc, error) {
	conf := metric.FitbitOAuthConf(clientID, clientSecret)
	url := conf.AuthCodeURL("state")

	verifier := func(ctx context.Context, authCode string) (*oauth2.Token, error) {
		token, err := conf.Exchange(ctx, authCode)
		if err != nil {
			return nil, err
		}
		return token, nil
	}

	return url, verifier, nil
}
