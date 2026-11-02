package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

type Key struct {
	Algorithm        *string `json:"algorithm,omitempty"`
	Certificate      *string `json:"certificate,omitempty"`
	ProviderId       *string `json:"providerId,omitempty"`
	ProviderPriority *int    `json:"providerPriority,omitempty"`
	PublicKey        *string `json:"publicKey,omitempty"`
	Kid              *string `json:"kid,omitempty"`
	Status           *string `json:"status,omitempty"`
	Type             *string `json:"type,omitempty"`
}

type Keys struct {
	Keys []Key `json:"keys"`
}

type Realm struct {
	Id                   string `json:"id,omitempty"`
	Realm                string `json:"realm"`
	Enabled              bool   `json:"enabled"`
	DisplayName          string `json:"displayName"`
	DisplayNameHtml      string `json:"displayNameHtml"`
	UserManagedAccess    bool   `json:"userManagedAccessAllowed"`
	OrganizationsEnabled bool   `json:"organizationsEnabled"`

	// Login Config
	RegistrationAllowed         bool   `json:"registrationAllowed"`
	RegistrationEmailAsUsername bool   `json:"registrationEmailAsUsername"`
	EditUsernameAllowed         bool   `json:"editUsernameAllowed"`
	ResetPasswordAllowed        bool   `json:"resetPasswordAllowed"`
	RememberMe                  bool   `json:"rememberMe"`
	VerifyEmail                 bool   `json:"verifyEmail"`
	LoginWithEmailAllowed       bool   `json:"loginWithEmailAllowed"`
	DuplicateEmailsAllowed      bool   `json:"duplicateEmailsAllowed"`
	SslRequired                 string `json:"sslRequired,omitempty"`

	//SMTP Server
	SmtpServer SmtpServer `json:"smtpServer"`

	// Themes
	LoginTheme   string `json:"loginTheme,omitempty"`
	AccountTheme string `json:"accountTheme,omitempty"`
	AdminTheme   string `json:"adminTheme,omitempty"`
	EmailTheme   string `json:"emailTheme,omitempty"`

	// Tokens
	DefaultSignatureAlgorithm           string `json:"defaultSignatureAlgorithm"`
	RevokeRefreshToken                  bool   `json:"revokeRefreshToken"`
	RefreshTokenMaxReuse                int    `json:"refreshTokenMaxReuse"`
	SsoSessionIdleTimeout               int    `json:"ssoSessionIdleTimeout,omitempty"`
	SsoSessionMaxLifespan               int    `json:"ssoSessionMaxLifespan,omitempty"`
	SsoSessionIdleTimeoutRememberMe     int    `json:"ssoSessionIdleTimeoutRememberMe,omitempty"`
	SsoSessionMaxLifespanRememberMe     int    `json:"ssoSessionMaxLifespanRememberMe,omitempty"`
	OfflineSessionIdleTimeout           int    `json:"offlineSessionIdleTimeout,omitempty"`
	OfflineSessionMaxLifespan           int    `json:"offlineSessionMaxLifespan,omitempty"`
	OfflineSessionMaxLifespanEnabled    bool   `json:"offlineSessionMaxLifespanEnabled,omitempty"`
	ClientSessionIdleTimeout            int    `json:"clientSessionIdleTimeout,omitempty"`
	ClientSessionMaxLifespan            int    `json:"clientSessionMaxLifespan,omitempty"`
	AccessTokenLifespan                 int    `json:"accessTokenLifespan,omitempty"`
	AccessTokenLifespanForImplicitFlow  int    `json:"accessTokenLifespanForImplicitFlow,omitempty"`
	AccessCodeLifespan                  int    `json:"accessCodeLifespan,omitempty"`
	AccessCodeLifespanLogin             int    `json:"accessCodeLifespanLogin,omitempty"`
	AccessCodeLifespanUserAction        int    `json:"accessCodeLifespanUserAction,omitempty"`
	ActionTokenGeneratedByUserLifespan  int    `json:"actionTokenGeneratedByUserLifespan,omitempty"`
	ActionTokenGeneratedByAdminLifespan int    `json:"actionTokenGeneratedByAdminLifespan,omitempty"`
	Oauth2DeviceCodeLifespan            int    `json:"oauth2DeviceCodeLifespan,omitempty"`
	Oauth2DevicePollingInterval         int    `json:"oauth2DevicePollingInterval,omitempty"`

	//internationalization
	InternationalizationEnabled bool     `json:"internationalizationEnabled"`
	SupportLocales              []string `json:"supportedLocales"`
	DefaultLocale               string   `json:"defaultLocale"`

	//extra attributes of a realm
	Attributes map[string]interface{} `json:"attributes"`

	// client-scope mapping defaults
	DefaultDefaultClientScopes  []string `json:"defaultDefaultClientScopes,omitempty"`
	DefaultOptionalClientScopes []string `json:"defaultOptionalClientScopes,omitempty"`

	BrowserSecurityHeaders BrowserSecurityHeaders `json:"browserSecurityHeaders"`

	BruteForceProtected          bool   `json:"bruteForceProtected"`
	PermanentLockout             bool   `json:"permanentLockout"`
	BruteForceStrategy           string `json:"bruteForceStrategy,omitempty"`       // Strategy to increase wait time
	MaxTemporaryLockouts         int    `json:"maxTemporaryLockouts,omitempty"`     //Max Temporary Lockouts
	FailureFactor                int    `json:"failureFactor"`                      //Max Login Failures
	MaxSecondaryAuthFailures     *int   `json:"maxSecondaryAuthFailures,omitempty"` //Max Secondary Auth Failures (requires Keycloak >= 26.6.0)
	WaitIncrementSeconds         int    `json:"waitIncrementSeconds"`
	QuickLoginCheckMilliSeconds  int    `json:"quickLoginCheckMilliSeconds"`
	MinimumQuickLoginWaitSeconds int    `json:"minimumQuickLoginWaitSeconds"`
	MaxFailureWaitSeconds        int    `json:"maxFailureWaitSeconds"` //Max Wait
	MaxDeltaTimeSeconds          int    `json:"maxDeltaTimeSeconds"`   //Failure Reset Time

	AdminPermissionsEnabled *bool `json:"adminPermissionsEnabled,omitempty"`
	AdminPermissionsClient  *struct {
		Id string `json:"id"`
	} `json:"adminPermissionsClient,omitempty"`

	PasswordPolicy string `json:"passwordPolicy"`

	//flow bindings
	BrowserFlow              *string `json:"browserFlow,omitempty"`
	RegistrationFlow         *string `json:"registrationFlow,omitempty"`
	DirectGrantFlow          *string `json:"directGrantFlow,omitempty"`
	ResetCredentialsFlow     *string `json:"resetCredentialsFlow,omitempty"`
	ClientAuthenticationFlow *string `json:"clientAuthenticationFlow,omitempty"`
	DockerAuthenticationFlow *string `json:"dockerAuthenticationFlow,omitempty"`
	FirstBrokerLoginFlow     *string `json:"firstBrokerLoginFlow,omitempty"`

	// OTP Policy
	OTPPolicyAlgorithm       string `json:"otpPolicyAlgorithm,omitempty"`
	OTPPolicyDigits          int    `json:"otpPolicyDigits,omitempty"`
	OTPPolicyInitialCounter  int    `json:"otpPolicyInitialCounter,omitempty"`
	OTPPolicyLookAheadWindow int    `json:"otpPolicyLookAheadWindow,omitempty"`
	OTPPolicyPeriod          int    `json:"otpPolicyPeriod,omitempty"`
	OTPPolicyCodeReusable    bool   `json:"otpPolicyCodeReusable"`
	OTPPolicyType            string `json:"otpPolicyType,omitempty"`

	// WebAuthn
	WebAuthnPolicyAcceptableAaguids               []string `json:"webAuthnPolicyAcceptableAaguids"`
	WebAuthnPolicyExtraOrigins                    []string `json:"webAuthnPolicyExtraOrigins,omitempty"`
	WebAuthnPolicyAttestationConveyancePreference string   `json:"webAuthnPolicyAttestationConveyancePreference"`
	WebAuthnPolicyAuthenticatorAttachment         string   `json:"webAuthnPolicyAuthenticatorAttachment"`
	WebAuthnPolicyAvoidSameAuthenticatorRegister  bool     `json:"webAuthnPolicyAvoidSameAuthenticatorRegister"`
	WebAuthnPolicyCreateTimeout                   int      `json:"webAuthnPolicyCreateTimeout"`
	// Deprecated: use WebAuthnPolicyDiscoverableCredential instead.
	WebAuthnPolicyRequireResidentKey          string   `json:"webAuthnPolicyRequireResidentKey"`
	WebAuthnPolicyDiscoverableCredential      string   `json:"webAuthnPolicyResidentKey,omitempty"`
	WebAuthnPolicyRpEntityName                string   `json:"webAuthnPolicyRpEntityName"`
	WebAuthnPolicyRpId                        string   `json:"webAuthnPolicyRpId"`
	WebAuthnPolicySignatureAlgorithms         []string `json:"webAuthnPolicySignatureAlgorithms"`
	WebAuthnPolicyUserVerificationRequirement string   `json:"webAuthnPolicyUserVerificationRequirement"`

	// WebAuthn Passwordless
	WebAuthnPolicyPasswordlessAcceptableAaguids               []string `json:"webAuthnPolicyPasswordlessAcceptableAaguids"`
	WebAuthnPolicyPasswordlessExtraOrigins                    []string `json:"webAuthnPolicyPasswordlessExtraOrigins,omitempty"`
	WebAuthnPolicyPasswordlessAttestationConveyancePreference string   `json:"webAuthnPolicyPasswordlessAttestationConveyancePreference"`
	WebAuthnPolicyPasswordlessAuthenticatorAttachment         string   `json:"webAuthnPolicyPasswordlessAuthenticatorAttachment"`
	WebAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister  bool     `json:"webAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister"`
	WebAuthnPolicyPasswordlessCreateTimeout                   int      `json:"webAuthnPolicyPasswordlessCreateTimeout"`
	// Deprecated: use WebAuthnPolicyPasswordlessDiscoverableCredential instead.
	WebAuthnPolicyPasswordlessRequireResidentKey          string   `json:"webAuthnPolicyPasswordlessRequireResidentKey"`
	WebAuthnPolicyPasswordlessDiscoverableCredential      string   `json:"webAuthnPolicyPasswordlessResidentKey,omitempty"`
	WebAuthnPolicyPasswordlessRpEntityName                string   `json:"webAuthnPolicyPasswordlessRpEntityName"`
	WebAuthnPolicyPasswordlessRpId                        string   `json:"webAuthnPolicyPasswordlessRpId"`
	WebAuthnPolicyPasswordlessSignatureAlgorithms         []string `json:"webAuthnPolicyPasswordlessSignatureAlgorithms"`
	WebAuthnPolicyPasswordlessUserVerificationRequirement string   `json:"webAuthnPolicyPasswordlessUserVerificationRequirement"`
	WebAuthnPolicyPasswordlessPasskeysEnabled             *bool    `json:"webAuthnPolicyPasswordlessPasskeysEnabled,omitempty"`

	// Roles
	DefaultRole *Role `json:"defaultRole,omitempty"`
}

