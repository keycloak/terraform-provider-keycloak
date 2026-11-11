package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const (
	clientScopeOwnerRealm  = "realm"
	clientScopeOwnerClient = "client"
)

type clientScopeAssignments struct {
	clientProtocol   string
	clientBearerOnly bool
	defaultScopes    []string
	optionalScopes   []string
	scopeIDs         map[string]string
	scopeProtocols   map[string]string
}

func resourceKeycloakClientScopeAssignments() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakClientScopeAssignmentsReconcile,
		ReadContext:   resourceKeycloakClientScopeAssignmentsRead,
		UpdateContext: resourceKeycloakClientScopeAssignmentsReconcile,
		DeleteContext: resourceKeycloakClientScopeAssignmentsDelete,
		CustomizeDiff: customizeDiffKeycloakClientScopeAssignments,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakClientScopeAssignmentsImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"default_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
			"optional_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
		},
	}
}

func clientScopeAssignmentsId(realmId, clientId string) string {
	if clientId == "" {
		return fmt.Sprintf("%s/%s", clientScopeOwnerRealm, realmId)
	}
	return fmt.Sprintf("%s/%s/%s", clientScopeOwnerClient, realmId, clientId)
}

func scopeNamesFromSet(value interface{}) []string {
	scopeSet, ok := value.(*schema.Set)
	if !ok || scopeSet == nil {
		return nil
	}

	names := make([]string, 0, scopeSet.Len())
	for _, value := range scopeSet.List() {
		if name, ok := value.(string); ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	return names
}

func validateNoDuplicateClientScopeAssignments(defaultScopes, optionalScopes []string) error {
	defaultScopeSet := make(map[string]struct{}, len(defaultScopes))
	for _, scopeName := range defaultScopes {
		defaultScopeSet[scopeName] = struct{}{}
	}

	for _, scopeName := range optionalScopes {
		if _, exists := defaultScopeSet[scopeName]; exists {
			return fmt.Errorf("validation error: client scope %q cannot be assigned as both default and optional", scopeName)
		}
	}

	return nil
}

func customizeDiffKeycloakClientScopeAssignments(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	if !diff.NewValueKnown("default_scopes") || !diff.NewValueKnown("optional_scopes") {
		return nil
	}
	return validateNoDuplicateClientScopeAssignments(
		scopeNamesFromSet(diff.Get("default_scopes")),
		scopeNamesFromSet(diff.Get("optional_scopes")),
	)
}

func readRealmClientScopeAssignments(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId string) (*clientScopeAssignments, error) {
	defaultScopes, err := keycloakClient.GetRealmDefaultClientScopes(ctx, realmId)
	if err != nil {
		return nil, err
	}
	optionalScopes, err := keycloakClient.GetRealmOptionalClientScopes(ctx, realmId)
	if err != nil {
		return nil, err
	}

	assignments := &clientScopeAssignments{
		scopeIDs:       make(map[string]string, len(defaultScopes)+len(optionalScopes)),
		scopeProtocols: make(map[string]string, len(defaultScopes)+len(optionalScopes)),
		defaultScopes:  make([]string, 0, len(defaultScopes)),
		optionalScopes: make([]string, 0, len(optionalScopes)),
	}
	for _, scope := range defaultScopes {
		assignments.defaultScopes = append(assignments.defaultScopes, scope.Name)
		assignments.scopeIDs[scope.Name] = scope.Id
		assignments.scopeProtocols[scope.Name] = scope.Protocol
	}
	for _, scope := range optionalScopes {
		assignments.optionalScopes = append(assignments.optionalScopes, scope.Name)
		assignments.scopeIDs[scope.Name] = scope.Id
		assignments.scopeProtocols[scope.Name] = scope.Protocol
	}
	sort.Strings(assignments.defaultScopes)
	sort.Strings(assignments.optionalScopes)

	return assignments, nil
}

func readClientClientScopeAssignments(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId, clientId string) (*clientScopeAssignments, error) {
	client, err := keycloakClient.GetGenericClient(ctx, realmId, clientId)
	if err != nil {
		return nil, err
	}

	assignments := &clientScopeAssignments{
		clientProtocol:   client.Protocol,
		clientBearerOnly: client.BearerOnly,
		scopeIDs:         make(map[string]string),
		scopeProtocols:   make(map[string]string),
	}
	switch client.Protocol {
	case keycloakClientScopeProtocolOpenidConnect:
		defaultScopes, err := keycloakClient.GetOpenidClientDefaultScopes(ctx, realmId, clientId)
		if err != nil {
			return nil, err
		}
		optionalScopes, err := keycloakClient.GetOpenidClientOptionalScopes(ctx, realmId, clientId)
		if err != nil {
			return nil, err
		}
		for _, scope := range defaultScopes {
			assignments.defaultScopes = append(assignments.defaultScopes, scope.Name)
			assignments.scopeIDs[scope.Name] = scope.Id
			assignments.scopeProtocols[scope.Name] = scope.Protocol
		}
		for _, scope := range optionalScopes {
			assignments.optionalScopes = append(assignments.optionalScopes, scope.Name)
			assignments.scopeIDs[scope.Name] = scope.Id
			assignments.scopeProtocols[scope.Name] = scope.Protocol
		}
	case keycloakClientScopeProtocolSaml:
		defaultScopes, err := keycloakClient.GetSamlClientDefaultScopes(ctx, realmId, clientId)
		if err != nil {
			return nil, err
		}
		optionalScopes, err := keycloakClient.GetSamlClientOptionalScopes(ctx, realmId, clientId)
		if err != nil {
			return nil, err
		}
		for _, scope := range defaultScopes {
			assignments.defaultScopes = append(assignments.defaultScopes, scope.Name)
			assignments.scopeIDs[scope.Name] = scope.Id
			assignments.scopeProtocols[scope.Name] = scope.Protocol
		}
		for _, scope := range optionalScopes {
			assignments.optionalScopes = append(assignments.optionalScopes, scope.Name)
			assignments.scopeIDs[scope.Name] = scope.Id
			assignments.scopeProtocols[scope.Name] = scope.Protocol
		}
	default:
		return nil, fmt.Errorf("validation error: client with id %s uses unsupported protocol %q", clientId, client.Protocol)
	}
	sort.Strings(assignments.defaultScopes)
	sort.Strings(assignments.optionalScopes)

	return assignments, nil
}

func readClientScopeAssignments(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId, clientId string) (*clientScopeAssignments, error) {
	if clientId == "" {
		return readRealmClientScopeAssignments(ctx, keycloakClient, realmId)
	}
	return readClientClientScopeAssignments(ctx, keycloakClient, realmId, clientId)
}

func resourceKeycloakClientScopeAssignmentsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	assignments, err := readClientScopeAssignments(ctx, keycloakClient, realmId, clientId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}
	if err := data.Set("default_scopes", assignments.defaultScopes); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("optional_scopes", assignments.optionalScopes); err != nil {
		return diag.FromErr(err)
	}
	data.SetId(clientScopeAssignmentsId(realmId, clientId))

	return nil
}

