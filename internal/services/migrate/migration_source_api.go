// Package migrate holds the Terraform resources of the VPSie Migrate service.
package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/vpsie/govpsie"
)

const sourcesPath = "/apps/v2/migrate/sources"

// reservedKeys are the body fields the resource owns; a connector setting or
// credential cannot use them.
var reservedKeys = map[string]bool{
	"type": true, "name": true, "datacenterIdentifier": true, "requestIp": true,
}

// migrationSource is a source as the API presents it. It never carries a
// credential.
type migrationSource struct {
	Identifier           string `json:"identifier"`
	Type                 string `json:"type"`
	Name                 string `json:"name"`
	Host                 string `json:"host"`
	Username             string `json:"username"`
	DatacenterIdentifier string `json:"datacenterIdentifier"`
	Status               string `json:"status"`
	CreatedAt            string `json:"createdAt"`
}

type sourceEnvelope struct {
	Data migrationSource `json:"data"`
}

type sourceListEnvelope struct {
	Data struct {
		Items []migrationSource `json:"items"`
	} `json:"data"`
}

// sourceAPI talks to /migrate/sources through the shared govpsie client, which
// has no Migrate service of its own.
type sourceAPI struct {
	client *govpsie.Client
}

// createBody merges the resource attributes, the connector settings and the
// credentials into the one flat body the API validates per connector type.
func createBody(typ, name, datacenter string, settings, credentials map[string]string) (map[string]string, error) {
	body := map[string]string{"type": typ, "name": name, "datacenterIdentifier": datacenter}
	for _, fields := range []map[string]string{settings, credentials} {
		for k, v := range fields {
			if reservedKeys[k] {
				return nil, fmt.Errorf("%q is set by the resource itself and cannot be used as a setting or credential", k)
			}
			if _, dup := body[k]; dup {
				return nil, fmt.Errorf("%q is given twice: keep each field in either settings or credentials", k)
			}
			body[k] = v
		}
	}
	return body, nil
}

func (a *sourceAPI) create(ctx context.Context, body map[string]string) (*migrationSource, error) {
	req, err := a.client.NewRequest(ctx, http.MethodPost, sourcesPath, body)
	if err != nil {
		return nil, err
	}
	var out sourceEnvelope
	if err := a.client.Do(ctx, req, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// get returns nil when the source no longer exists. The API answers a missing
// source with a 404 whose body the client cannot tell from other failures, so a
// failed read is confirmed against the list before it counts as gone.
func (a *sourceAPI) get(ctx context.Context, identifier string) (*migrationSource, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, sourcesPath+"/"+url.PathEscape(identifier), nil)
	if err != nil {
		return nil, err
	}
	var out sourceEnvelope
	getErr := a.client.Do(ctx, req, &out)
	if getErr == nil {
		return &out.Data, nil
	}
	listed, err := a.listed(ctx, identifier)
	if err != nil {
		return nil, getErr
	}
	if listed {
		return nil, getErr
	}
	return nil, nil
}

func (a *sourceAPI) listed(ctx context.Context, identifier string) (bool, error) {
	const limit = 100
	for offset := 0; ; offset += limit {
		req, err := a.client.NewRequest(ctx, http.MethodGet,
			fmt.Sprintf("%s?limit=%d&offset=%d", sourcesPath, limit, offset), nil)
		if err != nil {
			return false, err
		}
		var out sourceListEnvelope
		if err := a.client.Do(ctx, req, &out); err != nil {
			return false, err
		}
		for _, item := range out.Data.Items {
			if item.Identifier == identifier {
				return true, nil
			}
		}
		if len(out.Data.Items) < limit {
			return false, nil
		}
	}
}

func (a *sourceAPI) rotateCredentials(ctx context.Context, identifier string, credentials map[string]string) error {
	req, err := a.client.NewRequest(ctx, http.MethodPut, sourcesPath+"/"+url.PathEscape(identifier)+"/credentials", credentials)
	if err != nil {
		return err
	}
	return a.client.Do(ctx, req, nil)
}

func (a *sourceAPI) delete(ctx context.Context, identifier string) error {
	req, err := a.client.NewRequest(ctx, http.MethodDelete, sourcesPath+"/"+url.PathEscape(identifier), nil)
	if err != nil {
		return err
	}
	return a.client.Do(ctx, req, nil)
}

// sortedKeys lists a map's keys in order, for stable messages.
func sortedKeys(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
