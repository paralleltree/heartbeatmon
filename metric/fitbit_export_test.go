package metric

import (
	"context"
	"time"
)

func NewTestFitbitClient(ctx context.Context, clientID, clientSecret string, accessToken, refreshToken string, baseURL string) *FitbitClient {
	return newFitbitClient(ctx, clientID, clientSecret, accessToken, refreshToken, baseURL, time.Time{})
}