type BrowserSecurityHeaders struct {
	ContentSecurityPolicy           string `json:"contentSecurityPolicy"`
	ContentSecurityPolicyReportOnly string `json:"contentSecurityPolicyReportOnly"`
	StrictTransportSecurity         string `json:"strictTransportSecurity"`
	XContentTypeOptions             string `json:"xContentTypeOptions"`
	XFrameOptions                   string `json:"xFrameOptions"`
	XRobotsTag                      string `json:"xRobotsTag"`
	XXSSProtection                  string `json:"xXSSProtection"`
	ReferrerPolicy                  string `json:"referrerPolicy"`
}

type SmtpServer struct {
	StartTls              types.KeycloakBoolQuoted `json:"starttls,omitempty"`
	Auth                  types.KeycloakBoolQuoted `json:"auth,omitempty"`
	Port                  string                   `json:"port,omitempty"`
	Host                  string                   `json:"host,omitempty"`
	ReplyTo               string                   `json:"replyTo,omitempty"`
	ReplyToDisplayName    string                   `json:"replyToDisplayName,omitempty"`
	From                  string                   `json:"from,omitempty"`
	FromDisplayName       string                   `json:"fromDisplayName,omitempty"`
	EnvelopeFrom          string                   `json:"envelopeFrom,omitempty"`
	Ssl                   types.KeycloakBoolQuoted `json:"ssl,omitempty"`
	AllowUtf8             types.KeycloakBoolQuoted `json:"allowutf8,omitempty"`
	User                  string                   `json:"user,omitempty"`
	Password              string                   `json:"password,omitempty"`
	AuthType              string                   `json:"authType,omitempty"`
	AuthTokenUrl          string                   `json:"authTokenUrl,omitempty"`
	AuthTokenClientId     string                   `json:"authTokenClientId,omitempty"`
	AuthTokenClientSecret string                   `json:"authTokenClientSecret,omitempty"`
	AuthTokenScope        string                   `json:"authTokenScope,omitempty"`
}

