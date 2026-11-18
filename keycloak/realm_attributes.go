package keycloak

// MergeRealmAttributes returns the attributes to send when updating a realm.
//
// Keycloak removes every realm attribute that is missing from an update request whose attributes
// field is not null. Sending only the configured attributes would therefore delete those set by
// extensions, scripts or other resources. The result starts from the realm's current attributes,
// removes every key that is in oldAttributes but no longer in newAttributes, and then sets all keys
// from newAttributes. The given maps are not modified.
func MergeRealmAttributes(currentAttributes, oldAttributes, newAttributes map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{}, len(currentAttributes)+len(newAttributes))
	for key, value := range currentAttributes {
		merged[key] = value
	}

	for key := range oldAttributes {
		if _, ok := newAttributes[key]; !ok {
			delete(merged, key)
		}
	}

	for key, value := range newAttributes {
		merged[key] = value
	}

	return merged
}