func (assignments *clientScopeAssignments) allScopes() []string {
	return append(append([]string{}, assignments.defaultScopes...), assignments.optionalScopes...)
}

func resolveClientScopeAssignmentIDs(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId, clientId string, current, desired *clientScopeAssignments) (map[string]string, error) {
	scopeIDs := make(map[string]string, len(current.scopeIDs))
	scopeProtocols := make(map[string]string, len(current.scopeProtocols))
	for name, id := range current.scopeIDs {
		scopeIDs[name] = id
	}
	for name, protocol := range current.scopeProtocols {
		scopeProtocols[name] = protocol
	}

	needsList := false
	for _, name := range current.allScopes() {
		if scopeIDs[name] == "" {
			needsList = true
			break
		}
	}
	for _, name := range desired.allScopes() {
		if scopeIDs[name] == "" || (clientId != "" && scopeProtocols[name] == "") {
			needsList = true
			break
		}
	}

	if needsList {
		scopes, err := keycloakClient.ListClientScopes(ctx, realmId)
		if err != nil {
			return nil, err
		}
		for _, scope := range scopes {
			scopeIDs[scope.Name] = scope.Id
			scopeProtocols[scope.Name] = scope.Protocol
		}
	}

	for _, name := range current.allScopes() {
		if scopeIDs[name] == "" {
			return nil, fmt.Errorf("validation error: currently assigned client scope %q no longer exists in realm %q", name, realmId)
		}
	}
	for _, name := range desired.allScopes() {
		if scopeIDs[name] == "" {
			return nil, fmt.Errorf("validation error: client scope %q does not exist in realm %q", name, realmId)
		}
		if clientId != "" && scopeProtocols[name] != current.clientProtocol {
			return nil, fmt.Errorf("validation error: client scope %q uses protocol %q, but client %q uses protocol %q", name, scopeProtocols[name], clientId, current.clientProtocol)
		}
	}

	return scopeIDs, nil
}

