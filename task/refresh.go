package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paralleltree/heartbeatmon/metric"
	"github.com/paralleltree/heartbeatmon/persistence"
)

type clientCredentialConfig struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type accessTokenCredentialConfig struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt,omitempty"`
}

type record struct {
	Time      time.Time `json:"time"`
	HeartRate int       `json:"heart_rate"`
}

func RefreshHeartrate(ctx context.Context, clientCredentialStore, accessTokenStore, recordStore persistence.PersistentStore) error {
	clientCredential, err := loadJSONData[clientCredentialConfig](ctx, clientCredentialStore)
	if err != nil {
		return fmt.Errorf("load client credential: %w", err)
	}
	accessTokenCredential, err := loadJSONData[accessTokenCredentialConfig](ctx, accessTokenStore)
	if err != nil {
		return fmt.Errorf("load access token credential: %w", err)
	}

	// fetch latest heartrate
	client := metric.NewFitbitClient(ctx, clientCredential.ClientID, clientCredential.ClientSecret, accessTokenCredential.AccessToken, accessTokenCredential.RefreshToken)
	if !accessTokenCredential.ExpiresAt.Equal(time.Time{}) {
		client = metric.NewFitbitClientWithExpiry(ctx, clientCredential.ClientID, clientCredential.ClientSecret, accessTokenCredential.AccessToken, accessTokenCredential.RefreshToken, accessTokenCredential.ExpiresAt)
	}
	records, err := client.GetHeartrate(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("get heartrate: %w", err)
	}

	// persist refreshed token
	newToken := client.GetToken()
	if accessTokenCredential.AccessToken != newToken.AccessToken {
		accessTokenCredential.AccessToken = newToken.AccessToken
		accessTokenCredential.RefreshToken = newToken.RefreshToken
		accessTokenCredential.ExpiresAt = newToken.Expiry
		if err := saveJSONData(ctx, accessTokenStore, accessTokenCredential); err != nil {
			return fmt.Errorf("save access token: %w", err)
		}
	}

	if len(records) == 0 {
		return nil
	}

	// save latest heartrate
	latest := record{
		Time:      records[len(records)-1].Time,
		HeartRate: records[len(records)-1].Value,
	}
	if err := saveJSONData(ctx, recordStore, latest); err != nil {
		return fmt.Errorf("save record: %w", err)
	}

	return nil
}

func loadJSONData[T any](ctx context.Context, store persistence.PersistentStore) (T, error) {
	var res T
	body, err := store.Load(ctx)
	if err != nil {
		return res, fmt.Errorf("load data: %w", err)
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return res, fmt.Errorf("unmarshal json: %w", err)
	}
	return res, nil
}

func saveJSONData[T any](ctx context.Context, store persistence.PersistentStore, data T) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := store.Save(ctx, body); err != nil {
		return fmt.Errorf("save object: %w", err)
	}
	return nil
}
