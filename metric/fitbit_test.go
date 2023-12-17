package metric_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/paralleltree/heartbeatmon/metric"
	"golang.org/x/oauth2"
)

func TestFitbitClient_GetHeartrate_When_Error(t *testing.T) {
	ctx := context.Background()
	clientID := "testId"
	clientSecret := "testSecret"
	accessToken := "testAccessToken"
	refreshToken := "testRefreshToken"
	date := time.Date(2023, 1, 1, 1, 2, 3, 0, time.UTC)

	cases := []struct {
		name           string
		serverResponse string
		statusCode     int
		wantErrorEntry metric.FitbitRequestErrorEntry
	}{
		{
			name: "when request was unauthorized",
			serverResponse: `{
				"success":false,
				"errors":[
					{"errorType":"expired_token","message":"Access token expired"}
				]}
			`,
			statusCode: http.StatusUnauthorized,
			wantErrorEntry: metric.FitbitRequestErrorEntry{
				ErrorType: "expired_token",
				Message:   "Access token expired",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			mux := http.NewServeMux()
			mux.HandleFunc("/1/user/-/activities/heart/date/2023-01-01/1d/1min.json", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				io.Copy(w, strings.NewReader(tt.serverResponse))
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			ctx = context.WithValue(ctx, oauth2.HTTPClient, server.Client())
			client := metric.NewTestFitbitClient(ctx, clientID, clientSecret, accessToken, refreshToken, server.URL)

			// act
			_, err := client.GetHeartrate(ctx, date)

			// assert
			if err == nil {
				t.Fatalf("error was not retuned")
			}

			reqErr, ok := err.(*metric.FitbitRequestError)
			if !ok {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.statusCode != reqErr.StatusCode {
				t.Fatalf("invalid status code: want %d, but got %d", tt.statusCode, reqErr.StatusCode)
			}

			errFound := false
			for _, e := range reqErr.Errors {
				if tt.wantErrorEntry.ErrorType == e.ErrorType && tt.wantErrorEntry.Message == e.Message {
					errFound = true
					break
				}
			}
			if !errFound {
				t.Fatalf("expected error was not found")
			}
		})
	}
}

func TestFitbitClient_GetHeartrate_When_Success(t *testing.T) {
	ctx := context.Background()
	jstTz := time.FixedZone("Asia/Tokyo", 9*60*60)
	clientID := "testId"
	clientSecret := "testSecret"
	accessToken := "testAccessToken"
	refreshToken := "testRefreshToken"

	cases := []struct {
		name string
		date time.Time
	}{
		{
			name: "given time is UTC",
			date: time.Date(2023, 1, 1, 1, 2, 3, 0, time.UTC),
		}, {
			name: "given time is JST",
			date: time.Date(2023, 1, 2, 1, 1, 1, 0, jstTz),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			mux := http.NewServeMux()
			mux.HandleFunc("/1/user/-/activities/heart/date/2023-01-01/1d/1min.json", func(w http.ResponseWriter, r *http.Request) {
				res := `{
					"activities-heart-intraday": {
						"dataset": [
							{
								"time": "01:02:03",
								"value": 78
							}
						]
					}
				}`
				w.WriteHeader(http.StatusOK)
				io.Copy(w, strings.NewReader(res))
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			ctx = context.WithValue(ctx, oauth2.HTTPClient, server.Client())
			client := metric.NewTestFitbitClient(ctx, clientID, clientSecret, accessToken, refreshToken, server.URL)

			// act
			gotRecords, err := client.GetHeartrate(ctx, tt.date)

			// assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wantRecords := []metric.FitbitHeartrateRecord{
				{
					Time:  time.Date(2023, 1, 1, 1, 2, 3, 0, time.UTC),
					Value: 78,
				},
			}

			if len(gotRecords) != len(wantRecords) {
				t.Fatalf("unexpected record size: want %d, but got %d", len(wantRecords), len(gotRecords))
			}
			for i, wantRecord := range wantRecords {
				if wantRecord != gotRecords[i] {
					t.Fatalf("unexpected record: want %v, but got %v", wantRecord, gotRecords[i])
				}
			}
		})
	}
}