func applyClientScopeAssignmentChanges(ctx context.Context, keycloakClient *keycloak.KeycloakClient, realmId, clientId, assignmentType string, scopeNames []string, scopeIDs map[string]string, add bool) error {
	for _, name := range scopeNames {
		scopeID := scopeIDs[name]
		if scopeID == "" {
			return fmt.Errorf("validation error: no ID found for client scope %q", name)
		}
		var err error
		if add {
			err = keycloakClient.AddClientScopeAssignment(ctx, realmId, clientId, assignmentType, scopeID)
		} else {
			err = keycloakClient.RemoveClientScopeAssignment(ctx, realmId, clientId, assignmentType, scopeID)
			if keycloak.ErrorIs404(err) {
				err = nil
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func scopeNameSet(names []string) map[string]struct{} {
	result := make(map[string]struct{}, len(names))
	for _, name := range names {
		result[name] = struct{}{}
	}
	return result
}

type clientScopeAssignmentChanges struct {
	removeDefaults  []string
	removeOptionals []string
	addDefaults     []string
	addOptionals    []string
}

func planClientScopeAssignmentChanges(current, desired *clientScopeAssignments) clientScopeAssignmentChanges {
	desiredDefaults := scopeNameSet(desired.defaultScopes)
	desiredOptionals := scopeNameSet(desired.optionalScopes)
	currentDefaults := scopeNameSet(current.defaultScopes)
	currentOptionals := scopeNameSet(current.optionalScopes)
	changes := clientScopeAssignmentChanges{}
	removed := map[string]struct{}{}

	for name := range currentDefaults {
		if _, remainsDefault := desiredDefaults[name]; !remainsDefault {
			changes.removeDefaults = append(changes.removeDefaults, name)
			removed[name] = struct{}{}
		}
	}
	for name := range currentOptionals {
		if _, remainsOptional := desiredOptionals[name]; !remainsOptional {
			changes.removeOptionals = append(changes.removeOptionals, name)
			removed[name] = struct{}{}
		}
	}

	// Client-scope DELETE is type-agnostic, so reattach a desired scope even if
	// it was already present in that category when another category removed it.
	for name := range desiredDefaults {
		_, alreadyDefault := currentDefaults[name]
		_, wasRemoved := removed[name]
		if !alreadyDefault || wasRemoved {
			changes.addDefaults = append(changes.addDefaults, name)
		}
	}
	for name := range desiredOptionals {
		_, alreadyOptional := currentOptionals[name]
		_, wasRemoved := removed[name]
		if !alreadyOptional || wasRemoved {
			changes.addOptionals = append(changes.addOptionals, name)
		}
	}
	sort.Strings(changes.removeDefaults)
	sort.Strings(changes.removeOptionals)
	sort.Strings(changes.addDefaults)
	sort.Strings(changes.addOptionals)
	return changes
}

func resourceKeycloakClientScopeAssignmentsReconcile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	desired := &clientScopeAssignments{
		defaultScopes:  scopeNamesFromSet(data.Get("default_scopes")),
		optionalScopes: scopeNamesFromSet(data.Get("optional_scopes")),
	}
	if err := validateNoDuplicateClientScopeAssignments(desired.defaultScopes, desired.optionalScopes); err != nil {
		return diag.FromErr(err)
	}

	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	current, err := readClientScopeAssignments(ctx, keycloakClient, realmId, clientId)
	if err != nil {
		if clientId != "" && keycloak.ErrorIs404(err) {
			return diag.Errorf("validation error: client with id %s does not exist", clientId)
		}
		return diag.FromErr(err)
	}

	changes := planClientScopeAssignmentChanges(current, desired)
	scopeIDs, err := resolveClientScopeAssignmentIDs(ctx, keycloakClient, realmId, clientId, current, desired)
	if err != nil {
		return diag.FromErr(err)
	}
	if current.clientProtocol == keycloakClientScopeProtocolOpenidConnect && current.clientBearerOnly && (len(changes.addDefaults) > 0 || len(changes.addOptionals) > 0) {
		return diag.Errorf("validation error: client with id %s uses access type BEARER-ONLY which does not use scopes", clientId)
	}

	// Retain the target in state if a later write fails after partial mutation.
	data.SetId(clientScopeAssignmentsId(realmId, clientId))
	if err := applyClientScopeAssignmentChanges(ctx, keycloakClient, realmId, clientId, "default", changes.removeDefaults, scopeIDs, false); err != nil {
		return diag.FromErr(err)
	}
	if err := applyClientScopeAssignmentChanges(ctx, keycloakClient, realmId, clientId, "optional", changes.removeOptionals, scopeIDs, false); err != nil {
		return diag.FromErr(err)
	}
	if err := applyClientScopeAssignmentChanges(ctx, keycloakClient, realmId, clientId, "default", changes.addDefaults, scopeIDs, true); err != nil {
		return diag.FromErr(err)
	}
	if err := applyClientScopeAssignmentChanges(ctx, keycloakClient, realmId, clientId, "optional", changes.addOptionals, scopeIDs, true); err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakClientScopeAssignmentsRead(ctx, data, meta)
}

func resourceKeycloakClientScopeAssignmentsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)

	current, err := readClientScopeAssignments(ctx, keycloakClient, realmId, clientId)
	if err != nil {
		if keycloak.ErrorIs404(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	// Do not delete stale assignments that have since been removed or moved.
	defaultScopes := data.Get("default_scopes").(*schema.Set).Intersection(schema.NewSet(schema.HashString, stringsToInterfaces(current.defaultScopes)))
	optionalScopes := data.Get("optional_scopes").(*schema.Set).Intersection(schema.NewSet(schema.HashString, stringsToInterfaces(current.optionalScopes)))
	if err := applyClientScopeAssignmentChanges(ctx, keycloakClient, realmId, clientId, "default", scopeNamesFromSet(defaultScopes), current.scopeIDs, false); err != nil {
		return diag.FromErr(err)
	}
	if err := applyClientScopeAssignmentChanges(ctx, keycloakClient, realmId, clientId, "optional", scopeNamesFromSet(optionalScopes), current.scopeIDs, false); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakClientScopeAssignmentsImport(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(data.Id(), "/")
	switch {
	case len(parts) == 2 && parts[0] == clientScopeOwnerRealm && parts[1] != "":
		if err := data.Set("realm_id", parts[1]); err != nil {
			return nil, err
		}
	case len(parts) == 3 && parts[0] == clientScopeOwnerClient && parts[1] != "" && parts[2] != "":
		if err := data.Set("realm_id", parts[1]); err != nil {
			return nil, err
		}
		if err := data.Set("client_id", parts[2]); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid import ID %q; expected realm/<realm_id> or client/<realm_id>/<client_id>", data.Id())
	}

	return []*schema.ResourceData{data}, nil
}
