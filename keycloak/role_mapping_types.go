package keycloak

// RoleMapping struct for the MappingRepresentation
// https://www.keycloak.org/docs-api/latest/rest-api/index.html#MappingsRepresentation
type RoleMapping struct {
	ClientMappings map[string]*ClientRoleMapping `json:"clientMappings"`
	RealmMappings  []*Role                       `json:"realmMappings"`
}

// ClientRoleMapping struct for the ClientMappingRepresentation
// https://www.keycloak.org/docs-api/latest/rest-api/index.html#ClientMappingsRepresentation
type ClientRoleMapping struct {
	Client   string  `json:"client"`
	Id       string  `json:"id"`
	Mappings []*Role `json:"mappings"`
}
