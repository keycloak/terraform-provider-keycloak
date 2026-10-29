package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Client scopes (openid and saml) control the consent screen with two attributes:
//   - display_on_consent_screen ("display.on.consent.screen")
//   - consent_screen_text ("consent.screen.text")
//
// When display_on_consent_screen is not set in the configuration, its value is derived from
// consent_screen_text (set -> true, unset -> false). This keeps the behaviour of earlier
// provider versions, which did not expose display_on_consent_screen at all.

// clientScopeDisplayOnConsentScreenDefault returns the value of display_on_consent_screen
// when it is not set in the configuration.
func clientScopeDisplayOnConsentScreenDefault(consentScreenText string) bool {
	return consentScreenText != ""
}

// getClientScopeDisplayOnConsentScreen returns the display_on_consent_screen value to send to Keycloak.
func getClientScopeDisplayOnConsentScreen(data *schema.ResourceData) (bool, error) {
	rawDisplayOnConsentScreen, diags := data.GetRawConfigAt(cty.GetAttrPath("display_on_consent_screen"))
	if diags.HasError() {
		return false, fmt.Errorf("error reading display_on_consent_screen: %s", diags[0].Summary)
	}

	if !rawDisplayOnConsentScreen.IsNull() {
		return data.Get("display_on_consent_screen").(bool), nil
	}

	return clientScopeDisplayOnConsentScreenDefault(data.Get("consent_screen_text").(string)), nil
}

// customizeDiffClientScopeDisplayOnConsentScreen plans the derived value of display_on_consent_screen
// when it is not set in the configuration, so that changes to it are visible in the plan.
func customizeDiffClientScopeDisplayOnConsentScreen(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return nil
	}

	if !rawConfig.GetAttr("display_on_consent_screen").IsNull() {
		return nil
	}

	if !d.NewValueKnown("consent_screen_text") {
		return d.SetNewComputed("display_on_consent_screen")
	}

	return d.SetNew("display_on_consent_screen", clientScopeDisplayOnConsentScreenDefault(d.Get("consent_screen_text").(string)))
}
