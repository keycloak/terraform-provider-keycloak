package keycloak

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-version"

	"golang.org/x/net/publicsuffix"

	"github.com/hashicorp/go-retryablehttp"
)

type KeycloakClient struct {
	baseUrl           string
	authnUrl          string
	realm             string
	clientCredentials *ClientCredentials
	httpClient        *http.Client
	initialLogin      bool
	userAgent         string
	version           *version.Version
	additionalHeaders map[string]string
	debug             bool
	redHatSSO         bool
}

type ClientCredentials struct {
	ClientId     string
	ClientSecret string
	Username     string
	Password     string
	GrantType    string
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

const (
	apiUrl   = "/admin"
	tokenUrl = "%s/realms/%s/protocol/openid-connect/token"
)

// https://access.redhat.com/articles/2342881
var redHatSSO7VersionMap = map[int]string{
	6: "18.0.0",
	5: "15.0.6",
	4: "9.0.17",
}

func NewKeycloakClient(ctx context.Context, url, basePath, adminUrl, clientId, clientSecret, realm, username, password string, initialLogin bool, clientTimeout int, caCert string, tlsInsecureSkipVerify bool, userAgent string, redHatSSO bool, additionalHeaders map[string]string) (*KeycloakClient, error) {
	clientCredentials := &ClientCredentials{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
	if password != "" && username != "" {
		clientCredentials.Username = username
		clientCredentials.Password = password
		clientCredentials.GrantType = "password"
	} else if clientSecret != "" {
		clientCredentials.GrantType = "client_credentials"
	} else {
		if initialLogin {
			return nil, fmt.Errorf("must specify client id, username and password for password grant, or client id and secret for client credentials grant")
		} else {
			tflog.Warn(ctx, "missing required keycloak credentials, but proceeding anyways as initial_login is false")
		}
	}

	httpClient, err := newHttpClient(tlsInsecureSkipVerify, clientTimeout, caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %v", err)
	}

	baseUrl := url + basePath
	if adminUrl != "" {
		baseUrl = adminUrl + basePath
	}
	keycloakClient := KeycloakClient{
		baseUrl:           baseUrl,
		authnUrl:          url + basePath,
		clientCredentials: clientCredentials,
		httpClient:        httpClient,
		initialLogin:      initialLogin,
		realm:             realm,
		userAgent:         userAgent,
		redHatSSO:         redHatSSO,
		additionalHeaders: additionalHeaders,
	}

	if keycloakClient.initialLogin {
		err = keycloakClient.login(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to perform initial login to Keycloak: %v", err)
		}
	}

	if tfLog, ok := os.LookupEnv("TF_LOG"); ok {
		if tfLog == "DEBUG" {
			keycloakClient.debug = true
		}
	}

	return &keycloakClient, nil
}

func (keycloakClient *KeycloakClient) login(ctx context.Context) error {
	accessTokenUrl := fmt.Sprintf(tokenUrl, keycloakClient.authnUrl, keycloakClient.realm)
	accessTokenData := keycloakClient.getAuthenticationFormData()

	tflog.Debug(ctx, "Login request", map[string]interface{}{
		"request": accessTokenData.Encode(),
	})

	accessTokenRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, accessTokenUrl, strings.NewReader(accessTokenData.Encode()))
	if err != nil {
		return err
	}

	for header, value := range keycloakClient.additionalHeaders {
		accessTokenRequest.Header.Set(header, value)
	}

	accessTokenRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if keycloakClient.userAgent != "" {
		accessTokenRequest.Header.Set("User-Agent", keycloakClient.userAgent)
	}

	accessTokenResponse, err := keycloakClient.httpClient.Do(accessTokenRequest)
	if err != nil {
		return err
	}
	if accessTokenResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending POST request to %s: %s", accessTokenUrl, accessTokenResponse.Status)
	}

	defer accessTokenResponse.Body.Close()

	body, _ := io.ReadAll(accessTokenResponse.Body)

	tflog.Debug(ctx, "Login response", map[string]interface{}{
		"response": string(body),
	})

	var clientCredentials ClientCredentials
	err = json.Unmarshal(body, &clientCredentials)
	if err != nil {
		return err
	}

	keycloakClient.clientCredentials.AccessToken = clientCredentials.AccessToken
	keycloakClient.clientCredentials.RefreshToken = clientCredentials.RefreshToken
	keycloakClient.clientCredentials.TokenType = clientCredentials.TokenType

	info, err := keycloakClient.GetServerInfo(ctx)
	if err != nil {
		return err
	}

	serverVersion := info.SystemInfo.ServerVersion
	if strings.Contains(serverVersion, ".GA") {
		serverVersion = strings.ReplaceAll(info.SystemInfo.ServerVersion, ".GA", "")
	} else {
		regex, err := regexp.Compile(`\.redhat-\w+`)

		if err != nil {
			fmt.Println("Error compiling regex:", err)
			return err
		}

		// Check if the pattern is found in serverVersion
		if regex.MatchString(serverVersion) {
			// Replace the matched pattern with an empty string
			serverVersion = regex.ReplaceAllString(serverVersion, "")
		}
	}

	v, err := version.NewVersion(serverVersion)
	if err != nil {
		return err
	}

	if keycloakClient.redHatSSO {
		keycloakVersion, err := version.NewVersion(redHatSSO7VersionMap[v.Segments()[1]])
		if err != nil {
			return err
		}

		keycloakClient.version = keycloakVersion
	} else {
		keycloakClient.version = v
	}

	return nil
}

