package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func genericProtocolMapperImport(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(data.Id(), "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid import. supported import formats: {{realmId}}/client/{{clientId}}/{{protocolMapperId}}, {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}")
	}

	parentResourceType := parts[1]
	parentResourceId := parts[2]

	data.Set("realm_id", parts[0])
	data.SetId(parts[3])

	if parentResourceType == "client" {
		data.Set("client_id", parentResourceId)
	} else if parentResourceType == "client-scope" {
		data.Set("client_scope_id", parentResourceId)
	} else {
		return nil, fmt.Errorf("the associated parent resource must be either a client or a client-scope")
	}

	return []*schema.ResourceData{data}, nil
}

// Temporary, duplicate helper function to support import of protocol mappers for which the schema
// supports the import flag. One of these functions can be removed once all protocol mappers support
// the import flag.
func genericProtocolMapperImportWithImportFlag(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(data.Id(), "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid import. supported import formats: {{realmId}}/client/{{clientId}}/{{protocolMapperId}}, {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}")
	}

	parentResourceType := parts[1]
	parentResourceId := parts[2]

	data.Set("realm_id", parts[0])
	data.SetId(parts[3])

	if parentResourceType == "client" {
		data.Set("client_id", parentResourceId)
	} else if parentResourceType == "client-scope" {
		data.Set("client_scope_id", parentResourceId)
	} else {
		return nil, fmt.Errorf("the associated parent resource must be either a client or a client-scope")
	}

	// Schema defaults are not applied automatically during terraform import, so we set
	// import = false explicitly to ensure ImportStateVerify passes.
	data.Set("import", false)

	return []*schema.ResourceData{data}, nil
}
