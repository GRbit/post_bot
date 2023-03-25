package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	credsPath = "credentials.json"
	tokenPath = "token.json"
)

// Create OAuth2 client ID and save downloaded JSON into credentials.json file in program working folder
// https://developers.google.com/workspace/guides/create-credentials#api-key

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		log.Info().Msg("No token found, getting token from web")

		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, xerrors.Errorf("getting token from web: %w", err)
		}

		if err := saveToken(tokenPath, tok); err != nil {
			return nil, xerrors.Errorf("saving token: %w", err)
		}
	} else {
		log.Info().Str("path", tokenPath).Msg("token loadded from file")
	}

	return config.Client(ctx, tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, xerrors.Errorf("reading authorization code: %w", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, xerrors.Errorf("exchanging auth code for token: %w", err)
	}

	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	log.Info().Str("path", path).Msg("saving token to file")

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return xerrors.Errorf("saving oauth token to '%s': %w", path, err)
	}

	defer f.Close()

	if err = json.NewEncoder(f).Encode(token); err != nil {
		return xerrors.Errorf("encoding token: %w", err)
	}

	return nil
}

func newSheetsService(ctx context.Context) (*sheets.Service, error) {
	b, err := os.ReadFile(credsPath)
	if err != nil {
		return nil, xerrors.Errorf("reading client secret file: %w", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		return nil, xerrors.Errorf("parsing client secret file to config: %w", err)
	}

	client, err := getClient(ctx, config)
	if err != nil {
		return nil, xerrors.Errorf("retrieving HTTP client: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, xerrors.Errorf("retrieving Sheets client: %w", err)
	}

	return srv, nil
}
