package metric

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/paralleltree/heartbeatmon/lib"
	"golang.org/x/oauth2"
)

const (
	FitbitAuthURL  = "https://www.fitbit.com/oauth2/authorize"
	FitbitTokenURL = "https://api.fitbit.com/oauth2/token"
)

type FitbitRequestError struct {
	URL        string
	StatusCode int
	RawBody    string
	Errors     []struct {
		ErrorType string `json:"errorType"`
		Message   string `json:"message"`
	} `json:"errors"`
}

func (err *FitbitRequestError) Error() string {
	if len(err.Errors) == 0 {
		return fmt.Sprintf("failed to request: %s", err.URL)
	}
	e := err.Errors[0]
	return fmt.Sprintf("failed to request: %s (%s): %s", err.URL, e.ErrorType, e.Message)
}

type FitbitHeartrateRecord struct {
	Time  time.Time
	Value int
}

type FitbitClient struct {
	baseURL   string
	token     *oauth2.Token
	oauthConf *oauth2.Config
}

func newFitbitClient(
	ctx context.Context,
	clientID, clientSecret string, accessToken, refreshToken string,
	baseURL string,
) *FitbitClient {
	// set current time to force refresh access token
	token := oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken, Expiry: time.Now()}
	return &FitbitClient{
		baseURL:   baseURL,
		token:     &token,
		oauthConf: FitbitOAuthConf(clientID, clientSecret),
	}
}

func NewFitbitClient(ctx context.Context, clientID, clientSecret string, accessToken, refreshToken string) *FitbitClient {
	return newFitbitClient(ctx, clientID, clientSecret, accessToken, refreshToken, "https://api.fitbit.com")
}

func (c *FitbitClient) GetToken() *oauth2.Token {
	return c.token
}

func (c *FitbitClient) GetHeartrate(ctx context.Context, date time.Time) ([]FitbitHeartrateRecord, error) {
	date = date.UTC().Truncate(time.Hour * 24)
	tokenSource := c.oauthConf.TokenSource(ctx, c.token)
	client := oauth2.NewClient(ctx, tokenSource)
	userID := "-"
	detailLevel := "1min"
	endpoint := fmt.Sprintf("%s/1/user/%s/activities/heart/date/%s/1d/%s.json?timezone=UTC", c.baseURL, userID, date.Format("2006-01-02"), detailLevel)

	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("get heartrate: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	switch resp.StatusCode {
	case 400, 401:
		errBody := FitbitRequestError{
			URL:        endpoint,
			StatusCode: resp.StatusCode,
			RawBody:    string(body),
		}
		if err := json.Unmarshal(body, &errBody); err != nil {
			return nil, fmt.Errorf("parse error response")
		}
		return nil, &errBody
	}

	type fitbitHeartRateIntraday struct {
		ActivitiesHeartIntraday struct {
			DataSet []struct {
				Time  string `json:"time"`
				Value int    `json:"value"`
			} `json:"dataset"`
		} `json:"activities-heart-intraday"`
	}

	intraday := &fitbitHeartRateIntraday{}
	if err := json.Unmarshal(body, intraday); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	res := []FitbitHeartrateRecord{}
	for _, v := range intraday.ActivitiesHeartIntraday.DataSet {
		t, err := lib.ParseTime(v.Time)
		if err != nil {
			return nil, fmt.Errorf("parse time: %w", err)
		}
		record := FitbitHeartrateRecord{
			Time:  date.Add(t),
			Value: v.Value,
		}
		res = append(res, record)
	}

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	c.token = newToken
	return res, nil
}

func FitbitOAuthConf(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"heartrate"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  FitbitAuthURL,
			TokenURL: FitbitTokenURL,
		},
	}
}
