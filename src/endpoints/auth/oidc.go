package auth

import (
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type Oidc struct {
	oidcConfig *oidc.Config
	verifier   *oidc.IDTokenVerifier
	config     oauth2.Config
}

// CreateOidc will create an oidc object, which can be used as authorisation endpoint
func (o *Oidc) CreateOidc(clientID, clientSecret, redirectURL, endpoint string) error {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, endpoint)
	if err != nil {
		return err
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}

	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "groups", "roles", "profile"},
	}

	o.oidcConfig = oidcConfig
	o.verifier = verifier
	o.config = config

	return nil
}

func (o Oidc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func User() (string, []string, []string, error) {
	return "", nil, nil, nil
}