func (keycloakClient *KeycloakClient) Refresh(ctx context.Context) error {
	refreshTokenUrl := fmt.Sprintf(tokenUrl, keycloakClient.authnUrl, keycloakClient.realm)
	refreshTokenData := keycloakClient.getAuthenticationFormData()

	tflog.Debug(ctx, "Refresh request", map[string]interface{}{
		"request": refreshTokenData.Encode(),
	})

	refreshTokenRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, refreshTokenUrl, strings.NewReader(refreshTokenData.Encode()))
	if err != nil {
		return err
	}

	for header, value := range keycloakClient.additionalHeaders {
		refreshTokenRequest.Header.Set(header, value)
	}

	refreshTokenRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if keycloakClient.userAgent != "" {
		refreshTokenRequest.Header.Set("User-Agent", keycloakClient.userAgent)
	}

	refreshTokenResponse, err := keycloakClient.httpClient.Do(refreshTokenRequest)
	if err != nil {
		return err
	}

	defer refreshTokenResponse.Body.Close()

	body, _ := io.ReadAll(refreshTokenResponse.Body)

	tflog.Debug(ctx, "Refresh response", map[string]interface{}{
		"response": string(body),
	})

	// Handle 401 "User or client no longer has role permissions for client key" until I better understand why that happens in the first place
	if refreshTokenResponse.StatusCode == http.StatusBadRequest {
		tflog.Debug(ctx, "Unexpected 400, attempting to log in again")

		return keycloakClient.login(ctx)
	}

	var clientCredentials ClientCredentials
	err = json.Unmarshal(body, &clientCredentials)
	if err != nil {
		return err
	}

	keycloakClient.clientCredentials.AccessToken = clientCredentials.AccessToken
	keycloakClient.clientCredentials.RefreshToken = clientCredentials.RefreshToken
	keycloakClient.clientCredentials.TokenType = clientCredentials.TokenType

	return nil
}

func (keycloakClient *KeycloakClient) getAuthenticationFormData() url.Values {
	authenticationFormData := url.Values{}
	authenticationFormData.Set("client_id", keycloakClient.clientCredentials.ClientId)
	authenticationFormData.Set("grant_type", keycloakClient.clientCredentials.GrantType)

	if keycloakClient.clientCredentials.GrantType == "password" {
		authenticationFormData.Set("username", keycloakClient.clientCredentials.Username)
		authenticationFormData.Set("password", keycloakClient.clientCredentials.Password)

		if keycloakClient.clientCredentials.ClientSecret != "" {
			authenticationFormData.Set("client_secret", keycloakClient.clientCredentials.ClientSecret)
		}

	} else if keycloakClient.clientCredentials.GrantType == "client_credentials" {
		authenticationFormData.Set("client_secret", keycloakClient.clientCredentials.ClientSecret)
	}

	return authenticationFormData
}

func (keycloakClient *KeycloakClient) addRequestHeaders(request *http.Request) {
	tokenType := keycloakClient.clientCredentials.TokenType
	accessToken := keycloakClient.clientCredentials.AccessToken

	for header, value := range keycloakClient.additionalHeaders {
		request.Header.Set(header, value)
	}

	request.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenType, accessToken))
	request.Header.Set("Accept", "application/json")

	if keycloakClient.userAgent != "" {
		request.Header.Set("User-Agent", keycloakClient.userAgent)
	}

	if request.Method == http.MethodPost || request.Method == http.MethodPut || request.Method == http.MethodDelete {
		request.Header.Set("Content-type", "application/json")
	}
}