func (keycloakClient *KeycloakClient) NewRealm(ctx context.Context, realm *Realm) error {
	generation := keycloakClient.realmCache.currentGeneration()
	_, _, err := keycloakClient.post(ctx, "/realms", realm)
	if err == nil {
		// Keycloak 26.0 hides a new realm's id from tokens issued before it existed, so remember it
		keycloakClient.realmCache.remember(generation, realm.Id, realm.Realm)
	}

	return err
}

func (keycloakClient *KeycloakClient) GetRealm(ctx context.Context, name string) (*Realm, error) {
	var realm Realm

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s", name), &realm, nil)
	if err != nil {
		return nil, err
	}
	return &realm, nil
}

func (keycloakClient *KeycloakClient) GetRealms(ctx context.Context) ([]*Realm, error) {
	var realms []*Realm

	err := keycloakClient.get(ctx, "/realms", &realms, nil)
	if err != nil {
		return nil, err
	}

	return realms, nil
}

func (keycloakClient *KeycloakClient) GetRealmKeys(ctx context.Context, name string) (*Keys, error) {
	var keys Keys

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/keys", name), &keys, nil)
	if err != nil {
		return nil, err
	}

	return &keys, nil
}

func (keycloakClient *KeycloakClient) UpdateRealm(ctx context.Context, realm *Realm) error {
	return keycloakClient.UpdateRealmByName(ctx, realm.Realm, realm)
}

