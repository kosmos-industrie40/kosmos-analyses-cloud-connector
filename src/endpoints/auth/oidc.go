package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"k8s.io/klog"
)

type Auth interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type oidcAuth struct {
	oidcConfig    *oidc.Config
	verifier      *oidc.IDTokenVerifier
	config        oauth2.Config
	state         string
	regexBase     *regexp.Regexp
	regexCallback *regexp.Regexp
	helper        Helper
	generator     TokenGenerate
}

func NewOidcAuth(userMgmt, basePath, clientSecret, clientId, serverAddress string, helper Helper) (Auth, error) {
	ctx := context.Background()
	issuer := userMgmt
	klog.Infof("issuer url: %s", issuer)

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return oidcAuth{}, fmt.Errorf("cannot create auth provider: %s", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: clientId,
	}

	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  fmt.Sprintf("%s/%s/callback", serverAddress, basePath),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	state := uuid.New().String()

	regexBase := regexp.MustCompile(fmt.Sprintf("/%s[/]?$", basePath))
	regexCallback := regexp.MustCompile(fmt.Sprintf("/%s/callback[/]?$", basePath))

	oidcDat := oidcAuth{
		oidcConfig:    oidcConfig,
		verifier:      verifier,
		config:        config,
		state:         state,
		regexBase:     regexBase,
		regexCallback: regexCallback,
		generator:     NewTokenGeneratorUuid(),
		helper:        helper,
	}

	klog.Infof("using basePath: %s and %s/callback as registered endpoints", basePath, basePath)

	return oidcDat, nil
}

func (o oidcAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.Infof("request url: %s", r.URL.Path)
	token := r.Header.Get("token")
	if token != "" {
		o.handleWithToken(w, r)
		return
	}

	url := r.URL.Path
	if o.regexBase.MatchString(url) {
		o.handleBase(w, r)
		return
	}

	if o.regexCallback.MatchString(url) {
		o.handleCallback(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (o oidcAuth) handleBase(w http.ResponseWriter, r *http.Request) {
	klog.Infof("receive request %s in base", r.Method)
	switch r.Method {
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case http.MethodGet:
		fallthrough
	case http.MethodPost:
		http.Redirect(w, r, o.config.AuthCodeURL(o.state), http.StatusTemporaryRedirect)
		return
	}
}

func (o oidcAuth) handleCallback(w http.ResponseWriter, r *http.Request) {
	klog.Infof("receive request %s in callback", r.Method)
	switch r.Method {
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case http.MethodGet:
		if r.URL.Query().Get("state") != o.state {
			klog.Errorf("state did not match")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		oauth2Token, err := o.config.Exchange(context.Background(), r.URL.Query().Get("code"))
		if err != nil {
			klog.Errorf("Failed to exchange token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			klog.Errorf("no id_token field in oauth2 token: %v", oauth2Token.Extra("id_token"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		idToken, err := o.verifier.Verify(context.Background(), rawIDToken)
		if err != nil {
			klog.Errorf("Failed to verify ID Token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// the access token, can be used as authorisation token
		// in this application we didn't use it, because it should
		// be easy possible to add other authentication mechanism
		//oauth2Token.AccessToken = "*REDACTED*"

		var claims struct {
			Groups []string `json:"groups"`
			Roles  []string `json:"roles"`
		}

		if err := idToken.Claims(&claims); err != nil {
			klog.Errorf("cannot get id claims: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		klog.Infof("groups: %v", strings.Join(claims.Groups, ", "))
		klog.Infof("groups: %v", strings.Join(claims.Roles, ", "))

		token := struct {
			Token string    `json:"token"`
			Valid time.Time `json:"valid"`
		}{
			o.generator.Generate(),
			oauth2Token.Expiry,
		}

		klog.V(2).Infof("claim goups is: %s", claims.Groups)
		if err := o.helper.CreateSession(token.Token, claims.Groups, claims.Roles, idToken.Expiry); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf("cannot create session: %s", err)
			return
		}

		data, err := json.Marshal(token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf("cannot marshal token: %s", err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if _, err := w.Write(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			klog.Errorf("cannot send token: %s", err)
			return
		}
	}
}

func (o oidcAuth) handleWithToken(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		klog.Infof("receive DELETE request with token")
		if err := o.helper.DeleteSession(r.Header.Get("token")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		klog.Infof("receive request %s with token", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
