---
page_title: "keycloak_realm_keys Data Source"
---

# keycloak\_realm\_keys Data Source

Use this data source to get the keys of a realm. Keys can be filtered by algorithm and status.

Remarks:

- A key must meet all filter criteria
- This data source may return more than one value.
- If no key matches the filter criteria, then an error will be returned.

## Example Usage

```hcl
resource "keycloak_realm" "realm" {
    realm   = "my-realm"
    enabled = true
}

data "keycloak_realm_keys" "realm_keys" {
  realm_id   = keycloak_realm.realm.id
  algorithms = ["AES", "RS256"]
  status     = ["ACTIVE", "PASSIVE"]
}

# show certificate of first key:
output "certificate" {
  value = data.keycloak_realm_keys.realm_keys.keys[0].certificate
}

```

## Argument Reference

- `realm_id` - (Required) The realm from which the keys will be retrieved.
- `algorithms` - (Optional) When specified, keys will be filtered by algorithm. The algorithms can be any of `HS256`, `RS256`,`AES`, etc.
- `status` - (Optional) When specified, keys will be filtered by status. The statuses can be any of `ACTIVE`, `DISABLED` and `PASSIVE`.

## Attributes Reference

- `keys` - (Computed) A list of keys that match the filter criteria. Each key has the following attributes:
    - `algorithm` - Key algorithm (string)
    - `certificate` - Key certificate (string)
    - `provider_id` - Key provider ID (string)
    - `provider_priority` - Key provider priority (int64)
    - `kid` - Key ID (string)
    - `public_key` - Key public key (string)
    - `status` - Key status (string)
    - `type` - Key type (string)