// UpdateRealmByName updates the realm that is currently called currentName. When realm.Realm
// differs from currentName, Keycloak renames the realm in place (its internal id stays the same).
func (keycloakClient *KeycloakClient) UpdateRealmByName(ctx context.Context, currentName string, realm *Realm) error {
	if currentName == realm.Realm {
		return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s", currentName), realm)
	}

	// forget before and after: a timeout may hide that the rename happened anyway
	keycloakClient.realmCache.forget()
	err := keycloakClient.put(ctx, fmt.Sprintf("/realms/%s", currentName), realm)
	generation := keycloakClient.realmCache.currentGeneration()
	keycloakClient.realmCache.forget()
	if err == nil {
		keycloakClient.realmCache.remember(generation+1, realm.Id, realm.Realm)
	}

	return err
}

// realmCacheTTL bounds how long what we learned about a realm is kept. A realm can be renamed outside
// this client, so nothing in the cache is trusted without asking Keycloak again.
const realmCacheTTL = 30 * time.Second

type realmCacheEntry struct {
	name string
	at   time.Time
}

// realmCache keeps two things about realms. Neither is ever returned as an answer:
//   - names: the name a realm had when we last saw it, by internal id. This is only a hint for where to look first,
//     and Keycloak has to confirm that the realm under that name still has the internal id.
//   - unreadable: names of realms for which Keycloak did not show an id to this account, so that the lookup is
//     not repeated for every resource.
//
// forget bumps the generation, so that a lookup that started before a rename cannot store what it saw afterwards.
type realmCache struct {
	mu         sync.Mutex
	generation uint64
	names      map[string]realmCacheEntry // internal id -> name
	unreadable map[string]time.Time       // name -> when
}

func (c *realmCache) nameHint(uuid string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.names[uuid]
	if !ok || time.Since(entry.at) > realmCacheTTL {
		return "", false
	}

	return entry.name, true
}

