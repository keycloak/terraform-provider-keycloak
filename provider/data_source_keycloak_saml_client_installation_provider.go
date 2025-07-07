package provider

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakSamlClientInstallationProvider() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakSamlClientInstallationProviderRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zip_files": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceKeycloakSamlClientInstallationProviderRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	providerId := data.Get("provider_id").(string)

	value, err := keycloakClient.GetSamlClientInstallationProvider(ctx, realmId, clientId, providerId)
	if err != nil {
		return diag.FromErr(err)
	}

	h := sha1.New()
	h.Write(value)
	id := base64.URLEncoding.EncodeToString(h.Sum(nil))

	data.SetId(id)
	data.Set("realm_id", realmId)
	data.Set("client_id", clientId)
	data.Set("provider_id", providerId)
	data.Set("value", string(value))

	zipFiles, err := readZipFiles(value)
	if err != nil {
		return diag.FromErr(err)
	}
	data.Set("zip_files", zipFiles)

	return nil
}

func readZipFiles(content []byte) (map[string]string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		if errors.Is(err, zip.ErrFormat) {
			return nil, nil
		}

		return nil, fmt.Errorf("error reading zip files: %w", err)
	}

	files := make(map[string]string, len(zipReader.File))
	for _, file := range zipReader.File {
		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("error opening zip file for reading: %w", err)
		}
		fileContent, err := io.ReadAll(fileReader)
		if err != nil {
			return nil, fmt.Errorf("error reading zip file content: %w", err)
		}
		files[file.FileInfo().Name()] = string(fileContent)

		err = fileReader.Close()
		if err != nil {
			return nil, fmt.Errorf("error closing zip file content: %w", err)
		}
	}

	return files, nil
}
