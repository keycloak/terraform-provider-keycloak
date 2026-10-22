package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Resources that live inside a realm identify it with realm_id (a few use "realm"), which is the realm's name. A
// realm can be renamed, but its internal id never changes. withRealmUuid adds an optional
// realm_uuid attribute (typically keycloak_realm.x.internal_id) to such resources. Keycloak's admin
// API only addresses realms by name, so the realm_uuid is resolved to the current name before each
// create, read, update and delete. A resource that uses realm_uuid is therefore not replaced when
// its realm is renamed.
//
// The wrapper is applied to the whole resources map, so individual resources need no changes.
func withRealmUuid(resource *schema.Resource) {
	addRealmUuid(resource, true)
}

// withRealmUuidDataSource does the same for data sources. They are read at plan time, so after
// renaming a realm a data source that is still configured with the new name fails to find it
// until the rename has been applied; with realm_uuid it resolves the realm that exists now.
func withRealmUuidDataSource(dataSource *schema.Resource) {
	addRealmUuid(dataSource, false)
}

func addRealmUuid(resource *schema.Resource, writable bool) {
	realmField := ""
	for _, field := range []string{"realm_id", "realm"} {
		if s, ok := resource.Schema[field]; ok && s.Type == schema.TypeString && s.Required {
			realmField = field
			break
		}
	}
	if realmField == "" {
		return
	}

	realmId := resource.Schema[realmField]

	// the realm name becomes computed so it can be derived from realm_uuid; it is kept in sync on every read
	exactlyOneOf := []string{realmField, "realm_uuid"}
	realmId.Required = false
	realmId.Optional = true
	realmId.Computed = true
	// the stored realm_uuid would otherwise win over an edited realm name, and a resource cannot move to another realm anyway
	realmId.ForceNew = writable || realmId.ForceNew
	realmId.ExactlyOneOf = exactlyOneOf

	resource.Schema["realm_uuid"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ForceNew:     writable,
		ExactlyOneOf: exactlyOneOf,
		Description:  "The internal id of the realm, as an alternative to realm_id. Unlike the realm name, it does not change when the realm is renamed.",
	}

	if writable {
		resource.CustomizeDiff = withRealmUuidDiff(realmField, resource.CustomizeDiff)
	}

	if resource.CreateContext != nil {
		resource.CreateContext = withRealmReference(realmField, resource.CreateContext, false)
	}
	if resource.ReadContext != nil {
		resource.ReadContext = withRealmReference(realmField, resource.ReadContext, writable)
	}
	if resource.UpdateContext != nil {
		resource.UpdateContext = withRealmReference(realmField, resource.UpdateContext, false)
	}
	if resource.DeleteContext != nil {
		deleteContext := resource.DeleteContext
		resource.DeleteContext = func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
			if err := resolveRealmIdFromUuid(ctx, meta.(*keycloak.KeycloakClient), data, realmField); err != nil {
				if keycloak.ErrorIs404(err) {
					return nil // the realm is gone, and so is everything in it
				}
				return diag.FromErr(err)
			}

			return deleteContext(ctx, data, meta)
		}
	}
}

func withRealmReference(realmField string, next func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics, removeIfRealmGone bool) func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
		keycloakClient := meta.(*keycloak.KeycloakClient)
		previousName := data.Get(realmField).(string)

		if err := resolveRealmIdFromUuid(ctx, keycloakClient, data, realmField); err != nil {
			// a resource whose realm is gone is gone too; a data source must report it
			if removeIfRealmGone {
				return handleNotFoundError(ctx, err, data)
			}
			return diag.FromErr(err)
		}

		diags := next(ctx, data, meta)
		if diags.HasError() || data.Id() == "" {
			return diags
		}

		if err := setRealmUuid(ctx, keycloakClient, data, realmField); err != nil {
			return append(diags, diag.FromErr(err)...)
		}

		renameRealmInId(data, previousName, data.Get(realmField).(string))

		return diags
	}
}

// resolveRealmIdFromUuid points realm_id at the current name of the realm that realm_uuid refers to.
// Resources that only know realm_id are left alone.
func resolveRealmIdFromUuid(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData, realmField string) error {
	realmUuid := data.Get("realm_uuid").(string)
	if realmUuid == "" {
		return nil
	}

	storedName := data.Get(realmField).(string)
	realmName, err := keycloakClient.GetRealmNameByUUID(ctx, realmUuid, storedName)
	if err != nil {
		// the account may have lost view-realm since the uuid was stored; the stored name is the best we have
		if keycloak.ErrorIs403(err) && storedName != "" {
			return nil
		}

		return err
	}

	return data.Set(realmField, realmName)
}

// setRealmUuid fills in realm_uuid for resources that were configured with realm_id only. This is
// best effort: an account that may manage a realm's contents but not view the realm itself gets no
// realm id from Keycloak, and that must not break resources that never asked for realm_uuid.
func setRealmUuid(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData, realmField string) error {
	if data.Get("realm_uuid").(string) != "" {
		return nil
	}

	realmUuid, err := keycloakClient.GetRealmUUID(ctx, data.Get(realmField).(string))
	if err != nil || realmUuid == "" {
		return nil
	}

	return data.Set("realm_uuid", realmUuid)
}

// realmNameOrIdPrefix is for the few resources whose id starts with the realm name. The id keeps the
// name the realm had when the resource was created, so prefer realm_id, which follows a rename.
func realmNameOrIdPrefix(data *schema.ResourceData, idPrefix string) string {
	if realmId := data.Get("realm_id").(string); realmId != "" {
		return realmId
	}

	return idPrefix
}

// withRealmUuidDiff guards the move from realm_id to realm_uuid on an existing resource. When the
// state has no realm_uuid (Keycloak hides realm ids from accounts without the view-realm role), the
// configured uuid looks like a different realm and would replace the resource. If it resolves to
// the realm the resource is already in, nothing changes; if it cannot be resolved, fail instead of
// replacing something that may hold data.
func withRealmUuidDiff(realmField string, next schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
		if keycloakClient, ok := meta.(*keycloak.KeycloakClient); ok && keycloakClient != nil && diff.Id() != "" && diff.HasChange("realm_uuid") {
			oldUuid, newUuid := diff.GetChange("realm_uuid")
			if oldUuid.(string) == "" && newUuid.(string) != "" && diff.NewValueKnown("realm_uuid") {
				realmName, err := keycloakClient.GetRealmNameByUUID(ctx, newUuid.(string), diff.Get(realmField).(string))
				if err != nil {
					return err
				}

				if realmName == diff.Get(realmField).(string) {
					if err := diff.Clear("realm_uuid"); err != nil {
						return err
					}
				}
			}
		}

		if next != nil {
			return next(ctx, diff, meta)
		}

		return nil
	}
}

// renameRealmInId keeps an id like "<realm>/<something>" in line with the realm after it was renamed.
func renameRealmInId(data *schema.ResourceData, previousName string, currentName string) {
	if previousName == "" || currentName == "" || previousName == currentName {
		return
	}

	if id := data.Id(); strings.HasPrefix(id, previousName+"/") {
		data.SetId(currentName + strings.TrimPrefix(id, previousName))
	}
}
