package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func parseGenericProtocolMapperImportId(importId string) (realmId string, clientId string, clientScopeId string, mapperReference string, lookupByName bool, err error) {
	parts := strings.Split(importId, "/")
	if len(parts) != 4 && len(parts) != 5 {
		return "", "", "", "", false, fmt.Errorf("invalid import. supported import formats: {{realmId}}/client/{{clientId}}/{{protocolMapperId}}, {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}, {{realmId}}/client/{{clientId}}/name/{{urlEncodedProtocolMapperName}}, {{realmId}}/client-scope/{{clientScopeId}}/name/{{urlEncodedProtocolMapperName}}")
	}

	realmId = parts[0]
	parentResourceType := parts[1]
	parentResourceId := parts[2]

	switch parentResourceType {
	case "client":
		clientId = parentResourceId
	case "client-scope":
		clientScopeId = parentResourceId
	default:
		return "", "", "", "", false, fmt.Errorf("the associated parent resource must be either a client or a client-scope")
	}

	if len(parts) == 4 {
		mapperReference = parts[3]
		return realmId, clientId, clientScopeId, mapperReference, false, nil
	}

	if parts[3] != "name" {
		return "", "", "", "", false, fmt.Errorf("invalid import format. expected 'name' segment for name-based import")
	}

	mapperName, decodeErr := url.PathUnescape(parts[4])
	if decodeErr != nil {
		return "", "", "", "", false, fmt.Errorf("unable to decode protocol mapper name %q: %w", parts[4], decodeErr)
	}

	if mapperName == "" {
		return "", "", "", "", false, fmt.Errorf("protocol mapper name cannot be empty")
	}

	return realmId, clientId, clientScopeId, mapperName, true, nil
}

func genericProtocolMapperImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId, clientId, clientScopeId, mapperReference, lookupByName, err := parseGenericProtocolMapperImportId(data.Id())
	if err != nil {
		return nil, err
	}

	data.Set("realm_id", realmId)

	if clientId != "" {
		data.Set("client_id", clientId)
	}

	if clientScopeId != "" {
		data.Set("client_scope_id", clientScopeId)
	}

	if lookupByName {
		mapper, err := keycloakClient.GetGenericProtocolMapperByName(ctx, realmId, clientId, clientScopeId, mapperReference)
		if err != nil {
			return nil, fmt.Errorf("failed to look up protocol mapper by name %q: %w", mapperReference, err)
		}
		if mapper == nil {
			return nil, fmt.Errorf("protocol mapper with name %q not found", mapperReference)
		}

		data.SetId(mapper.Id)
	} else {
		data.SetId(mapperReference)
	}

	return []*schema.ResourceData{data}, nil
}