/*
*
Sends an HTTP request and refreshes credentials on 403 or 401 errors
*/
func (keycloakClient *KeycloakClient) sendRequest(ctx context.Context, request *http.Request, body []byte) ([]byte, string, error) {
	if !keycloakClient.initialLogin {
		keycloakClient.initialLogin = true
		err := keycloakClient.login(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("error logging in: %s", err)
		}
	}

	requestMethod := request.Method
	requestPath := request.URL.Path

	requestLogArgs := map[string]interface{}{
		"method": requestMethod,
		"path":   requestPath,
	}

	if body != nil {
		request.Body = io.NopCloser(bytes.NewReader(body))
		requestLogArgs["body"] = string(body)
	}

	tflog.Debug(ctx, "Sending request", requestLogArgs)

	keycloakClient.addRequestHeaders(request)

	response, err := keycloakClient.httpClient.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("error sending request: %v", err)
	}

	// Unauthorized: Token could have expired
	// Forbidden: After creating a realm, following GETs for the realm return 403 until you refresh
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		tflog.Debug(ctx, "Got unexpected response, attempting refresh", map[string]interface{}{
			"status": response.Status,
		})

		err := keycloakClient.Refresh(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("error refreshing credentials: %s", err)
		}

		keycloakClient.addRequestHeaders(request)

		if body != nil {
			request.Body = io.NopCloser(bytes.NewReader(body))
		}
		response, err = keycloakClient.httpClient.Do(request)
		if err != nil {
			return nil, "", fmt.Errorf("error sending request after refresh: %v", err)
		}
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	responseLogArgs := map[string]interface{}{
		"status": response.Status,
	}

	if len(responseBody) != 0 && request.URL.Path != "/auth/admin/serverinfo" {
		responseLogArgs["body"] = string(responseBody)
	}

	tflog.Debug(ctx, "Received response", responseLogArgs)

	if response.StatusCode >= 400 {
		errorMessage := fmt.Sprintf("error sending %s request to %s: %s.", request.Method, request.URL.Path, response.Status)

		if len(responseBody) != 0 {
			errorMessage = fmt.Sprintf("%s Response body: %s", errorMessage, responseBody)
		}

		return nil, "", &ApiError{
			Code:    response.StatusCode,
			Message: errorMessage,
		}
	}

	return responseBody, response.Header.Get("Location"), nil
}

func (keycloakClient *KeycloakClient) get(ctx context.Context, path string, resource interface{}, params map[string]string) error {
	body, err := keycloakClient.getRaw(ctx, path, params)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, resource)
}

func (keycloakClient *KeycloakClient) getRaw(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	resourceUrl := keycloakClient.baseUrl + apiUrl + path

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceUrl, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		query := url.Values{}
		for k, v := range params {
			query.Add(k, v)
		}
		request.URL.RawQuery = query.Encode()
	}

	body, _, err := keycloakClient.sendRequest(ctx, request, nil)
	return body, err
}

func (keycloakClient *KeycloakClient) sendRaw(ctx context.Context, path string, requestBody []byte) ([]byte, error) {
	resourceUrl := keycloakClient.baseUrl + apiUrl + path

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, resourceUrl, nil)
	if err != nil {
		return nil, err
	}

	body, _, err := keycloakClient.sendRequest(ctx, request, requestBody)

	return body, err
}

func (keycloakClient *KeycloakClient) post(ctx context.Context, path string, requestBody interface{}) ([]byte, string, error) {
	resourceUrl := keycloakClient.baseUrl + apiUrl + path

	payload, err := keycloakClient.marshal(requestBody)
	if err != nil {
		return nil, "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, resourceUrl, nil)
	if err != nil {
		return nil, "", err
	}

	body, location, err := keycloakClient.sendRequest(ctx, request, payload)

	return body, location, err
}

func (keycloakClient *KeycloakClient) put(ctx context.Context, path string, requestBody interface{}) error {
	resourceUrl := keycloakClient.baseUrl + apiUrl + path

	payload, err := keycloakClient.marshal(requestBody)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, resourceUrl, nil)
	if err != nil {
		return err
	}

	_, _, err = keycloakClient.sendRequest(ctx, request, payload)

	return err
}

func (keycloakClient *KeycloakClient) delete(ctx context.Context, path string, requestBody interface{}) error {
	resourceUrl := keycloakClient.baseUrl + apiUrl + path

	var (
		payload []byte
		err     error
	)

	if requestBody != nil {
		payload, err = keycloakClient.marshal(requestBody)
		if err != nil {
			return err
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, resourceUrl, nil)
	if err != nil {
		return err
	}

	_, _, err = keycloakClient.sendRequest(ctx, request, payload)

	return err
}

func (keycloakClient *KeycloakClient) marshal(body interface{}) ([]byte, error) {
	if keycloakClient.debug {
		return json.MarshalIndent(body, "", "    ")
	}

	return json.Marshal(body)
}

func newHttpClient(tlsInsecureSkipVerify bool, clientTimeout int, caCert string) (*http.Client, error) {
	cookieJar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsInsecureSkipVerify},
		Proxy:           http.ProxyFromEnvironment,
	}

	if caCert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		transport.TLSClientConfig.RootCAs = caCertPool
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 1
	retryClient.RetryWaitMin = time.Second * 1
	retryClient.RetryWaitMax = time.Second * 3

	httpClient := retryClient.StandardClient()
	httpClient.Timeout = time.Second * time.Duration(clientTimeout)
	httpClient.Transport = transport
	httpClient.Jar = cookieJar

	return httpClient, nil
}
