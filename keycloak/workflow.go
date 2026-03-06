package keycloak

import (
	"context"
	"fmt"
)

type WorkflowStep struct {
	Id     string              `json:"id,omitempty"`
	Uses   string              `json:"uses"`
	After  string              `json:"after,omitempty"`
	Config map[string][]string `json:"config,omitempty"`
}

type Workflow struct {
	Id                string         `json:"id,omitempty"`
	Realm             string         `json:"-"`
	Name              string         `json:"name"`
	On                string         `json:"on"`
	Enabled           bool           `json:"enabled"`
	Conditions        string         `json:"conditions,omitempty"`
	Steps             []WorkflowStep `json:"steps"`
	CancelInProgress  string         `json:"cancelInProgress,omitempty"`
	RestartInProgress string         `json:"restartInProgress,omitempty"`
}

func (keycloakClient *KeycloakClient) NewWorkflow(ctx context.Context, workflow *Workflow) error {
	_, location, err := keycloakClient.post(ctx, fmt.Sprintf("/realms/%s/workflows", workflow.Realm), workflow)
	if err != nil {
		return err
	}
	workflow.Id = getIdFromLocationHeader(location)

	return nil
}

func (keycloakClient *KeycloakClient) GetWorkflow(ctx context.Context, realm, id string) (*Workflow, error) {
	var workflow Workflow

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/workflows/%s", realm, id), &workflow, nil)
	if err != nil {
		return nil, err
	}

	workflow.Realm = realm

	return &workflow, nil
}

func (keycloakClient *KeycloakClient) GetWorkflowByName(ctx context.Context, realm, name string) (*Workflow, error) {
	var workflows []Workflow

	err := keycloakClient.get(ctx, fmt.Sprintf("/realms/%s/workflows", realm), &workflows, nil)
	if err != nil {
		return nil, err
	}

	for _, wf := range workflows {
		if wf.Name == name {
			return keycloakClient.GetWorkflow(ctx, realm, wf.Id)
		}
	}

	return nil, fmt.Errorf("workflow with name %s not found in realm %s", name, realm)
}

func (keycloakClient *KeycloakClient) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {
	return keycloakClient.put(ctx, fmt.Sprintf("/realms/%s/workflows/%s", workflow.Realm, workflow.Id), workflow)
}

func (keycloakClient *KeycloakClient) DeleteWorkflow(ctx context.Context, realm, id string) error {
	return keycloakClient.delete(ctx, fmt.Sprintf("/realms/%s/workflows/%s", realm, id), nil)
}
