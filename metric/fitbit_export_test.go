package metric

import "context"

func NewTestFitbitClient(ctx context.Context, clientID, clientSecret string, accessToken, refreshToken string, baseURL string) *FitbitClient {
	return newFitbitClient(ctx, clientID, clientSecret, accessToken, refreshToken, baseURL, false)
}
