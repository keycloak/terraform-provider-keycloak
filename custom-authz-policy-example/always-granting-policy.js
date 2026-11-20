// An always-granting JavaScript authorization policy used by the terraform-provider-keycloak
// acceptance tests. Once deployed (in a JAR with META-INF/keycloak-scripts.json) it is
// referenced as a policy of type "script-always-granting-policy.js".
$evaluation.grant();
