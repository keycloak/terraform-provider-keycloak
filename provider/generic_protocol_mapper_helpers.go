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