func (c *realmCache) isUnreadable(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	at, ok := c.unreadable[name]

	return ok && time.Since(at) <= realmCacheTTL
}

func (c *realmCache) currentGeneration() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.generation
}

// remember stores the name of a realm as a hint unless the cache was cleared since generation was read. A name
// belongs to one realm only, so any other id that was remembered under it is dropped.
func (c *realmCache) remember(generation uint64, uuid string, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if generation != c.generation || uuid == "" || name == "" {
		return
	}
	if c.names == nil {
		c.names = map[string]realmCacheEntry{}
	}

	for otherUuid, entry := range c.names {
		if entry.name == name && otherUuid != uuid {
			delete(c.names, otherUuid)
		}
	}
	delete(c.unreadable, name)
	c.names[uuid] = realmCacheEntry{name: name, at: time.Now()}
}

func (c *realmCache) markUnreadable(generation uint64, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if generation != c.generation || name == "" {
		return
	}
	if c.unreadable == nil {
		c.unreadable = map[string]time.Time{}
	}

	c.unreadable[name] = time.Now()
}

func (c *realmCache) forget() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++
	c.names, c.unreadable = nil, nil
}

// GetRealmNameByUUID returns the current name of the realm with the given internal id. Keycloak's admin API only
// addresses realms by name, so the realms are searched for the id. The names we expect it to have (knownName, the
// name last seen in state, and what this client remembers) are tried first, but only an id that Keycloak confirms
// counts: a name can be reused by another realm after a rename. If that fails the realms are listed.
func (keycloakClient *KeycloakClient) GetRealmNameByUUID(ctx context.Context, uuid string, knownName string) (string, error) {
	generation := keycloakClient.realmCache.currentGeneration()

	candidates := []string{}
	if knownName != "" {
		candidates = append(candidates, knownName)
	}
	if hint, ok := keycloakClient.realmCache.nameHint(uuid); ok && hint != knownName {
		candidates = append(candidates, hint)
	}

	forbidden := false
	for _, name := range candidates {
		realm, err := keycloakClient.GetRealm(ctx, name)
		switch {
		case err == nil && realm.Id == uuid:
			keycloakClient.realmCache.remember(generation, uuid, name)

			return name, nil
		case err == nil:
			// another realm has this name now, or Keycloak does not show the id: look further
		case ErrorIs403(err):
			forbidden = true
		case ErrorIs404(err):
			// renamed or deleted: look further
		default:
			return "", err
		}
	}

	for attempt := 0; ; attempt++ {
		found, idsHidden, err := keycloakClient.findRealmNameInList(ctx, generation, uuid)
		if err != nil {
			return "", err
		}
		if found != "" {
			return found, nil
		}

		// a token issued before the realm was created (or renamed) may not show it; try once with a new one
		if attempt == 0 {
			if err := keycloakClient.Refresh(ctx); err != nil {
				// a failed refresh says nothing about the realm: do not report it as missing
				return "", err
			}

			continue
		}

		// A realm we may not read is not a realm that does not exist: callers treat "not found" as "gone".
		if idsHidden || forbidden {
			return "", &ApiError{Code: http.StatusForbidden, Message: fmt.Sprintf("cannot look up the realm with internal id %s: Keycloak hides realm ids, or the realm itself, from accounts without the view-realm role, which is needed to use realm_uuid", uuid)}
		}

		return "", &ApiError{Code: http.StatusNotFound, Message: fmt.Sprintf("realm with internal id %s not found", uuid)}
	}
}

// findRealmNameInList lists the realms, remembers them, and returns the name of the one with the given
// internal id (empty when it is not there). idsHidden reports realms that Keycloak did not show an id for.
func (keycloakClient *KeycloakClient) findRealmNameInList(ctx context.Context, generation uint64, uuid string) (name string, idsHidden bool, err error) {
	realms, err := keycloakClient.GetRealms(ctx)
	if err != nil {
		return "", false, err
	}

	for _, realm := range realms {
		if realm.Id == "" {
			keycloakClient.realmCache.markUnreadable(generation, realm.Realm)
			idsHidden = true

			continue
		}

		keycloakClient.realmCache.remember(generation, realm.Id, realm.Realm)
		if realm.Id == uuid {
			name = realm.Realm
		}
	}

	return name, idsHidden, nil
}

