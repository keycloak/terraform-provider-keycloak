package keycloak

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/keycloak/terraform-provider-keycloak/mutex"

	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/net/publicsuffix"

	"github.com/hashicorp/go-retryablehttp"
)

type KeycloakClient struct {
	baseUrl             string
	authUrl             string
	realm               string
	clientCredentials   *ClientCredentials
	httpClient          *http.Client
	initialLogin        bool
	userAgent           string
	version             *version.Version
	additionalHeaders   map[string]string
	debug               bool
	redHatSSO           bool
	accessTokenProvided bool
	Mutex               *mutex.KeyValue
}

type ClientCredentials struct {
	ClientId      string
	ClientSecret  string
	JWTSigningKey string
	JWTSigningAlg string
	Username      string
	Password      string
	GrantType     string
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	TokenType     string `json:"token_type"`
}

const (
	apiUrl    = "/admin"
	issuerUrl = "%s/realms/%s"
	tokenUrl  = "%s/realms/%s/protocol/openid-connect/token"
)

// https://access.redhat.com/articles/2342881
var redHatSSO7VersionMap = map[int]string{
	6: "18.0.0",
	5: "15.0.6",
	4: "9.0.17",
}

func NewKeycloakClient(ctx context.Context, url, basePath, adminUrl, clientId, clientSecret, realm, username, password, accessToken, jwtSigningAlg, jwtSigningKey string, initialLogin bool, clientTimeout int, caCert string, tlsInsecureSkipVerify bool, tlsClientCert string, tlsClientPrivateKey string, userAgent string, redHatSSO bool, additionalHeaders map[string]string) (*KeycloakClient, error) {
	clientCredentials := &ClientCredentials{
		ClientId:      clientId,
		ClientSecret:  clientSecret,
		JWTSigningKey: jwtSigningKey,
		JWTSigningAlg: jwtSigningAlg,
	}

	if password != "" && username != "" {
		clientCredentials.Username = username
		clientCredentials.Password = password
		clientCredentials.GrantType = "password"
	} else if clientSecret != "" || jwtSigningKey != "" {
		clientCredentials.GrantType = "client_credentials"
	} else if accessToken != "" {
		clientCredentials.AccessToken = accessToken
		clientCredentials.TokenType = "bearer"
	} else {
		if initialLogin {
			return nil, fmt.Errorf("must specify client id, username and password for password grant, either client id and client secret or JWT Signing Key for client credentials grant")
		} else {
			tflog.Warn(ctx, "missing required keycloak credentials, but proceeding anyways as initial_login is false")
		}
	}

	httpClient, err := newHttpClient(tlsInsecureSkipVerify, clientTimeout, caCert, tlsClientCert, tlsClientPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %v", err)
	}

	authUrl := url + basePath
	baseUrl := authUrl
	if adminUrl != "" {
		baseUrl = adminUrl + basePath
	}

	keycloakClient := KeycloakClient{
		baseUrl:             baseUrl,
		authUrl:             authUrl,
		clientCredentials:   clientCredentials,
		httpClient:          httpClient,
		initialLogin:        initialLogin,
		realm:               realm,
		userAgent:           userAgent,
		redHatSSO:           redHatSSO,
		additionalHeaders:   additionalHeaders,
		accessTokenProvided: accessToken != "",
		Mutex:               mutex.New(),
	}

	if accessToken == "" && keycloakClient.initialLogin {
		err = keycloakClient.login(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to perform initial login to Keycloak: %v", err)
		}
	}

	if tfLog, ok := os.LookupEnv("TF_LOG"); ok && tfLog == "DEBUG" {
		keycloakClient.debug = true
	}

	return &keycloakClient, nil
}

