## What's Changed

## 5.5.0 (October 13, 2025)

FEATURES:

* feat: add admin url variable to separate authn and admin api urls (forked and added reviewer suggestions) by @au-ee in https://github.com/keycloak/terraform-provider-keycloak/pull/1305
* provider/github: init OIDC provider for GitHub by @RaitoBezarius in https://github.com/keycloak/terraform-provider-keycloak/pull/1281
* Add support for mtls client cert auth by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1317
* feat: Add support for authentication with provided Keycloak Access Token (#1319) by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1320
* Add OAUTH authentication option to SMTP servers by @lucymcphail in https://github.com/keycloak/terraform-provider-keycloak/pull/1293
* [issue-1328] Add support for sub mapper by @ethan-k in https://github.com/keycloak/terraform-provider-keycloak/pull/1323
* Update to KC 26.3.3 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1292
* Update to KC 26.3.4 by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1316
* Update KC 26.3.5 and KC 26.4.0 by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1340

IMPROVEMENTS:

* Document client roles in default roles by @ingokuba in https://github.com/keycloak/terraform-provider-keycloak/pull/1295
* docs: add keycloak built-in required actions docs by @Download in https://github.com/keycloak/terraform-provider-keycloak/pull/1308
* Update gradle version to 9.0.0 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1291
* feat: update go toolchain to 1.24.4 by @LaughDonor in https://github.com/keycloak/terraform-provider-keycloak/pull/1307
* chore: added documentation of standard_token_exchange_enabled and allow_refresh_token_in_standard_token_exchange by @eicki in https://github.com/keycloak/terraform-provider-keycloak/pull/1296
* chore: Update terraform version in CI to 1.13.3 by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1321
* Update golang to 1.24, golang.org/x/net to v.0.44.0 and terraform-plu… by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1322
* chore(deps): bump github.com/golang-jwt/jwt/v5 from 5.2.3 to 5.3.0 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1269
* chore(deps): bump golang.org/x/net from 0.42.0 to 0.43.0 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1271
* chore(deps): bump goreleaser/goreleaser-action from 6.3.0 to 6.4.0 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1275
* chore(deps): bump actions/checkout from 4 to 5 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1276
* chore(deps): bump org.jetbrains.kotlin.jvm from 2.2.0 to 2.2.10 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1278
* chore(deps): bump gradle/actions from 4.4.1 to 4.4.2 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1272
* chore(deps): bump com.gradleup.shadow from 8.3.8 to 9.1.0 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1288
* chore(deps): bump actions/github-script from 7 to 8 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1303
* chore(deps): bump actions/setup-go from 5 to 6 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1302
* chore(deps): bump org.jetbrains.kotlin.jvm from 2.2.10 to 2.2.20 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1313
* chore(deps): bump gradle/actions from 4.4.2 to 4.4.3 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1311
* chore(deps): bump com.gradleup.shadow from 9.1.0 to 9.2.2 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1330
* chore(deps): bump gradle/actions from 4.4.3 to 5.0.0 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1336

BUG FIXES:

* fix: handles arrays and maps in client policy conditions (#1256) by @remcojansen in https://github.com/keycloak/terraform-provider-keycloak/pull/1338
* fix: Add missing standard_flow_enabled to OIDC client examples (#1325) by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1326
* Fix path for realm_localization documentation by @ingokuba in https://github.com/keycloak/terraform-provider-keycloak/pull/1314

Huge thanks to all the individuals who have contributed towards this release:

- [au-ee](https://github.com/au-ee)
- [RaitoBezarius](https://github.com/RaitoBezarius)
- [lucymcphail](https://github.com/lucymcphail)
- [ethan-k](https://github.com/ethan-k)
- [ingokuba](https://github.com/ingokuba)
- [Download](https://github.com/Download)
- [LaughDonor](https://github.com/LaughDonor)
- [eicki](https://github.com/eicki)
- [KyriosGN0](https://github.com/KyriosGN0)
- [remcojansen](https://github.com/remcojansen)

## 5.4.0 (July 28, 2025)

FEATURES:

* Update to KC26.3.2 (#1259) by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1260
* Add support for fetchRoles in keycloak_openid_client_role_policy resource by @elshirak in https://github.com/keycloak/terraform-provider-keycloak/pull/1241
* Support signed jwt client auth type by @tuankiettruong in https://github.com/keycloak/terraform-provider-keycloak/pull/1250

IMPROVEMENTS:

* Rename realm_client_policy_profile_profile.md by @mawenge in https://github.com/keycloak/terraform-provider-keycloak/pull/1249
* docs: Add mention of keycloak_authentication_bindings to realm authentication settings by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1255
* docs: add example to keycloak_custom_identity_provider_mapper docs about how use the group mapper  by @KyriosGN0 in https://github.com/keycloak/terraform-provider-keycloak/pull/1251
* Add import instructions & example for client_scope by @ajoga in https://github.com/keycloak/terraform-provider-keycloak/pull/1248
* Add RSA_SHAxxx_MGF1 signature algorithms to SAML identity provider config by @yrd in https://github.com/keycloak/terraform-provider-keycloak/pull/1199
* chore(deps): bump golang.org/x/net from 0.41.0 to 0.42.0 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1247
* chore(deps): bump github.com/golang-jwt/jwt/v5 from 5.2.2 to 5.2.3 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1261

BUG FIXES:

* Fix/1246 group names with slash dont work by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1252

Huge thanks to all the individuals who have contributed towards this release:

- [yrd](https://github.com/yrd)
- [ajoga](https://github.com/ajoga)
- [KyriosGN0](https://github.com/KyriosGN0)
- [elshirak](https://github.com/elshirak)
- [tuankiettruong](https://github.com/tuankiettruong)
- [mawenge](https://github.com/mawenge)

## 5.3.0 (July 11, 2025)

FEATURES:

* Add support for ephemeral/write-only arguments for keycloak_oidc_identity_provider & keycloak_openid_client by @refucktor in https://github.com/keycloak/terraform-provider-keycloak/pull/1190
* Update to KC26.2.5 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1214
* Add support for Standard Token Exchange configuration in OpenID Clients by @sventorben in https://github.com/keycloak/terraform-provider-keycloak/pull/1189
* Feat: support organization feature by @tuankiettruong in https://github.com/keycloak/terraform-provider-keycloak/pull/1209
* feat: Enable importing user by username by @viniciusd in https://github.com/keycloak/terraform-provider-keycloak/pull/1113
* expose EnableStandardTokenExchange  and allow refresh token in standard token exchange in keycloak_openid_client  by @KyriosGN0 in https://github.com/keycloak/terraform-provider-keycloak/pull/1229
* feat: Add new resources for realm client policy profiles by @laureat-natzka in https://github.com/keycloak/terraform-provider-keycloak/pull/1138
* Openid client secret rotation by @laureat-natzka in https://github.com/keycloak/terraform-provider-keycloak/pull/1234
* 1235 update for KC26.3.0 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1236
* feat: support role IDs in default role resource by @horus in https://github.com/keycloak/terraform-provider-keycloak/pull/764

IMPROVEMENTS:

* feat(protocol-mapper): add "add to token introspection" flag for realm role protocol mapper by @maxjoehnk in https://github.com/keycloak/terraform-provider-keycloak/pull/947
* Add support for disableTypeClaimCheck by @ybasket in https://github.com/keycloak/terraform-provider-keycloak/pull/1181
* optimize resolving `ParentId` property with group-by-path API by @kherock in https://github.com/keycloak/terraform-provider-keycloak/pull/1225
* Use the ID from the user imported via username by @viniciusd in https://github.com/keycloak/terraform-provider-keycloak/pull/1228
* 1242 update to KC26.3.1 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1243
* chore(deps): bump dario.cat/mergo from 1.0.1 to 1.0.2 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1196
* chore(deps): bump golang.org/x/net from 0.39.0 to 0.40.0 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1197
* chore(deps): bump org.jetbrains.kotlin.jvm from 2.1.20 to 2.1.21 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1203
* chore(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.36.1 to 2.37.0 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1202
* chore(deps): bump gradle/actions from 4.3.1 to 4.4.0 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1204
* chore(deps): bump golang.org/x/net from 0.40.0 to 0.41.0 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1216
* chore(deps): bump github.com/cloudflare/circl from 1.6.0 to 1.6.1 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1219
* chore(deps): bump gradle/actions from 4.4.0 to 4.4.1 in /.github/workflows by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1220
* chore(deps): bump org.jetbrains.kotlin.jvm from 2.1.21 to 2.2.0 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1231
* chore(deps): bump com.gradleup.shadow from 8.3.6 to 8.3.7 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1230
* chore(deps): bump github.com/hashicorp/go-retryablehttp from 0.7.7 to 0.7.8 by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1224
* chore(deps): bump com.gradleup.shadow from 8.3.7 to 8.3.8 in /custom-user-federation-example by @dependabot[bot] in https://github.com/keycloak/terraform-provider-keycloak/pull/1237

BUG FIXES:

* Docs: Updated Markdown to format correctly by @bubbletroubles in https://github.com/keycloak/terraform-provider-keycloak/pull/1201
* Removed feature override so KC_FEATURES environment variable is used by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1232

Huge thanks to all the individuals who have contributed towards this release:

- [refucktor](https://github.com/refucktor)
- [bubbletroubles](https://github.com/bubbletroubles)
- [sventorben](https://github.com/sventorben)
- [maxjoehnk](https://github.com/maxjoehnk)
- [ybasket](https://github.com/ybasket)
- [kherock](https://github.com/kherock)
- [tuankiettruong](https://github.com/tuankiettruong)
- [viniciusd](https://github.com/viniciusd)
- [KyriosGN0](https://github.com/KyriosGN0)
- [laureat-natzka](https://github.com/laureat-natzka)
- [horus](https://github.com/horus)

## 5.2.0 (April 14, 2025)

FEATURES:

* Feature/423 add hardcoded attribute mapper by @angeloxx in https://github.com/keycloak/terraform-provider-keycloak/pull/950
* feat: Add option to configure always_display_in_console for clients by @mchittka in https://github.com/keycloak/terraform-provider-keycloak/pull/1114
* Feature/keycloak_required_action config values by @laureat-natzka in https://github.com/keycloak/terraform-provider-keycloak/pull/996
* feat: add all supported alogithms for rsa-enc keystore by @tboerger in https://github.com/keycloak/terraform-provider-keycloak/pull/1092
* feat: add consent required flag on saml clients by @frpicard in https://github.com/keycloak/terraform-provider-keycloak/pull/1130
* 1093 added support for KC26.1.0 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1094
* feat: add openid_client_authorization_client_scope_policy resource by @fitbeard in https://github.com/keycloak/terraform-provider-keycloak/pull/1128
* feat: add realm translation resource by @jonathandavies-CS in https://github.com/keycloak/terraform-provider-keycloak/pull/1123
* 1149 update to KC26.1.4 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1150
* Add tests for saml aggregate attributes by @robson90 in https://github.com/keycloak/terraform-provider-keycloak/pull/1171
* Allow the alias of Google IdP to be set by @irby in https://github.com/keycloak/terraform-provider-keycloak/pull/1177
* feat: add identity provider hardcoded group mapper by @gkfabs in https://github.com/keycloak/terraform-provider-keycloak/pull/886
* feat: Support extra_origins in web_authn_policy and web_authn_passwordless_policy by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1173

IMPROVEMENTS:

* 1100 run ci on release branches by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1101
* Adds support for the ExcludeIssuerFromAuthResponse option on OpenIdClient by @sebght in https://github.com/keycloak/terraform-provider-keycloak/pull/934
* chore(deps): bump golang.org/x/net from 0.33.0 to 0.36.0 by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1145
* Fixed SHA commit values for non-GH GH actions by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1151
* Create dependabot.yml by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1152
* chore(deps): bump com.gradleup.shadow from 8.3.0 to 8.3.6 in /custom-user-federation-example by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1157
* Feature/update go 1.23 by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1158
* chore(deps): bump org.jetbrains.kotlin.jvm from 2.0.20 to 2.1.10 in /custom-user-federation-example by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1156
* chore(deps): bump org.jetbrains.kotlin.jvm from 2.1.10 to 2.1.20 in /custom-user-federation-example by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1160
* fix: reuse TCP Connection by @Filirom1 in https://github.com/keycloak/terraform-provider-keycloak/pull/1163
* Retry on timeout, when server doesn't respond by @Filirom1 in https://github.com/keycloak/terraform-provider-keycloak/pull/1161
* chore(deps): bump crazy-max/ghaction-import-gpg from 6.2.0 to 6.3.0 in /.github/workflows by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1168
* chore(deps): bump gradle/actions from 4.3.0 to 4.3.1 in /.github/workflows by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1167
* chore(deps): bump golang.org/x/net from 0.37.0 to 0.38.0 by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1165
* chore(deps): bump goreleaser/goreleaser-action from 6.2.1 to 6.3.0 in /.github/workflows by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1166
* Increase MaxIdleConnsPerHost in http.Transport by @Filirom1 in https://github.com/keycloak/terraform-provider-keycloak/pull/1169
* chore(deps): bump golang.org/x/net from 0.38.0 to 0.39.0 by @dependabot in https://github.com/keycloak/terraform-provider-keycloak/pull/1183

BUG FIXES:

* 1091 fixed provider crash with initial_login set to false by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1097
* 1099 fix multi-valued user attributes cannot be enabled by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1106
* doc: fix token exchange doc urls by @emouty in https://github.com/keycloak/terraform-provider-keycloak/pull/1105
* fix: fix the wrong error return value by @drawdrop in https://github.com/keycloak/terraform-provider-keycloak/pull/1134
* fix: Render json `null` to indicate an empty value of type KeycloakSliceHashDelimited (#1142). by @thomasdarimont in https://github.com/keycloak/terraform-provider-keycloak/pull/1144
* 1146 removed vulnerable github action tj-actions/changed-files by @sschu in https://github.com/keycloak/terraform-provider-keycloak/pull/1147
* fix: Typo in realm_keys.md by @simonregn in https://github.com/keycloak/terraform-provider-keycloak/pull/1174

Huge thanks to all the individuals who have contributed towards this release:

- [@angeloxx](https://github.com/angeloxx)
- [@mchittka](https://github.com/mchittka)
- [@laureat-natzka](https://github.com/laureat-natzka)
- [@tboerger](https://github.com/tboerger)
- [@frpicard](https://github.com/frpicard)
- [@fitbeard](https://github.com/fitbeard)
- [@jonathandavies-CS](https://github.com/jonathandavies-CS)
- [@robson90](https://github.com/robson90)
- [@irby](https://github.com/irby)
- [@gkfabs](https://github.com/gkfabs)
- [@Filirom1](https://github.com/Filirom1)
- [@emouty](https://github.com/emouty)
- [@drawdrop](https://github.com/drawdrop)
- [@simonregn](https://github.com/simonregn)

## 5.1.0 (January 24, 2025)

FEATURES:

* feat: allow configuration of display_name for oidc identity providers ([#1033](https://github.com/keycloak/terraform-provider-keycloak/pull/1033))
* Configure default and optional client scopes per realm via dedicated resources ([#1079](https://github.com/keycloak/terraform-provider-keycloak/pull/1079))

IMPROVEMENTS:

* Ignore changes on markdown files in pipeline triggers ([#1070](https://github.com/keycloak/terraform-provider-keycloak/pull/1070))
* Updated KC26 version to KC26.0.8 ([#1075](https://github.com/keycloak/terraform-provider-keycloak/pull/1075))
* Add aggregate attributes flag to keycloak_saml_user_attribute_protocol_mapper ([#1080](https://github.com/keycloak/terraform-provider-keycloak/pull/1080))
* 1073 add organization setting to realm ([#1076](https://github.com/keycloak/terraform-provider-keycloak/pull/1076))
* refactor: speed up audience mapper validation ([#961](https://github.com/keycloak/terraform-provider-keycloak/pull/961))
* 1067 Add first broker login flow to authentication flow bindings ([#1084](https://github.com/keycloak/terraform-provider-keycloak/pull/1084))

BUG FIXES:

* chore: Remove temporary note from readme.md ([#1068](https://github.com/keycloak/terraform-provider-keycloak/pull/1068))
* fix: allow definition of multivalued user profile attributes ([#1071](https://github.com/keycloak/terraform-provider-keycloak/pull/1071))
* chore: fixed deprecation warning in release pipeline ([#1072](https://github.com/keycloak/terraform-provider-keycloak/pull/1072))

Huge thanks to all the individuals who have contributed towards this release:

- [@bnjns](https://github.com/bnjns)
- [@pboehm](https://github.com/pboehm)
- [@robson90](https://github.com/robson90)
- [@sybereal](https://github.com/sybereal)

## 5.0.0 (January 13, 2025)

FEATURES:

- Fix user & userprofile tests ([#1042](https://github.com/keycloak/terraform-provider-keycloak/pull/1042))
- add support for keycloak version up to current (26.0.7) ([#1028](https://github.com/keycloak/terraform-provider-keycloak/pull/1028))
- feat:(roles): Importable role by rolename ([#1043](https://github.com/keycloak/terraform-provider-keycloak/pull/1043))
- feat(user): Importable user by username ([#1041](https://github.com/keycloak/terraform-provider-keycloak/pull/1041))
- feat(authFlow): added optional priority ([#1040](https://github.com/keycloak/terraform-provider-keycloak/pull/1040))
- feat: Recognise the 'apple' provider ID in attribute importer mapper ([#1036](https://github.com/keycloak/terraform-provider-keycloak/pull/1036))
- Add support for remote debugging ([#1048](https://github.com/keycloak/terraform-provider-keycloak/pull/1048))
- unit tests are now working with KC 21 to 26 ([#1028](https://github.com/keycloak/terraform-provider-keycloak/pull/1028))
	- Please check IdP provider sync mode as the default has changed to "LEGACY"
	- Keycloak 25: SAML clients have a default 'saml_organization'. If 'saml_organization' isn't specified in the provider configuration, the provider will delete this scope.

IMPROVEMENTS:

- Add attribute_force_default to ldap_user_attribute_mapper ([#1057](https://github.com/keycloak/terraform-provider-keycloak/pull/1057))
- chore: Ignore local scratch dir ([#1063](https://github.com/keycloak/terraform-provider-keycloak/pull/1063))
- Code cleanups ([#1050](https://github.com/keycloak/terraform-provider-keycloak/pull/1050))
- Updated dependencies ([#1051](https://github.com/keycloak/terraform-provider-keycloak/pull/1051))
- Updated example to KC26.0.7 ([#1046](https://github.com/keycloak/terraform-provider-keycloak/pull/1046))
- update golang.org/x/net and golang.org/x/crypto ([#1034](https://github.com/keycloak/terraform-provider-keycloak/pull/1034))
- Update index.md: The default client timeout is 15 seconds ([#1018](https://github.com/keycloak/terraform-provider-keycloak/pull/1018))
- Small readme fixed and version updates ([#1035](https://github.com/keycloak/terraform-provider-keycloak/pull/1035))
- Update expired certificates ([#1045](https://github.com/keycloak/terraform-provider-keycloak/pull/1045))
- Fixed examples and user-federation-example ([#1039](https://github.com/keycloak/terraform-provider-keycloak/pull/1039))

BUG FIXES:

- Prevent force replacing resources when `import` value in state is nil ([#1054](https://github.com/keycloak/terraform-provider-keycloak/pull/1054))
- fix: Revise Keycloak environment handling (#1061) ([#1062](https://github.com/keycloak/terraform-provider-keycloak/pull/1062))
- fix: Revise Keycloak debugging configuration (#1059) ([#1060](https://github.com/keycloak/terraform-provider-keycloak/pull/1060))

Huge thanks to all the individuals who have contributed towards this release:

- [@AbrohamLincoln](https://github.com/AbrohamLincoln)
- [@denniskniep](https://github.com/denniskniep)
- [@giner](https://github.com/giner)
- [@greed42](https://github.com/greed42)
- [@markus-qvest-seidl](https://github.com/markus-qvest-seidl)
- [@sschum](https://github.com/sschum)
- [@kherock](https://github.com/kherock)
- [@Vincevrp](https://github.com/Vincevrp)

## 4.5.0 (December 6, 2024)

IMPROVEMENTS:

- various dependency and readme updates ([#1014](https://github.com/keycloak/terraform-provider-keycloak/pull/1014), [#1015](https://github.com/keycloak/terraform-provider-keycloak/pull/1015), [#1022](https://github.com/keycloak/terraform-provider-keycloak/pull/1022), [#1024](https://github.com/keycloak/terraform-provider-keycloak/pull/1024))
- move to the Keycloak organization ([#1017](https://github.com/keycloak/terraform-provider-keycloak/pull/1017))
- update go to 1.22 ([#1013](https://github.com/keycloak/terraform-provider-keycloak/pull/1013))
- update license for project to Apache 2.0 [#1002](https://github.com/keycloak/terraform-provider-keycloak/pull/1002)

Huge thanks to everybody supporting the license change to Apache 2.0!

## 4.4.0 (January 8, 2024)

FEATURES:

- new resource: `keycloak_ldap_custom_mapper` ([#863](https://github.com/mrparkers/terraform-provider-keycloak/pull/863))

IMPROVEMENTS:

- add `provider_id` attribute to `keycloak_realm_keystore_rsa` resource ([#858](https://github.com/mrparkers/terraform-provider-keycloak/pull/858))
- compute `extra_config` for client data sources ([#885](https://github.com/mrparkers/terraform-provider-keycloak/pull/885))
- support updated RedHat SSO version strings ([#903](https://github.com/mrparkers/terraform-provider-keycloak/pull/903))
- add `required_actions` attribute to `keycloak_user` resource ([#867](https://github.com/mrparkers/terraform-provider-keycloak/pull/867))
- allow use of custom client authenticator types ([#845](https://github.com/mrparkers/terraform-provider-keycloak/pull/845))
- allow users to set `referrer_policy` header within `security_defenses` for `keycloak_realm` resource ([#879](https://github.com/mrparkers/terraform-provider-keycloak/pull/879))

BUG FIXES:

- set the correct value for `identityProviderMapper` when using keycloak-oidc identity provider ([#850](https://github.com/mrparkers/terraform-provider-keycloak/pull/850))

Huge thanks to all the individuals who have contributed towards this release:

- [@NataliaKhodiakova](https://github.com/NataliaKhodiakova)
- [@olivierboudet](https://github.com/olivierboudet)
- [@guthypeter](https://github.com/guthypeter)
- [@kherock](https://github.com/kherock)
- [@smoehrle](https://github.com/smoehrle)
- [@wombat](https://github.com/wombat)
- [@scheying](https://github.com/scheying)
- [@JackHartley](https://github.com/JackHartley)
- [@stale-vegait](https://github.com/stale-vegait)
- [@jrpedrianes](https://github.com/jrpedrianes)

## 4.3.1 (June 8, 2023)

BUG FIXES:

- avoid validation errors during plan for missing realm and openid client resources ([#841](https://github.com/mrparkers/terraform-provider-keycloak/pull/841))

Huge thanks to all the individuals who have contributed towards this release:

- [@kherock](https://github.com/kherock)

## v4.3.0 (June 5, 2023)

IMPROVEMENTS:

- update supported Keycloak versions ([#828](https://github.com/mrparkers/terraform-provider-keycloak/pull/828))
- update to go 1.19 ([#848](https://github.com/mrparkers/terraform-provider-keycloak/pull/848))
- add new `login_hint` attribute for `keycloak_saml_identity_provider` resource ([#830](https://github.com/mrparkers/terraform-provider-keycloak/pull/830))

BUG FIXES:

- avoid inserting authenticator execution configs with empty strings as ID ([#840](https://github.com/mrparkers/terraform-provider-keycloak/pull/840))

Huge thanks to all the individuals who have contributed towards this release:

- [@malfter](https://github.com/malfter)
- [@toddkazakov](https://github.com/toddkazakov)
- [@mohammedalics](https://github.com/mohammedalics)
- [@kherock](https://github.com/kherock)

## v4.2.0 (March 6, 2023)

IMPROVEMENTS:

- allow the `internal_id` attribute for the `keycloak_realm` resource to be set during apply instead of read-only ([#807](https://github.com/mrparkers/terraform-provider-keycloak/pull/807))
- allow for multivalue attributes in `extra_config` attribute for `keycloak_custom_user_federation` resource ([#761](https://github.com/mrparkers/terraform-provider-keycloak/pull/761))

BUG FIXES:

- allow users with backslashes in their name to be assigned to groups via `keycloak_group_memberships` resource ([#778](https://github.com/mrparkers/terraform-provider-keycloak/pull/778))
- correctly set `nameIDPolicyFormat` when updating value in `extra_config` in `keycloak_saml_identity_provider` resource ([#793](https://github.com/mrparkers/terraform-provider-keycloak/pull/793))
- treat empty attributes as nil values when importing `keycloak_ldap_user_federation` resource ([#784](https://github.com/mrparkers/terraform-provider-keycloak/pull/784))
- treat empty attributes as nil values when importing `keycloak_custom_user_federation` resource ([#809](https://github.com/mrparkers/terraform-provider-keycloak/pull/809))

Huge thanks to all the individuals who have contributed towards this release:

- [@pablo-ruth](https://github.com/pablo-ruth)
- [@Redestros](https://github.com/Redestros)
- [@MatrixCrawler](https://github.com/MatrixCrawler)
- [@imykolenko](https://github.com/imykolenko)
- [@ChrisHubinger](https://github.com/ChrisHubinger)
- [@chifu1234](https://github.com/chifu1234)

## v4.1.0 (December 4, 2022)

IMPROVEMENTS:

- add `IMPORT` mode to `keycloak_ldap_role_mapper` resource ([#768](https://github.com/mrparkers/terraform-provider-keycloak/pull/768))
- add `RSA_SHA256_MGF1` and `RSA_SHA512_MGF1` signature algorithms to `keycloak_saml_client` resource ([#757](https://github.com/mrparkers/terraform-provider-keycloak/pull/757))
- add `valid_post_logout_redirect_uris` attribute to `keycloak_openid_client` resource ([#777](https://github.com/mrparkers/terraform-provider-keycloak/pull/777))

BUG FIXES:

- fix incorrect import ID for `keycloak_openid_client_authorization_*` resources ([#763](https://github.com/mrparkers/terraform-provider-keycloak/pull/763))
- fix payload used during deletion of `keycloak_generic_role_mapper` resource to prevent more mappers from unintentionally being removed ([#772](https://github.com/mrparkers/terraform-provider-keycloak/pull/772))

## v4.0.1 (October 13, 2022)

BUG FIXES:

- restored the default value for the `client_authenticator_type` attribute within the `keycloak_openid_client` resource ([#755](https://github.com/mrparkers/terraform-provider-keycloak/pull/755))

## v4.0.0 (October 10, 2022)

BREAKING CHANGES:

- updated the default value of the `base_path` provider attribute, it is now an empty string ([#733](https://github.com/mrparkers/terraform-provider-keycloak/pull/733))
  - this change was made due to the Quarkus distribution of Keycloak removing the `/auth` context from API urls. if you
    are currently using the Quarkus version of Keycloak, you no longer need to specify the `base_path` provider attribute
    as an empty string. if you are currently using the legacy version of Keycloak, you will need to add the `base_path`
    provider attribute and set it to `/auth`.
- renamed resources:
  - `keycloak_generic_client_protocol_mapper` has been renamed to `keycloak_generic_protocol_mapper` ([#742](https://github.com/mrparkers/terraform-provider-keycloak/pull/742))
  - `keycloak_generic_client_role_mapper` has been renamed to `keycloak_generic_role_mapper`. ([#748](https://github.com/mrparkers/terraform-provider-keycloak/pull/748))
  - the old versions of these resources will remain functional until the next major release, but they will display a deprecation
    warning whenever they are used.
  - to migrate to these new resources, you can follow these steps:
    - use `terraform state rm` to remove each of the old resources from state.
    - use `terraform import` to import the new resources into state. you can refer to the documentation for each of these
      resources to see how they should be imported.

FEATURES:

- new resource: `keycloak_ldap_hardcoded_attribute_mapper` ([#725](https://github.com/mrparkers/terraform-provider-keycloak/pull/725))
- new data source: `keycloak_openid_client_scope` ([#743](https://github.com/mrparkers/terraform-provider-keycloak/pull/743))

IMPROVEMENTS:

- add `red_hat_sso` provider attribute which can be set to `true` if you're using RedHat SSO. this helps the provider understand which version of Keycloak is being used ([#721](https://github.com/mrparkers/terraform-provider-keycloak/pull/721))
- support json encoded validation config for `keycloak_realm_user_profile` resource ([#705](https://github.com/mrparkers/terraform-provider-keycloak/pull/705))
- update go version to 1.18, update several dependencies, update supported Keycloak versions to include v19 ([#733](https://github.com/mrparkers/terraform-provider-keycloak/pull/733))
- add `attribute_default_value` and `is_binary_attribute` attributes to `keycloak_ldap_user_attribute_mapper` resource ([#735](https://github.com/mrparkers/terraform-provider-keycloak/pull/735))
- update `keycloak_ldap_user_federation` resource to add support for deleting default mappers that are normally created by Keycloak ([#744](https://github.com/mrparkers/terraform-provider-keycloak/pull/744))
- add `issuer` attribute to `keycloak_oidc_identity_provider` resource ([#746](https://github.com/mrparkers/terraform-provider-keycloak/pull/746))
- update `keycloak_openid_client` resource to add support for importing Keycloak-created clients without needing to run `terraform import` ([#747](https://github.com/mrparkers/terraform-provider-keycloak/pull/747))

Huge thanks to all the individuals who have contributed towards this release:

- [@jermarchand](https://github.com/jermarchand)
- [@Useurmind](https://github.com/Useurmind)
- [@bplotnick-humane](https://github.com/bplotnick-humane)
- [@joeyberkovitz](https://github.com/joeyberkovitz)
- [@gerbermichi](https://github.com/gerbermichi)
- [@meckhardt](https://github.com/meckhardt)
- [@spirius](https://github.com/spirius)

## v3.10.0 (July 28, 2022)

IMPROVEMENTS:

- add authn context attributes for `keycloak_saml_identity_provider` resource ([#703](https://github.com/mrparkers/terraform-provider-keycloak/pull/703))
- add `resource_type` attribute for `keycloak_openid_client_authorization_permission` resource ([#702](https://github.com/mrparkers/terraform-provider-keycloak/pull/702))

Huge thanks to all the individuals who have contributed towards this release:

- [@dnic](https://github.com/dnic)
- [@1337andre](https://github.com/1337andre)
- [@JessieAMorris](https://github.com/JessieAMorris)

## v3.9.1 (July 11, 2022)

BUG FIXES:

- update usage of component API for `keycloak_ldap_user_federation` and `keycloak_custom_user_federation` resources ([#707](https://github.com/mrparkers/terraform-provider-keycloak/pull/707))
  - this fixes an issue that prevented these resources from being used within the `master` realm.

## v3.9.0 (June 23, 2022)

IMPROVEMENTS:

- improve import error messages for several resources ([#691](https://github.com/mrparkers/terraform-provider-keycloak/pull/691))
- allow usage of environment variable to configure base API path ([#695](https://github.com/mrparkers/terraform-provider-keycloak/pull/695))

BUG FIXES:

- use realm name instead of internal ID for authentication bindings ([#687](https://github.com/mrparkers/terraform-provider-keycloak/pull/687))

Huge thanks to all the individuals who have contributed towards this release:

- [@dmeyerholt](https://github.com/dmeyerholt)
- [@Korsarro69](https://github.com/Korsarro69)

## v3.8.1 (May 9, 2022)

BUG FIXES:

- fix a potential problem with the `keycloak_custom_user_federation` resource incorrectly assuming some Keycloak API fields are numbers.

## v3.8.0 (May 4, 2022)

FEATURES:

- new resource: `keycloak_realm_user_profile` ([#658](https://github.com/mrparkers/terraform-provider-keycloak/pull/658))
- new resource: `keycloak_authentication_bindings` ([#668](https://github.com/mrparkers/terraform-provider-keycloak/pull/668))

IMPROVEMENTS:

- support custom provider ID in `keycloak_saml_identity_provider` resource ([#656](https://github.com/mrparkers/terraform-provider-keycloak/pull/656))
- support sync settings in `keycloak_custom_user_federation` resource ([#663](https://github.com/mrparkers/terraform-provider-keycloak/pull/663))
- support Transient NameID format for `keycloak_saml_identity_provider` resource ([#661](https://github.com/mrparkers/terraform-provider-keycloak/pull/661))
- update all resources to use new terraform lifecycles with context support ([#675](https://github.com/mrparkers/terraform-provider-keycloak/pull/675))
- support use-refresh-tokens for client credentials in `keycloak_openid_client` resource ([#678](https://github.com/mrparkers/terraform-provider-keycloak/pull/678))
- support `client_session_idle_timeout` and `client_session_max_lifespan` arguments in `keycloak_realm` resource ([#653](https://github.com/mrparkers/terraform-provider-keycloak/pull/653))

Huge thanks to all the individuals who have contributed towards this release:

- [@tomrutsaert](https://github.com/tomrutsaert)
- [@maximepiton](https://github.com/maximepiton)
- [@marwol-fdir](https://github.com/marwol-fdir)
- [@puzzlermike](https://github.com/puzzlermike)
- [@maximepiton](https://github.com/maximepiton)
- [@camjjack](https://github.com/camjjack)
- [@daviddelannoy](https://github.com/daviddelannoy)
- [@fapian](https://github.com/fapian)

## v3.7.0 (February 2, 2022)

IMPROVEMENTS:

- add support for the oauth2 device authorization grant ([#578](https://github.com/mrparkers/terraform-provider-keycloak/pull/578))
- add `client_authenticator_type` attribute to `keycloak_openid_client` resource ([#627](https://github.com/mrparkers/terraform-provider-keycloak/pull/627))
- add missing documentation for `keycloak_user_template_importer_identity_provider_mapper` resource ([#635](https://github.com/mrparkers/terraform-provider-keycloak/pull/635))
- add attributes for customizing consent screen for `keycloak_openid_client` resource ([#646](https://github.com/mrparkers/terraform-provider-keycloak/pull/646))
- upgrade to the latest version of the `terraform-plugin-sdk` ([#644](https://github.com/mrparkers/terraform-provider-keycloak/pull/644))
- add attributes for configuring frontchannel logout on `keycloak_openid_client` resource ([#644](https://github.com/mrparkers/terraform-provider-keycloak/pull/644))
- bump supported keycloak versions ([#650](https://github.com/mrparkers/terraform-provider-keycloak/pull/650))

BUG FIXES:

- fix keycloak version check for `keycloak_default_roles` resource ([#637](https://github.com/mrparkers/terraform-provider-keycloak/pull/637))

Huge thanks to all the individuals who have contributed towards this release:

- [@shellrausch](https://github.com/shellrausch)
- [@m-v-k](https://github.com/m-v-k)
- [@oysteinhauan](https://github.com/oysteinhauan)
- [@Kidsan](https://github.com/Kidsan)

## v3.6.0 (November 9, 2021)

FEATURES:

- new resource: `keycloak_group_permissions` ([#617](https://github.com/mrparkers/terraform-provider-keycloak/pull/617))

BUG FIXES:

- `xml_sign_key_info_key_name_transformer` attribute for `keycloak_saml_identity_provider` resource used incorrect spelling, causing it to not be set correctly ([#614](https://github.com/mrparkers/terraform-provider-keycloak/pull/614))
- when querying protocol mappers from the Keycloak API, treat quoted boolean attributes as `false` when receiving an empty string. this should fix issues when importing protocol mappers that were created by Keycloak ([#622](https://github.com/mrparkers/terraform-provider-keycloak/pull/622))

Huge thanks to all the individuals who have contributed towards this release:

- [@jkemming](https://github.com/jkemming)
- [@hoeggi](https://github.com/hoeggi)

## v3.5.1 (October 14, 2021)

BUG FIXES:

- remove `defaultRole` from realm JSON before sending requests to Keycloak to fix compatibility with Keycloak versions older than v13 ([#612](https://github.com/mrparkers/terraform-provider-keycloak/pull/612))

## v3.5.0 (October 13, 2021)

FEATURES:

- new resource: `keycloak_default_roles` ([#599](https://github.com/mrparkers/terraform-provider-keycloak/pull/599))
- new resources: `keycloak_realm_keystore_aes_generated`, `keycloak_realm_keystore_ecdsa_generated`, `keycloak_realm_keystore_hmac_generated`, `keycloak_realm_keystore_java_keystore`, `keycloak_realm_keystore_rsa`, and `keycloak_realm_keystore_rsa_generated` ([#582](https://github.com/mrparkers/terraform-provider-keycloak/pull/582))
- new resource: `keycloak_openid_audience_resolve_protocol_mapper` ([#606](https://github.com/mrparkers/terraform-provider-keycloak/pull/606))

IMPROVEMENTS:

- add `start_tls` and `use_password_modify_extended_op` attributes to `keycloak_ldap_user_federation` resource ([#601](https://github.com/mrparkers/terraform-provider-keycloak/pull/601))
- `keycloak_openid_client_default_scopes` and `keycloak_openid_client_optional_scopes` resources will now completely reconcile assigned scopes on create ([#594](https://github.com/mrparkers/terraform-provider-keycloak/pull/594))
  - this means that creating these resources will now remove default / optional scopes that are not specified within the resource configuration. see [#498](https://github.com/mrparkers/terraform-provider-keycloak/issues/498) for more context.

BUG FIXES:

- allow all `extra_config` attributes for `keycloak_custom_identity_provider_mapper` resource ([#607](https://github.com/mrparkers/terraform-provider-keycloak/pull/607))
- `backchannel_logout_session_required` and `backchannel_logout_revoke_offline_sessions` attributes for `keycloak_openid_client` resource were swapped ([#600](https://github.com/mrparkers/terraform-provider-keycloak/pull/600))

Huge thanks to all the individuals who have contributed towards this release:

- [@Vlad-Kirichenko](https://github.com/Vlad-Kirichenko)
- [@thyming](https://github.com/thyming)
- [@francois-travais](https://github.com/francois-travais)

## v3.4.0 (September 16, 2021)

FEATURES:

- add backchannel support for `keycloak_openid_client` resource ([#583](https://github.com/mrparkers/terraform-provider-keycloak/pull/583))
- add support for `extra_config` for `keycloak_openid_client` resource ([#579](https://github.com/mrparkers/terraform-provider-keycloak/pull/579))
- add support for `extra_config` for `keycloak_saml_client` resource ([#589](https://github.com/mrparkers/terraform-provider-keycloak/pull/589))
- add `signature_key_name` attribute to `keycloak_saml_client` resource ([#588](https://github.com/mrparkers/terraform-provider-keycloak/pull/588))
- add `login_theme` attribute to `keycloak_saml_client` resource ([#590](https://github.com/mrparkers/terraform-provider-keycloak/pull/590))
- new data source: `keycloak_user_realm_roles` ([#596](https://github.com/mrparkers/terraform-provider-keycloak/pull/596))
- add OTP policy attributes to `keycloak_realm` resource ([#585](https://github.com/mrparkers/terraform-provider-keycloak/pull/585))
- add computed attributes `encryption_certificate_sha1`, `signing_certificate_sha1`, and `signing_private_key_sha1` for `keycloak_saml_client` resource ([#589](https://github.com/mrparkers/terraform-provider-keycloak/pull/589))

IMPROVEMENTS:

- the behavior of the `extra_config` attribute among all resources that support it has been standardized ([#584](https://github.com/mrparkers/terraform-provider-keycloak/pull/584)) ([#589](https://github.com/mrparkers/terraform-provider-keycloak/pull/589))
  - validation has been added to ensure that `extra_config` can't be used to override values that are supported by that particular resource's top level schema
  - `extra_config` will no longer contain "computed" attributes, meaning that attributes not supplied by the user will not be written back to `extra_config`
  - attributes that have been removed from `extra_config` will be sent back to the Keycloak API as an empty string. this appears to be the only way to "unset" these on the Keycloak server
  - REMINDER: `extra_config` should only be used to support custom attributes, or attributes that are not yet officially supported by this provider. future releases of this provider could cause breaking changes for users using `extra_config`. please use this attribute carefully, especially when upgrading to newer versions of the provider.
- request / response bodies to / from the Keycloak API will be properly formatted when `TF_LOG` is set to `DEBUG` ([#589](https://github.com/mrparkers/terraform-provider-keycloak/pull/589))
- the list of officially supported Keycloak versions has been updated to 13.x, 14.x, and 15.x ([#589](https://github.com/mrparkers/terraform-provider-keycloak/pull/589)). older versions may still work, but they will no longer be tested against in CI.
- the behavior of the `keycloak_saml_client` attributes `encryption_certificate`, `signing_certificate`, and `signing_private_key` has been changed.
  - previously, it was meant to be possible to unset these attributes by setting them to an empty string. this was meant to remove the certs / keys on the Keycloak server. however, this never really worked correctly, so this behavior has been removed.
  - these values will now be autogenerated by Keycloak when omitted.

BUG FIXES:

- fix possible crash when using `keycloak_users_permissions` resource ([#591](https://github.com/mrparkers/terraform-provider-keycloak/pull/591))

Huge thanks to all the individuals who have contributed towards this release:

- [@cw1o](https://github.com/cw1o)
- [@daviddelannoy](https://github.com/daviddelannoy)
- [@f-stibane](https://github.com/f-stibane)
- [@jjarman-infinity](https://github.com/jjarman-infinity)
- [@nolte](https://github.com/nolte)
- [@olivierboudet](https://github.com/olivierboudet)

## v3.3.0 (August 9, 2021)

IMPROVEMENTS:

- add `use_refresh_tokens` attribute to `keycloak_openid_client` resource ([#573](https://github.com/mrparkers/terraform-provider-keycloak/pull/573))

Huge thanks to all the individuals who have contributed towards this release:

- [@whiskeysierra](https://github.com/whiskeysierra)

## v3.2.1 (July 23, 2021)

BUG FIXES:

- re-add previously removed `LOAD_ROLES_BY_MEMBER_ATTRIBUTE_RECURSIVELY` role retrieval strategy for the `keycloak_ldap_role_mapper` resource ([#560](https://github.com/mrparkers/terraform-provider-keycloak/pull/560))
- perform initial login during version check if needed. this fixes a potential panic within the `keycloak_ldap_group_mapper` resource ([#564](https://github.com/mrparkers/terraform-provider-keycloak/pull/564))

Huge thanks to all the individuals who have contributed towards this release:

- [@DOboznyi](https://github.com/DOboznyi)
- [@Kent1](https://github.com/Kent1)

## v3.2.0 (July 14, 2021)

IMPROVEMENTS:

- stopped throwing an error for missing provider credentials when `initial_login` is set to `false`. this should help with scenarios where Keycloak itself is being created by Terraform (such as with the `helm_release` resource) ([#552](https://github.com/mrparkers/terraform-provider-keycloak/pull/552))
- upgrade to go v1.16, bump terraform plugin SDK ([#551](https://github.com/mrparkers/terraform-provider-keycloak/pull/551))
  - this enables builds for previously unsupported platforms, such as `darwin_arm64`
  - this should fix any potential issues with using this provider with Terraform v1.0.1

BUG FIXES:

- fix possible panic when creating identity provider mappers ([#556](https://github.com/mrparkers/terraform-provider-keycloak/pull/556))

Huge thanks to all the individuals who have contributed towards this release:

- [@meckhardt](https://github.com/meckhardt)

## v3.1.1 (June 8, 2021)

There was an internal problem with the v3.1.0 release, causing a checksum error when running `terraform init`.  Please use
this release instead.

## v3.1.0 (June 8, 2021)

An internal error during the release process caused this release to fail when running `terraform init`.  Please use v3.1.1
instead.

FEATURES:

- new resource: `keycloak_custom_identity_provider_mapper` ([#515](https://github.com/mrparkers/terraform-provider-keycloak/pull/515))
- new data source: `keycloak_client_description_converter` ([#518](https://github.com/mrparkers/terraform-provider-keycloak/pull/518))

IMPROVEMENTS:

- use pagination for `keycloak_group_memberships` resource ([#527](https://github.com/mrparkers/terraform-provider-keycloak/pull/527))

BUG FIXES:

- handle deleted role when removing role assignment from `keycloak_group_roles` resource ([#538](https://github.com/mrparkers/terraform-provider-keycloak/pull/538))

Huge thanks to all the individuals who have contributed towards this release:

- [@bl00mber](https://github.com/bl00mber)
- [@hamiltont](https://github.com/hamiltont)
- [@Kyos](https://github.com/Kyos)
- [@pstanton237](https://github.com/pstanton237)
- [@sl-benoitoyez](https://github.com/sl-benoitoyez)

## v3.0.1 (May 5, 2021)

BUG FIXES:

- add validation for `extra_config` attribute for identity providers to prevent conflicts with the top-level identity provider schema ([#523](https://github.com/mrparkers/terraform-provider-keycloak/pull/523))
  - note: this may cause errors with existing provider configuration that uses this attribute. however, any provider configuration that breaks here was most likely not working in the first place.
- fix definition of roles in `keycloak_openid_client_role_policy` resource to use a set instead of a list ([#524](https://github.com/mrparkers/terraform-provider-keycloak/pull/524))

## v3.0.0 (April 12, 2021)

BREAKING CHANGES:

- add a new required `entity_id` attribute for `keycloak_saml_identity_provider` resource ([#512](https://github.com/mrparkers/terraform-provider-keycloak/pull/512))
- removed attributes that were deprecated in v2.0.0 ([#514](https://github.com/mrparkers/terraform-provider-keycloak/pull/514))
  - `keycloak_openid_user_session_note_protocol_mapper` resource: remove `session_note_label` attribute
  - `keycloak_user` data source: remove `federated_identities` attribute
  - `keycloak_ldap_user_federation` resource: remove `cache_policy` attribute

FEATURES:

- new data source: `keycloak_authentication_flow` ([#486](https://github.com/mrparkers/terraform-provider-keycloak/pull/486))
- new resource: `keycloak_user_groups` ([#505](https://github.com/mrparkers/terraform-provider-keycloak/pull/505))

IMPROVEMENTS:

- support multivalue attributes for users, groups and roles ([#499](https://github.com/mrparkers/terraform-provider-keycloak/pull/499))
- add `trust_email` attribute to `keycloak_ldap_user_federation` resource ([#267](https://github.com/mrparkers/terraform-provider-keycloak/pull/267))
- add `principal_type`, `principal_attribute`, `gui_order`, and `sync_mode` attributes to `keycloak_saml_identity_provider` resource ([#508](https://github.com/mrparkers/terraform-provider-keycloak/pull/508))
- allows non-authoritative usage of `keycloak_group_roles` resource via `exhaustive` attribute ([#501](https://github.com/mrparkers/terraform-provider-keycloak/pull/501))
- allows non-authoritative usage of `keycloak_user_roles` resource via `exhaustive` attribute ([#513](https://github.com/mrparkers/terraform-provider-keycloak/pull/513))
- add ability to set additional request headers as provider config ([#507](https://github.com/mrparkers/terraform-provider-keycloak/pull/507))

BUG FIXES:

- fixed marshalling of `false` value in Keycloak API attributes that use quoted booleans ([#495](https://github.com/mrparkers/terraform-provider-keycloak/pull/495))
- handle group not found for `keycloak_group_roles` resource ([#497](https://github.com/mrparkers/terraform-provider-keycloak/pull/497))
- fix `keycloak_attribute_importer_identity_provider_mapper` and `keycloak_user_template_importer_identity_provider_mapper` resources for usage with Facebook/Google ([#482](https://github.com/mrparkers/terraform-provider-keycloak/pull/482))

Huge thanks to all the individuals who have contributed towards this release:

- [@alex-hempel](https://github.com/alex-hempel)
- [@lathspell](https://github.com/lathspell)
- [@max-rocket-internet](https://github.com/max-rocket-internet)
- [@Photonios](https://github.com/Photonios)
- [@PSanetra](https://github.com/PSanetra)
- [@sl-benoitoyez](https://github.com/sl-benoitoyez)
- [@StatueFungus](https://github.com/StatueFungus)
- [@vlaurin](https://github.com/vlaurin)
- [@yesteph](https://github.com/yesteph)
- [@Zeldhyr](https://github.com/Zeldhyr)

## v2.3.0 (March 1, 2021)

FEATURES:

- new resource: `keycloak_saml_script_protocol_mapper` ([#473](https://github.com/mrparkers/terraform-provider-keycloak/pull/473))

IMPROVEMENTS:

- support custom attributes in `keycloak_role` resource ([#475](https://github.com/mrparkers/terraform-provider-keycloak/pull/475))

BUG FIXES:

- remove mutex usage in keycloak client, which in some cases resulted in deadlock when retrieving tokens from Keycloak ([#489](https://github.com/mrparkers/terraform-provider-keycloak/pull/489))

Huge thanks to all the individuals who have contributed towards this release:

- [@dbolack](https://github.com/dbolack)
- [@dullest](https://github.com/dullest)
- [@lathspell](https://github.com/lathspell)

## v2.2.0 (January 23, 2021)

FEATURES:

- add new `keycloak_realm` attributes for handling default client scopes ([#464](https://github.com/mrparkers/terraform-provider-keycloak/pull/464))
- new data source: `keycloak_saml_client` ([#468](https://github.com/mrparkers/terraform-provider-keycloak/pull/468))

IMPROVEMENTS:

- revised the configuration for the custom user federation example ([#425](https://github.com/mrparkers/terraform-provider-keycloak/pull/425))
- increased the default http client timeout to 15 seconds ([#469](https://github.com/mrparkers/terraform-provider-keycloak/pull/469))

BUG FIXES:

- fix panic when using `keycloak_user` data source with invalid username ([#460](https://github.com/mrparkers/terraform-provider-keycloak/pull/460))
- fix version handling with RedHat SSO ([#462](https://github.com/mrparkers/terraform-provider-keycloak/pull/462))

Huge thanks to all the individuals who have contributed towards this release:

- [@Filirom1](https://github.com/Filirom1)
- [@thomasdarimont](https://github.com/thomasdarimont)

## v2.1.0 (January 10, 2021)

FEATURES:

- new resource: `keycloak_openid_client_permissions` ([#364](https://github.com/mrparkers/terraform-provider-keycloak/pull/364))
- new resource: `keycloak_users_permissions` ([#400](https://github.com/mrparkers/terraform-provider-keycloak/pull/400))
- new resource: `keycloak_openid_client_script_protocol_mapper` ([#453](https://github.com/mrparkers/terraform-provider-keycloak/pull/453))

IMPROVEMENTS:

- add `authorization.decision_strategy` attribute to `keycloak_openid_client` resource ([#392](https://github.com/mrparkers/terraform-provider-keycloak/pull/392))
- support `IMPORT` mode for `keycloak_ldap_group_mapper` resource ([#397](https://github.com/mrparkers/terraform-provider-keycloak/pull/397))
- add client session length attributes to `keycloak_openid_client` resource ([#415](https://github.com/mrparkers/terraform-provider-keycloak/pull/415))
- update to go 1.5 ([#445](https://github.com/mrparkers/terraform-provider-keycloak/pull/360))
- add `groups_path` attribute to `keycloak_ldap_group_mapper` resource ([#436](https://github.com/mrparkers/terraform-provider-keycloak/pull/436))
- add `authentication_flow_binding_overrides` attribute to `keycloak_saml_client` resource ([#448](https://github.com/mrparkers/terraform-provider-keycloak/pull/448))

BUG FIXES:

- fix inconsistent plan when enabling service account in `keycloak_openid_client` resource ([#437](https://github.com/mrparkers/terraform-provider-keycloak/pull/437))
- fix import for `keycloak_openid_client_service_account_realm_role` resource ([#441](https://github.com/mrparkers/terraform-provider-keycloak/pull/441))
- remove unneeded validation checks for registration attributes for `keycloak_realm` resource ([#438](https://github.com/mrparkers/terraform-provider-keycloak/pull/438))
- allow commas in `config` attribute for `keycloak_custom_user_federation` resource ([#455](https://github.com/mrparkers/terraform-provider-keycloak/pull/455))

Huge thanks to all the individuals who have contributed towards this release:

- [@AdrienFromToulouse](https://github.com/AdrienFromToulouse)
- [@hcl31415](https://github.com/hcl31415)
- [@jermarchand](https://github.com/jermarchand)
- [@PaulGgithub](https://github.com/PaulGgithub)
- [@pths](https://github.com/pths)
- [@randomswdev](https://github.com/randomswdev)
- [@spirius](https://github.com/spirius)
- [@toddkazakov](https://github.com/toddkazakov)
- [@xinau](https://github.com/xinau)

## v2.0.0 (September 20, 2020)

BREAKING CHANGES:

- migrate to v2 of the terraform-plugin-sdk, which [drops support for Terraform 0.11 and below](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#dropped-support-for-terraform-0-11-and-below) ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))

DEPRECATIONS:

- the `cache_policy` attribute within the `keycloak_ldap_user_federation` resource has been deprecated in favor of a new `cache` attribute ([#376](https://github.com/mrparkers/terraform-provider-keycloak/pull/376))
- the `federated_identities` computed attribute within the `keycloak_user` data source has been deprecated in favor of a new `federated_identity` computed attribute ([1b6284c](https://github.com/mrparkers/terraform-provider-keycloak/commit/1b6284c70dbdb67f42fafe16abeb681541d06cbf))
- the `session_note_label` attribute within the `keycloak_openid_user_session_note_protocol_mapper` resource has been deprecated in favor of a new `session_note` attribute ([#365](https://github.com/mrparkers/terraform-provider-keycloak/pull/365))

FEATURES:

- this provider can now be installed automatically with Terraform 0.13 via the Terraform registry: https://registry.terraform.io/providers/mrparkers/keycloak/latest
- new data source: `keycloak_user` ([#360](https://github.com/mrparkers/terraform-provider-keycloak/pull/360))
- new data source: `keycloak_authentication_execution` ([#360](https://github.com/mrparkers/terraform-provider-keycloak/pull/360))

IMPROVEMENTS:

- add remember me timeout attributes to `keycloak_realm` resource ([#374](https://github.com/mrparkers/terraform-provider-keycloak/pull/374))
- add `offline_session_max_lifespan_enabled` attribute to `keycloak_realm` resource ([#377](https://github.com/mrparkers/terraform-provider-keycloak/pull/377))
- add `web_authn_policy` and `web_authn_passwordless_policy` attributes to `keycloak_realm` resource ([#356](https://github.com/mrparkers/terraform-provider-keycloak/pull/356))

BUG FIXES:

- fix `keycloak_group` data source to support more than one returned group ([#351](https://github.com/mrparkers/terraform-provider-keycloak/pull/351))
- fix import syntax for `keycloak_openid_client_*_policy` resources ([#367](https://github.com/mrparkers/terraform-provider-keycloak/pull/367))
- fix `parent_id` attribute not being set when importing `keycloak_group` resource ([#372](https://github.com/mrparkers/terraform-provider-keycloak/pull/372))
- automatically register an unregistered required action when using the `keycloak_required_action` resource ([#385](https://github.com/mrparkers/terraform-provider-keycloak/pull/385))
- fix `keycloak_openid_user_session_note_protocol_mapper` resource API call to correctly set the session note ([#365](https://github.com/mrparkers/terraform-provider-keycloak/pull/365))
- add missing attributes for `keycloak_group` data source ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))
- add missing attributes for `keycloak_openid_client_service_account_user` data source ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))
- add missing attributes for `keycloak_realm` data source ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))
- fix `config` attribute for `keycloak_custom_user_federation` resource ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))
- fix `kerberos` attribute for `keycloak_ldap_user_federation` resource ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))
- add missing `disable_user_info` attribute for `keycloak_oidc_identity_provider` resource ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))
- fix empty `path` sub-attribute under `groups` attribute within `keycloak_openid_client_authorization_group_policy` resource ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))
- fix `role` attribute for `keycloak_openid_client_authorization_role_policy` resource ([#369](https://github.com/mrparkers/terraform-provider-keycloak/pull/369))

Huge thanks to all the individuals who have contributed towards this release:

- [@cgroschupp](https://github.com/cgroschupp)
- [@gansb](https://github.com/gansb)
- [@hcl31415](https://github.com/hcl31415)
- [@jermarchand](https://github.com/jermarchand)
- [@klausenbusk](https://github.com/klausenbusk)
- [@paulvollmer](https://github.com/paulvollmer)
- [@pmellati](https://github.com/pmellati)
- [@rjmasikome](https://github.com/rjmasikome)
- [@RomanNess](https://github.com/RomanNess)

## 1.20.0 (July 20, 2020)

FEATURES:

- new resource: `keycloak_user_roles` ([#315](https://github.com/mrparkers/terraform-provider-keycloak/pull/315))
- new resource: `keycloak_identity_provider_token_exchange_scope_permission` ([#318](https://github.com/mrparkers/terraform-provider-keycloak/pull/318))
- new resources: `keycloak_saml_client_scope`, `keycloak_saml_client_default_scopes` ([#320](https://github.com/mrparkers/terraform-provider-keycloak/pull/320))

IMPROVEMENTS:

- add `default_signature_algorithm` attribute for `keycloak_realm` resource ([#282](https://github.com/mrparkers/terraform-provider-keycloak/pull/282))
- add `parent_id` attribute to `keycloak_custom_user_federation` resource ([#325](https://github.com/mrparkers/terraform-provider-keycloak/pull/325))
- add `extra_config` attribute to identity provider mapper resources ([#316](https://github.com/mrparkers/terraform-provider-keycloak/pull/316))
- add `include_in_token_scope` and `gui_order` attributes to `keycloak_openid_client_scope` resource ([#320](https://github.com/mrparkers/terraform-provider-keycloak/pull/320))
- add `base_path` provider attribute, improve login error messages ([#332](https://github.com/mrparkers/terraform-provider-keycloak/pull/332))
- add encryption attributes to `keycloak_saml_client` resource ([#342](https://github.com/mrparkers/terraform-provider-keycloak/pull/342))
- add `signature_algorithm` attribute to `keycloak_saml_client` resource ([#345](https://github.com/mrparkers/terraform-provider-keycloak/pull/345))

BUG FIXES:

- fix import for `keycloak_openid_client_service_account_role` resource ([#314](https://github.com/mrparkers/terraform-provider-keycloak/pull/314))
- fix realm role support for `keycloak_generic_client_role_mapper` resource ([#316](https://github.com/mrparkers/terraform-provider-keycloak/pull/316))
- fix `keycloak_group` data source to support nested groups ([#334](https://github.com/mrparkers/terraform-provider-keycloak/pull/334))
- fix `keycloak_group` data source / resource to support group names with backslash character ([#337](https://github.com/mrparkers/terraform-provider-keycloak/pull/337))

Huge thanks to all the individuals who have contributed towards this release:

- [@chanhht](https://github.com/chanhht)
- [@dmeyerholt](https://github.com/dmeyerholt)
- [@elmarx](https://github.com/elmarx)
- [@hcl31415](https://github.com/hcl31415)
- [@hnnsngl](https://github.com/hnnsngl)
- [@jgrgt](https://github.com/jgrgt)
- [@lathspell](https://github.com/lathspell)
- [@m-v-k](https://github.com/m-v-k)
- [@tomrutsaert](https://github.com/tomrutsaert)
- [@Useurmind](https://github.com/Useurmind)
- [@wadahiro](https://github.com/wadahiro)

## 1.19.0 (June 5, 2020)

FEATURES:

- new resource: `keycloak_openid_user_client_role_protocol_mapper` ([#299](https://github.com/mrparkers/terraform-provider-keycloak/pull/299))
- new resource: `keycloak_openid_user_session_note_protocol_mapper` ([#309](https://github.com/mrparkers/terraform-provider-keycloak/pull/309))

IMPROVEMENTS:

- add `login_theme` attribute to `keycloak_openid_client` resource ([#278](https://github.com/mrparkers/terraform-provider-keycloak/pull/278))
- add `aggregate_attributes` attribute to `keycloak_openid_user_attribute_protocol_mapper` resource ([#272](https://github.com/mrparkers/terraform-provider-keycloak/pull/272))
- add `user_managed_access` attribute to `keycloak_realm` resource ([#275](https://github.com/mrparkers/terraform-provider-keycloak/pull/275))
- support deployed JavaScript policies for `keycloak_openid_client_js_policy` resource ([#275](https://github.com/mrparkers/terraform-provider-keycloak/pull/275))
- add `internal_id` computed attribute to `keycloak_realm` resource and data source ([#270](https://github.com/mrparkers/terraform-provider-keycloak/pull/270))
- surface Keycloak API errors to users during `terraform plan` and `terraform apply` ([#304](https://github.com/mrparkers/terraform-provider-keycloak/pull/304))
- add `kerberos` configuration for `keycloak_ldap_user_federation` resource ([#290](https://github.com/mrparkers/terraform-provider-keycloak/pull/290))
- test all major versions of Keycloak in CI ([#294](https://github.com/mrparkers/terraform-provider-keycloak/pull/294))
- add import support for `keycloak_generic_client_role_mapper` resource ([#310](https://github.com/mrparkers/terraform-provider-keycloak/pull/310))
- use terraform-plugin-sdk user agent string in http client ([#311](https://github.com/mrparkers/terraform-provider-keycloak/pull/311))

BUG FIXES:

- fix: mark `group_id` attribute as required for `keycloak_group_roles` resource ([#292](https://github.com/mrparkers/terraform-provider-keycloak/pull/292))

Huge thanks to all the individuals who have contributed towards this release:

- [@alevit33](https://github.com/alevit33)
- [@arminfelder](https://github.com/arminfelder)
- [@awilliamsOM1](https://github.com/awilliamsOM1)
- [@dlechevalier](https://github.com/dlechevalier)
- [@dmeyerholt](https://github.com/dmeyerholt)
- [@elmarx](https://github.com/elmarx)
- [@hawknewton](https://github.com/hawknewton)
- [@javefang](https://github.com/javefang)
- [@jgrgt](https://github.com/jgrgt)
- [@pascal-hofmann](https://github.com/pascal-hofmann)
- [@tomrutsaert](https://github.com/tomrutsaert)
- [@Useurmind](https://github.com/Useurmind)
- [@wadahiro](https://github.com/wadahiro)

## 1.18.0 (April 17, 2020)

FEATURES:

* new resource: `keycloak_ldap_hardcoded_group_mapper` ([#264](https://github.com/mrparkers/terraform-provider-keycloak/pull/264))
* new data source: `keycloak_saml_client_installation_provider` ([#263](https://github.com/mrparkers/terraform-provider-keycloak/pull/263))
* new resource: `keycloak_ldap_role_mapper` ([#265](https://github.com/mrparkers/terraform-provider-keycloak/pull/265))

IMPROVEMENTS:

* add `tls_insecure_skip_verify` provider attribute ([#237](https://github.com/mrparkers/terraform-provider-keycloak/pull/237))
* add `client_scope_id` attribute to `keycloak_generic_client_role_mapper` resource ([#253](https://github.com/mrparkers/terraform-provider-keycloak/pull/253))
* add `email_verified` attribute to `keycloak_user` resource ([#256](https://github.com/mrparkers/terraform-provider-keycloak/pull/256))
* add `JSON` as a valid `claim_value_type` for openid protocol mapper resources ([#260](https://github.com/mrparkers/terraform-provider-keycloak/pull/260))
* add `force_name_id_format` attribute to `keycloak_saml_client` resource ([#261](https://github.com/mrparkers/terraform-provider-keycloak/pull/261))
* add `consent_required` and `authentication_flow_binding_overrides` attributes for `keycloak_openid_client` resource ([#262](https://github.com/mrparkers/terraform-provider-keycloak/pull/262))
* add `root_url` attribute to `keycloak_openid_client` resource ([#248](https://github.com/mrparkers/terraform-provider-keycloak/pull/248))
* add federated identity support for `keycloak_user` resource ([#274](https://github.com/mrparkers/terraform-provider-keycloak/pull/274))

BUG FIXES:

* correctly handle manually deleted clients when refreshing a `keycloak_openid_client_default_scopes` resource ([#252](https://github.com/mrparkers/terraform-provider-keycloak/pull/252))
* correctly handle manually deleted clients when refreshing a `keycloak_openid_client_optional_scopes` resource

Huge thanks to all the individuals who have contributed towards this release:

- [@dmeyerholt](https://github.com/dmeyerholt)
- [@dwaynebailey](https://github.com/dwaynebailey)
- [@Filirom1](https://github.com/Filirom1)
- [@languitar](https://github.com/languitar)
- [@lathspell](https://github.com/lathspell)
- [@tomrutsaert](https://github.com/tomrutsaert)
- @[Trois-Six](https://github.com/Trois-Six)
- [@Xide](https://github.com/Xide)

## 1.17.1 (March 12, 2020)

BUG FIXES:

* fix: allow `defaultScope` and `acceptsPromptNoneForwardFromClient` attributes to be set for `keycloak_oidc_identity_provider`. Previously, these attributes could only be set via `extra_config`, which stopped working as of v1.17.0. This release introduces these attributes as top-level attributes for the `keycloak_oidc_identity_provider` resource.

Huge thanks to all the individuals who have contributed towards this release:

- [@tomrutsaert](https://github.com/tomrutsaert)

## 1.17.0 (March 10, 2020)

FEATURES:

* new resources: `keycloak_authentication_flow`, `keycloak_authentication_subflow`, `keycloak_authentication_execution` ([#215](https://github.com/mrparkers/terraform-provider-keycloak/pull/215))
* new resource: `keycloak_authentication_execution_config` ([#241](https://github.com/mrparkers/terraform-provider-keycloak/pull/241))
* new resource: `keycloak_oidc_google_identity_provider` ([#240](https://github.com/mrparkers/terraform-provider-keycloak/pull/240))
* new resource: `keycloak_ldap_msad_user_account_control_mapper` ([#244](https://github.com/mrparkers/terraform-provider-keycloak/pull/244))
* new resources: `keycloak_openid_client_group_policy`, `keycloak_openid_client_role_policy`, `keycloak_openid_client_aggregate_policy`, `keycloak_openid_client_js_policy`, `keycloak_openid_client_time_policy`, `keycloak_openid_client_user_policy`, `keycloak_openid_client_client_policy` ([#246](https://github.com/mrparkers/terraform-provider-keycloak/pull/246))
* new resource: `keycloak_generic_client_role_mapper` ([#242](https://github.com/mrparkers/terraform-provider-keycloak/pull/242))

IMPROVEMENTS:

* add `client_scope_id` attribute to `keycloak_generic_client_protocol_mapper` resource ([#229](https://github.com/mrparkers/terraform-provider-keycloak/pull/229))
* add `root_ca_certificate` attribute to provider config ([#227](https://github.com/mrparkers/terraform-provider-keycloak/pull/227))
* add `scopes` attribute to `keycloak_openid_client_authorization_permission` resource ([#220](https://github.com/mrparkers/terraform-provider-keycloak/pull/220))
* add `access_token_lifespan` attribute to `keycloak_openid_client` resource ([#233](https://github.com/mrparkers/terraform-provider-keycloak/pull/233))

Huge thanks to all the individuals who have contributed towards this release:

- [@Amad27](https://github.com/Amad27)
- [@BernhardBerbuir](https://github.com/BernhardBerbuir)
- [@Guarionex](https://github.com/Guarionex)
- [@moritz31](https://github.com/moritz31)
- [@mukuru-shaun](https://github.com/mukuru-shaun)
- [@ndrpnt](https://github.com/ndrpnt)
- [@sanderginn](https://github.com/sanderginn)
- [@tomrutsaert](https://github.com/tomrutsaert)
- [@yspotts](https://github.com/yspotts)

## 1.16.0 (February 13, 2020)

FEATURES:

* new resource: `keycloak_realm_events` ([#211](https://github.com/mrparkers/terraform-provider-keycloak/pull/211))
* new resource: `resource_keycloak_openid_client_service_account_role` ([#202](https://github.com/mrparkers/terraform-provider-keycloak/pull/202))

IMPROVEMENTS:

* add base_url attribute to `keycloak_openid_client` resource ([#201](https://github.com/mrparkers/terraform-provider-keycloak/pull/201))
* allow configuration of the client timeout by an environment variable ([#206](https://github.com/mrparkers/terraform-provider-keycloak/pull/206))
* add consent_required attribute to `keycloak_openid_client` resource ([#207](https://github.com/mrparkers/terraform-provider-keycloak/pull/207))
* add admin_url attribute to `keycloak_openid_client` resource ([#203](https://github.com/mrparkers/terraform-provider-keycloak/pull/203))
* add display_name_html attribute to `keycloak_realm` resource and data source ([#209](https://github.com/mrparkers/terraform-provider-keycloak/pull/209))
* switch to terraform-plugin-sdk ([#214](https://github.com/mrparkers/terraform-provider-keycloak/pull/214))

BUG FIXES:

* URL encode role names to allow for special characters ([#213](https://github.com/mrparkers/terraform-provider-keycloak/pull/213))

Huge thanks to all the individuals who have contributed towards this release:

- [@bturbes](https://github.com/bturbes)
- [@cthiebault](https://github.com/cthiebault)
- [@LoicAG](https://github.com/LoicAG)
- [@SvenHamers](https://github.com/SvenHamers)
- [@waldemarschmalz](https://github.com/waldemarschmalz)

## 1.15.0 (January 20, 2020)

FEATURES:

* new resource: `keycloak_ldap_hardcoded_role_mapper` ([#195](https://github.com/mrparkers/terraform-provider-keycloak/pull/195))

IMPROVEMENTS:

* add `full_scope_allowed` attribute to `keycloak_openid_client` resource ([#193](https://github.com/mrparkers/terraform-provider-keycloak/pull/193))
* add `exclude_session_state_from_auth_response` attribute to `keycloak_openid_client` resource ([#191](https://github.com/mrparkers/terraform-provider-keycloak/pull/191))
* allow empty value for `pkce_code_challenge_method` attribute on `keycloak_openid_client` resource ([#198](https://github.com/mrparkers/terraform-provider-keycloak/pull/198))
* support attributes for `keycloak_group` resource ([#199](https://github.com/mrparkers/terraform-provider-keycloak/pull/199))

Huge thanks to all the individuals who have contributed towards this release:

- [@aromeyer](https://github.com/aromeyer)
- [@daviddesre](https://github.com/daviddesre)
- [@dlechevalier](https://github.com/dlechevalier)
- [@madddi](https://github.com/madddi)
- [@pths](https://github.com/pths)

## 1.14.0 (December 18, 2019)

FEATURES:

* add `keycloak_openid_client_service_account_user` data source ([#181](https://github.com/mrparkers/terraform-provider-keycloak/pull/181))
* add `keycloak_group` data source ([#185](https://github.com/mrparkers/terraform-provider-keycloak/pull/185))

IMPROVEMENTS:

* support Keycloak v8.0.0 ([#183](https://github.com/mrparkers/terraform-provider-keycloak/pull/183))
* new functionality for `keycloak_realm`: brute_force_detection, ssl_required, and custom attributes ([#183](https://github.com/mrparkers/terraform-provider-keycloak/pull/183))
* allow you to prevent refresh token reuse with a new `revoke_refresh_token` attribute for the `keycloak_realm` resource ([#183](https://github.com/mrparkers/terraform-provider-keycloak/pull/183))
  * **note: please refer to the [docs](https://mrparkers.github.io/terraform-provider-keycloak/resources/keycloak_realm/#tokens) for the new configuration values if you currently use `refresh_token_max_reuse`**

Huge thanks to all the individuals who have contributed towards this release:

- [@madddi](https://github.com/madddi)
- [@tomrutsaert](https://github.com/tomrutsaert)
- [@Trois-Six](https://github.com/Trois-Six)

## 1.13.2 (November 27, 2019)

BUG FIXES:

* fix: update Keycloak API call to handle groups with more than 100 members ([#179](https://github.com/mrparkers/terraform-provider-keycloak/pull/179))

Huge thanks to all the individuals who have contributed towards this release:

- [@Shubhammathur22](https://github.com/Shubhammathur22)
- [@vhiairrassary](https://github.com/vhiairrassary)

## 1.13.1 (November 19, 2019)

BUG FIXES:

* correctly handle Keycloak role names that contain a forward slash ([#175](https://github.com/mrparkers/terraform-provider-keycloak/pull/175))

Huge thanks to all the individuals who have contributed towards this release:

- [@Shubhammathur22](https://github.com/Shubhammathur22)

## 1.13.0 (November 8, 2019)

IMPROVEMENTS:

* use cookiejar for Keycloak API requests ([#173](https://github.com/mrparkers/terraform-provider-keycloak/pull/173))

Huge thanks to all the individuals who have contributed towards this release:

- [@alexashley](https://github.com/alexashley)

## 1.12.0 (November 4, 2019)

IMPROVEMENTS:

* add `pkce_code_challenge_method` attribute for `keycloak_openid_client` resource ([#170](https://github.com/mrparkers/terraform-provider-keycloak/pull/170))

BUG FIXES:

* always use valid client secret for `keycloak_oidc_identity_provider` resource ([#171](https://github.com/mrparkers/terraform-provider-keycloak/pull/171))
* fix state issues for `keycloak_openid_client_service_account_role` resource ([#171](https://github.com/mrparkers/terraform-provider-keycloak/pull/171))

Huge thanks to all the individuals who have contributed towards this release:

- [@AndrewChubatiuk](https://github.com/AndrewChubatiuk)
- [@PSanetra](https://github.com/PSanetra)

## 1.11.1 (October 17, 2019)

BUG FIXES:

* fix required attribute for `keycloak_realm` data source ([#166](https://github.com/mrparkers/terraform-provider-keycloak/pull/166))
* automatically retry role deletion if the first attempt fails ([#168](https://github.com/mrparkers/terraform-provider-keycloak/pull/168))

Huge thanks to all the individuals who have contributed towards this release:

- [@aroemers](https://github.com/aroemers)

## 1.11.0 (October 14, 2019)

FEATURES:

* new resource: `keycloak_openid_user_realm_role_protocol_mapper` ([#159](https://github.com/mrparkers/terraform-provider-keycloak/pull/159))
* new data source: `keycloak_realm` ([#160](https://github.com/mrparkers/terraform-provider-keycloak/pull/160))

IMPROVEMENTS:

* added `timeout` provider attribute ([#155](https://github.com/mrparkers/terraform-provider-keycloak/pull/155))
* always export `serviceAccountId` for `keycloak_openid_client` resource ([#162](https://github.com/mrparkers/terraform-provider-keycloak/pull/162))

BUG FIXES:

* fix default value for `reset_credentials_flow` attribute in `keycloak_realm` resource ([#158](https://github.com/mrparkers/terraform-provider-keycloak/pull/158))

Huge thanks to all the individuals who have contributed towards this release:

- [@darin-sai](https://github.com/darin-sai)
- [@drcrees](https://github.com/drcrees)
- [@LoicAG](https://github.com/LoicAG)
- [@marcoreni](https://github.com/marcoreni)
- [@Trois-Six](https://github.com/Trois-Six)
- [@xiaoyang-connyun](https://github.com/xiaoyang-connyun)

## 1.10.0 (September 6, 2019)

note: this release contains a [bug](https://github.com/mrparkers/terraform-provider-keycloak/issues/156) in the `keycloak_realm` resource that incorrectly sets the default attribute for `reset_credentials_flow` to `"registration"`. Please ensure that you set this attribute manually to override the incorrect default until a future release fixes this issue.

FEATURES:

* new resource: `keycloak_required_action` ([#131](https://github.com/mrparkers/terraform-provider-keycloak/pull/131))
* new resource: `keycloak_default_groups` ([#146](https://github.com/mrparkers/terraform-provider-keycloak/pull/146))
* new resources: `keycloak_role`, `keycloak_group_roles`, `keycloak_openid_hardcoded_role_protocol_mapper` ([#143](https://github.com/mrparkers/terraform-provider-keycloak/pull/143))
* new data source: `keycloak_role` ([#143](https://github.com/mrparkers/terraform-provider-keycloak/pull/143))

IMPROVEMENTS:

* add `security_defences` attribute to `keycloak_realm` resource ([#130](https://github.com/mrparkers/terraform-provider-keycloak/pull/130))
* support custom config for `keycloak_custom_user_federation` resource ([#134](https://github.com/mrparkers/terraform-provider-keycloak/pull/134))
* add `initial_login` provider attribute to optionally avoid requests during provider setup ([#136](https://github.com/mrparkers/terraform-provider-keycloak/pull/136))
* support custom config for `keycloak_oidc_identity_provider` resource ([#137](https://github.com/mrparkers/terraform-provider-keycloak/pull/137))
* add `password_policy` attribute for `keycloak_realm` resource ([#139](https://github.com/mrparkers/terraform-provider-keycloak/pull/139))
* add flow binding attributes for `keycloak_realm` resource ([#140](https://github.com/mrparkers/terraform-provider-keycloak/pull/140))

BUG FIXES:

* fix user attributes to handle attributes longer than 255 characters ([#132](https://github.com/mrparkers/terraform-provider-keycloak/pull/132))
* fix import for `keycloak_oidc_identity_provider` ([#142](https://github.com/mrparkers/terraform-provider-keycloak/pull/142))

Huge thanks to all the individuals who have contributed towards this release:

- [@AndrewChubatiuk](https://github.com/AndrewChubatiuk)
- [@BernhardBerbuir](https://github.com/BernhardBerbuir)
- [@camelpunch](https://github.com/camelpunch)
- [@fharding1](https://github.com/fharding1)
- [@ltscif](https://github.com/ltscif)
- [@nagytech](https://github.com/nagytech)
- [@tomrutsaert](https://github.com/tomrutsaert)

## 1.9.0 (June 20, 2019)

FEATURES:

* add `full_scope_allowed` attribute to `keycloak_saml_client` resource ([#118](https://github.com/mrparkers/terraform-provider-keycloak/pull/118))
* add `internationalization` attribute to `keycloak_realm` resource ([#124](https://github.com/mrparkers/terraform-provider-keycloak/pull/124))
* add `smtp_server` attribute to `keycloak_realm` resource ([#122](https://github.com/mrparkers/terraform-provider-keycloak/pull/122))

IMPROVEMENTS:

* allow the provider to use a confidential client with the password grant ([#114](https://github.com/mrparkers/terraform-provider-keycloak/pull/114))
* update Terraform SDK to 0.12.1 ([#120](https://github.com/mrparkers/terraform-provider-keycloak/pull/120))
* bump dependency versions for custom user federation example ([#121](https://github.com/mrparkers/terraform-provider-keycloak/pull/121))
* add static binary to release for use within Alpine Docker iamges ([#129](https://github.com/mrparkers/terraform-provider-keycloak/pull/129))

Huge thanks to all the individuals who have contributed towards this release:

- [@alexashley](https://github.com/alexashley)
- [@BernhardBerbuir](https://github.com/BernhardBerbuir)
- [@Floby](https://github.com/Floby)
- [@tomrutsaert](https://github.com/tomrutsaert)

## 1.8.0 (May 14, 2019)

FEATURES:

* new resources: `keycloak_openid_client_authorization_resource`, `keycloak_openid_client_authorization_scope`, `keycloak_openid_client_authorization_permission`, `keycloak_openid_client_service_account_role` ([#104](https://github.com/mrparkers/terraform-provider-keycloak/pull/104))
  - note: docs for these resources will be released at a later date. for now, please refer to the source files.
* new data sources: keycloak_openid_client, keycloak_openid_client_authorization_policy ([#104](https://github.com/mrparkers/terraform-provider-keycloak/pull/104))
  - note: docs for these data sources will be released at a later date. for now, please refer to the source files.

IMPROVEMENTS:

* chore: update provider SDK to 0.12 ([#107](https://github.com/mrparkers/terraform-provider-keycloak/pull/107))
* chore: support Keycloak v6.0.1 ([#106](https://github.com/mrparkers/terraform-provider-keycloak/pull/106))
* chore: renames provider resource/data files ([#105](https://github.com/mrparkers/terraform-provider-keycloak/pull/105))

Huge thanks to all the individuals who have contributed towards this release:

- [@AndrewChubatiuk](https://github.com/AndrewChubatiuk)
- [@ctrox](https://github.com/ctrox)

## 1.7.0 (April 18, 2019)

FEATURES:

* new resources: `keycloak_identity_provider` and mappers ([#92](https://github.com/mrparkers/terraform-provider-keycloak/pull/92))
  - note: docs for these resources will be released at a later date. for now, please refer to the source files.

IMPROVEMENTS:

* new attributes added for `keycloak_saml_client` resource ([#103](https://github.com/mrparkers/terraform-provider-keycloak/pull/103))

Huge thanks to all the individuals who have contributed towards this release:

- [@AndrewChubatiuk](https://github.com/AndrewChubatiuk)
- [@ctrox](https://github.com/ctrox)

## 1.6.0 (March 6, 2019)

FEATURES:

* new resource: `keycloak_openid_client_optional_scopes` ([#96](https://github.com/mrparkers/terraform-provider-keycloak/pull/96))
* new resource: `keycloak_openid_audience_protocol_mapper` ([#97](https://github.com/mrparkers/terraform-provider-keycloak/pull/97))

## 1.5.0 (February 22, 2019)

FEATURES:

* add support for non-master realms and resource owner password grant for Keycloak authentication ([#88](https://github.com/mrparkers/terraform-provider-keycloak/pull/88))

IMPROVEMENTS:

* support Keycloak v4.8.3.Final and Terraform v0.11.11 ([#93](https://github.com/mrparkers/terraform-provider-keycloak/pull/93))

BUG FIXES:

* handle 404 errors when reading a group for group memberships ([#95](https://github.com/mrparkers/terraform-provider-keycloak/pull/95))

Huge thanks to all the individuals who have contributed towards this release:

- [@AndrewChubatiuk](https://github.com/AndrewChubatiuk)

## 1.4.0 (January 28, 2019)

FEATURES:

* new resource: `keycloak_saml_user_property_protocol_mapper` ([#85](https://github.com/mrparkers/terraform-provider-keycloak/pull/85))

## 1.3.0 (January 25, 2019)

FEATURES:

* new resource: `keycloak_saml_user_attribute_protocol_mapper` ([#84](https://github.com/mrparkers/terraform-provider-keycloak/pull/84))

## 1.2.0 (January 24, 2019)

FEATURES:

* new resource: `keycloak_saml_client` ([#82](https://github.com/mrparkers/terraform-provider-keycloak/pull/82))

IMPROVEMENTS:

* add validation for usernames to ensure they are always lowercase ([#83](https://github.com/mrparkers/terraform-provider-keycloak/pull/83))

## 1.1.0 (January 7, 2019)

IMPROVEMENTS:

* openid_client: add web_origins attribute ([#81](https://github.com/mrparkers/terraform-provider-keycloak/pull/81))
* user: add initial_password attribute ([#77](https://github.com/mrparkers/terraform-provider-keycloak/pull/77))

BUG FIXES:

* ldap mappers: don't assume component fields are returned by Keycloak API ([#80](https://github.com/mrparkers/terraform-provider-keycloak/pull/80))

Huge thanks to all the individuals who have contributed towards this release:

- [@Floby](https://github.com/Floby)

## 1.0.0 (December 16, 2018)

Initial Release!

Docs: https://mrparkers.github.io/terraform-provider-keycloak
