package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/okta/okta-jwt-verifier-golang"
	"golang.org/x/oauth2"
)

type OktaService struct {
	config      *oauth2.Config
	issuer      string
	clientID    string
	jwtVerifier *jwtverifier.JwtVerifier
}

type OktaUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewOktaService(orgURL, clientID, clientSecret string, redirectURL string) (*OktaService, error) {
	toValidate := map[string]string{
		"clientId":     clientID,
		"issuer":      fmt.Sprintf("%s/oauth2/default", orgURL),
		"audience":     "api://default",
	}

	jwtVerifier := jwtverifier.JwtVerifier{
		Issuer:           toValidate["issuer"],
		ClaimsToValidate: toValidate,
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/default/v1/authorize", orgURL),
			TokenURL: fmt.Sprintf("%s/oauth2/default/v1/token", orgURL),
		},
	}

	return &OktaService{
		config:      config,
		issuer:      toValidate["issuer"],
		clientID:    clientID,
		jwtVerifier: &jwtVerifier,
	}, nil
}

func (s *OktaService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state)
}

func (s *OktaService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Exchange(ctx, code)
}

func (s *OktaService) ValidateToken(ctx context.Context, tokenString string) (*OktaUserInfo, error) {
	jwt, err := s.jwtVerifier.VerifyAccessToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	claims := jwt.Claims.(map[string]interface{})
	
	// Get user info from Okta
	userInfo := &OktaUserInfo{
		Sub:   claims["sub"].(string),
		Email: claims["email"].(string),
		Name:  claims["name"].(string),
	}

	return userInfo, nil
} 