func (keycloakClient *KeycloakClient) login(ctx context.Context) error {

	if !keycloakClient.accessTokenProvided {
		accessTokenUrl := fmt.Sprintf(tokenUrl, keycloakClient.authUrl, keycloakClient.realm)
		accessTokenData, err := keycloakClient.getAuthenticationFormData(ctx, accessTokenUrl)
		if err != nil {
			return err
		}

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
	} else {
		tflog.Debug(ctx, "Using provided access_token", map[string]interface{}{
			"access_token": keycloakClient.clientCredentials.AccessToken,
		})
	}

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

	if keycloakClient.accessTokenProvided {
		// If an access_token was provided, we skip refresh
		return nil
	}

	refreshTokenUrl := fmt.Sprintf(tokenUrl, keycloakClient.authUrl, keycloakClient.realm)
	refreshTokenData, err := keycloakClient.getAuthenticationFormData(ctx, refreshTokenUrl)
	if err != nil {
		return err
	}

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

func (keycloakClient *KeycloakClient) getAuthenticationFormData(ctx context.Context, kc_url string) (url.Values, error) {
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
		if keycloakClient.clientCredentials.JWTSigningKey != "" {
			signedJWT, err := newSignedJWT(
				ctx,
				fmt.Sprintf(issuerUrl, keycloakClient.baseUrl, keycloakClient.realm),
				keycloakClient.clientCredentials.ClientId,
				keycloakClient.clientCredentials.JWTSigningAlg,
				keycloakClient.clientCredentials.JWTSigningKey,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create signed JWT: %v", err)
			}
			authenticationFormData.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
			authenticationFormData.Set("client_assertion", signedJWT)
		} else {
			authenticationFormData.Set("client_secret", keycloakClient.clientCredentials.ClientSecret)
		}

	}

	return authenticationFormData, nil
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

	if request.Header.Get("Content-type") == "" && (request.Method == http.MethodPost || request.Method == http.MethodPut || request.Method == http.MethodDelete) {
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
	defer response.Body.Close()

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
		defer response.Body.Close()
	}

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

func RetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return true, ctx.Err()
	}

	// 429 Too Many Requests is recoverable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	// Check the response code. We retry on 500-range responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side. This will catch
	// invalid response codes as well, like 0 and 999.
	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true, nil
	}

	return false, nil
}

func newHttpClient(tlsInsecureSkipVerify bool, clientTimeout int, caCert string, tlsClientCert string, tlsClientPrivateKey string) (*http.Client, error) {
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
	transport.MaxIdleConnsPerHost = 100

	if caCert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		transport.TLSClientConfig.RootCAs = caCertPool
	}

	if tlsClientCert != "" && tlsClientPrivateKey != "" {
		clientKeyPairCert, err := tls.X509KeyPair([]byte(tlsClientCert), []byte(tlsClientPrivateKey))
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{clientKeyPairCert}
	}

	retryClient := retryablehttp.NewClient()
	retryClient.CheckRetry = RetryPolicy
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = time.Second * 1
	retryClient.RetryWaitMax = time.Second * 60

	httpClient := retryClient.StandardClient()
	httpClient.Timeout = time.Second * time.Duration(clientTimeout)
	httpClient.Transport = transport
	httpClient.Jar = cookieJar

	return httpClient, nil
}

func newSignedJWT(ctx context.Context, url, clientId, alg, jwtSigningKey string) (string, error) {
	// Create the Claims
	jti, err := uuid.GenerateUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT ID: %v", err)
	}

	claims := jwt.MapClaims{
		"jti": jti,
		"iss": clientId,
		"sub": clientId,
		"aud": url,
		"exp": jwt.NewNumericDate(time.Now().Add(time.Second * 60)),
		"iat": jwt.NewNumericDate(time.Now()),
	}

	signingMethod := jwt.GetSigningMethod(alg)
	if signingMethod == nil {
		return "", fmt.Errorf("unsupported signing method: %s", alg)
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.GetSigningMethod(alg), claims)

	var key any
	if _, isRsa := signingMethod.(*jwt.SigningMethodRSA); isRsa {
		key, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(jwtSigningKey))
	} else if _, isEcdsa := signingMethod.(*jwt.SigningMethodECDSA); isEcdsa {
		key, err = jwt.ParseECPrivateKeyFromPEM([]byte(jwtSigningKey))
	} else if _, isEd25519 := signingMethod.(*jwt.SigningMethodEd25519); isEd25519 {
		key, err = jwt.ParseEdPrivateKeyFromPEM([]byte(jwtSigningKey))
	} else {
		err = fmt.Errorf("unsupported signing method: %s", signingMethod)
	}

	if err != nil {
		return "", err
	}
	tokenString, err := token.SignedString(key)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	jwtClientAssertionArgs := map[string]any{
		"jti": jti,
	}
	tflog.Debug(ctx, "Generated client_assertion", jwtClientAssertionArgs)

	return tokenString, nil
}

// Expose the underlying http client for tests
func (kc *KeycloakClient) GetHttpClient() *http.Client {
	return kc.httpClient
}