// GetRealmUUID returns the internal id of the realm with the given name, or an empty string when Keycloak does not
// show it to this account. The id is always asked for again, because the name may belong to another realm than it
// did a moment ago; only the fact that an account cannot see it is remembered, for a short while.
func (keycloakClient *KeycloakClient) GetRealmUUID(ctx context.Context, name string) (string, error) {
	if keycloakClient.realmCache.isUnreadable(name) {
		return "", nil
	}

	generation := keycloakClient.realmCache.currentGeneration()

	realm, err := keycloakClient.GetRealm(ctx, name)
	if err != nil {
		if ErrorIs403(err) {
			keycloakClient.realmCache.markUnreadable(generation, name)
		}

		return "", err
	}

	if realm.Id == "" {
		keycloakClient.realmCache.markUnreadable(generation, name)

		return "", nil
	}

	keycloakClient.realmCache.remember(generation, realm.Id, name)

	return realm.Id, nil
}

// AuthRealm is the realm the provider authenticates against.
func (keycloakClient *KeycloakClient) AuthRealm() string {
	return keycloakClient.realm
}

func (keycloakClient *KeycloakClient) DeleteRealm(ctx context.Context, name string) error {
	keycloakClient.realmCache.forget()

	err := keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s", name), nil)
	if err != nil {
		// For whatever reason, this fails sometimes with a 500 during acceptance tests. try again
		return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s", name), nil)
	}

	return nil
}

func (keycloakClient *KeycloakClient) ValidateRealm(ctx context.Context, realm *Realm) error {
	if realm.DuplicateEmailsAllowed == true && realm.RegistrationEmailAsUsername == true {
		return fmt.Errorf("validation error: DuplicateEmailsAllowed cannot be true if RegistrationEmailAsUsername is true")
	}

	if realm.DuplicateEmailsAllowed == true && realm.LoginWithEmailAllowed == true {
		return fmt.Errorf("validation error: DuplicateEmailsAllowed cannot be true if LoginWithEmailAllowed is true")
	}

	if realm.SslRequired != "none" && realm.SslRequired != "external" && realm.SslRequired != "all" {
		return fmt.Errorf("validation error: SslRequired should be 'none', 'external' or 'all'")
	}

	// validate if the given theme exists on the server. the keycloak API allows you to use any random string for a theme
	serverInfo, err := keycloakClient.GetServerInfo(ctx)
	if err != nil {
		return err
	}

	if realm.LoginTheme != "" && !serverInfo.ThemeIsInstalled("login", realm.LoginTheme) {
		return fmt.Errorf("validation error: theme \"%s\" does not exist on the server", realm.LoginTheme)
	}

	if realm.AccountTheme != "" && !serverInfo.ThemeIsInstalled("account", realm.AccountTheme) {
		return fmt.Errorf("validation error: theme \"%s\" does not exist on the server", realm.AccountTheme)
	}

	if realm.AdminTheme != "" && !serverInfo.ThemeIsInstalled("admin", realm.AdminTheme) {
		return fmt.Errorf("validation error: theme \"%s\" does not exist on the server", realm.AdminTheme)
	}

	if realm.EmailTheme != "" && !serverInfo.ThemeIsInstalled("email", realm.EmailTheme) {
		return fmt.Errorf("validation error: theme \"%s\" does not exist on the server", realm.EmailTheme)
	}

	if realm.InternationalizationEnabled == true && !contains(realm.SupportLocales, realm.DefaultLocale) {
		return fmt.Errorf("validation error: DefaultLocale should be in the SupportLocales")
	}

	if realm.PasswordPolicy != "" {
		policies := strings.Split(realm.PasswordPolicy, " and ")
		for _, policyTypeRepresentation := range policies {
			policy := strings.Split(policyTypeRepresentation, "(")
			if !serverInfo.providerInstalled("password-policy", policy[0]) {
				return fmt.Errorf("validation error: password-policy \"%s\" does not exist on the server, installed providers: %s", policy[0], serverInfo.getInstalledProvidersNames("password-policy"))
			}
		}
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
