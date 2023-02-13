package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/dropbox/godropbox/errors"
	go_github "github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

const (
	// Google Sheets required scopes
	googleContactsReadOnlyScope = "https://www.googleapis.com/auth/contacts.readonly"
	googleSheetsScope           = "https://www.googleapis.com/auth/spreadsheets"
	googleDriveScope            = "https://www.googleapis.com/auth/drive"
)

type OAuthConfig struct {
	Token      *oauth2.Token  `json:"token"`
	OAuth2Conf *oauth2.Config `json:"oauth2_conf"`

	// PublicConf contains any additional public information that should be stored in the Config,
	// such as an email address or userId.
	PublicConf map[string]string `json:"public_conf"`
}

// NewOAuthConfig creates a new Config that uses OAuth 2.0 for authentication for
// the service specified. It accepts the clientId and clientSecret to use when requesting
// an access token from the relevant authorization endpoint. It also accepts the redirectURL,
// which should be the same as the one used when retrieving the authorizationCode.
// The authorizationCode is the grant type that is used in exchange for the access token.
func NewOAuthConfig(
	ctx context.Context,
	service shared.Service,
	clientId string,
	clientSecret string,
	redirectURL string,
	authorizationCode string,
) (Config, error) {
	oauthConf, err := newOAuth2Config(service, clientId, clientSecret, redirectURL)
	if err != nil {
		return nil, err
	}

	token, err := oauthConf.Exchange(ctx, authorizationCode)
	if err != nil {
		return nil, err
	}

	publicConf, err := getPublicConfig(ctx, service, token)
	if err != nil {
		return nil, err
	}

	return &OAuthConfig{
		Token:      token,
		OAuth2Conf: oauthConf,
		PublicConf: publicConf,
	}, nil
}

// newOAuth2Config returns an *oauth2.Config for the service specified using
// clientId and clientSecret. It returns an error, if any.
func newOAuth2Config(
	service shared.Service,
	clientId string,
	clientSecret string,
	redirectURL string,
) (*oauth2.Config, error) {
	oauthConf := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}

	switch service {
	case shared.GoogleSheets:
		oauthConf.Endpoint = google.Endpoint
		oauthConf.Scopes = []string{googleContactsReadOnlyScope, googleSheetsScope, googleDriveScope}
	case shared.Github:
		oauthConf.Endpoint = github.Endpoint
	case shared.Salesforce:
		oauthConf.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://login.salesforce.com/services/oauth2/token",
		}
	default:
		return nil, errors.Newf("OAuth2 Config is not supported for: %v", service)
	}

	return oauthConf, nil
}

// getPublicConfig fetches additional information about the end user, such as email.
// Returns a map[string]string of information and an error, if any.
func getPublicConfig(ctx context.Context, service shared.Service, token *oauth2.Token) (map[string]string, error) {
	switch service {
	case shared.GoogleSheets:
		return getGoogleSheetsPublicConfig(token)
	case shared.Github:
		return getGithubPublicConfig(ctx, token)
	case shared.Salesforce:
		return getSalesforcePublicConfig(token)
	default:
		return nil, errors.Newf("OAuth 2 Config is not supported for: %v", service)
	}
}

func getGoogleSheetsPublicConfig(token *oauth2.Token) (map[string]string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/userinfo/v2/me?access_token=%s", token.AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := struct {
		Email string `json:"email"`
	}{}
	if err := json.Unmarshal(content, &s); err != nil {
		return nil, err
	}

	return map[string]string{
		"email": s.Email,
	}, nil
}

func GithubUserIdToString(id int64) string {
	return strconv.FormatInt(id, 16)
}

func getSalesforcePublicConfig(token *oauth2.Token) (map[string]string, error) {
	instanceURL := token.Extra("instance_url")
	if instanceURL == nil {
		return nil, errors.New("Unable to get Salesforce instance URL.")
	}

	return map[string]string{
		"instance_url": fmt.Sprintf("%v", instanceURL), // Cast interface type to string
	}, nil
}

func getGithubPublicConfig(ctx context.Context, token *oauth2.Token) (map[string]string, error) {
	tokenSource := oauth2.StaticTokenSource(token)
	httpClient := oauth2.NewClient(ctx, tokenSource)
	client := go_github.NewClient(httpClient)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}
	return map[string]string{
		GithubUserField: GithubUserIdToString(*user.ID),
	}, nil
}

func (oc *OAuthConfig) getType() configType {
	return oauthConfigType
}

func (oc *OAuthConfig) Marshal() ([]byte, error) {
	marshalMap := make(map[string]string, len(oc.PublicConf)+1)
	// Connectors need the oauth access token
	marshalMap["access_token"] = oc.Token.AccessToken

	// Connectors may also need public config
	for k, v := range oc.PublicConf {
		marshalMap[k] = v
	}
	return json.Marshal(marshalMap)
}

func (oc *OAuthConfig) PublicConfig() map[string]string {
	// All public information is stored in oc.PublicConf
	return oc.PublicConf
}

func (oc *OAuthConfig) Refresh(ctx context.Context) (bool, error) {
	if oc.Token.Valid() {
		// Token is valid, do nothing
		return false, nil
	}

	// This is based off of this answer: https://stackoverflow.com/a/46487481
	tokenSource := oc.OAuth2Conf.TokenSource(ctx, oc.Token)
	newToken, err := tokenSource.Token() // This call refreshes the token.
	if err != nil {
		return false, err
	}

	oc.Token = newToken
	return true, nil
}
