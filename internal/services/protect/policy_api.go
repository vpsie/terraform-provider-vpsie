// Package protect holds the Terraform resources of VPSie Backup and Protect.
package protect

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

const policiesPath = "/apps/v2/protect/policies"

type namedRef struct {
	Identifier *string `json:"identifier"`
	Name       *string `json:"name"`
}

type retention struct {
	KeepRecentHours int64 `json:"keepRecentHours"`
	KeepDaily       int64 `json:"keepDaily"`
	KeepWeekly      int64 `json:"keepWeekly"`
}

// policy is a protection policy as the API presents it.
type policy struct {
	Identifier       string    `json:"identifier"`
	Name             string    `json:"name"`
	Tier             string    `json:"tier"`
	State            string    `json:"state"`
	Datacenter       namedRef  `json:"datacenter"`
	Plan             namedRef  `json:"plan"`
	RPOMinutes       int64     `json:"rpoMinutes"`
	Retention        retention `json:"retention"`
	ImmutabilityDays int64     `json:"immutabilityDays"`
	Consistency      string    `json:"consistency"`
	VerifyMode       string    `json:"verifyMode"`
	RTOTargetMinutes *int64    `json:"rtoTargetMinutes"`
	Version          int64     `json:"version"`
	CreatedAt        string    `json:"createdAt"`
}

type policyEnvelope struct {
	Data policy `json:"data"`
}

type policyListEnvelope struct {
	Data struct {
		Items []policy `json:"items"`
	} `json:"data"`
}

// createRequest is the POST body. Optional fields the user left unset are
// omitted so the platform applies its own defaults.
type createRequest struct {
	Name             string    `json:"name"`
	Tier             string    `json:"tier"`
	Datacenter       string    `json:"datacenter"`
	Plan             string    `json:"plan,omitempty"`
	RPOMinutes       int64     `json:"rpoMinutes"`
	Retention        retention `json:"retention"`
	ImmutabilityDays *int64    `json:"immutabilityDays,omitempty"`
	Consistency      string    `json:"consistency,omitempty"`
	VerifyMode       string    `json:"verifyMode,omitempty"`
	RTOTargetMinutes *int64    `json:"rtoTargetMinutes,omitempty"`
}

// editRequest is the PATCH body: the version read and every editable field.
// rtoTargetMinutes is always sent so that removing it from the configuration
// clears it.
type editRequest struct {
	Version          int64     `json:"version"`
	Name             string    `json:"name"`
	RPOMinutes       int64     `json:"rpoMinutes"`
	Retention        retention `json:"retention"`
	ImmutabilityDays *int64    `json:"immutabilityDays,omitempty"`
	Consistency      string    `json:"consistency,omitempty"`
	VerifyMode       string    `json:"verifyMode,omitempty"`
	RTOTargetMinutes *int64    `json:"rtoTargetMinutes"`
}

type policyAPI struct {
	api *rawapi.API
}

func (p *policyAPI) create(ctx context.Context, body createRequest) (*policy, error) {
	var out policyEnvelope
	if err := p.api.Call(ctx, rawapi.Post, policiesPath, body, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

func (p *policyAPI) edit(ctx context.Context, identifier string, body editRequest) (*policy, error) {
	var out policyEnvelope
	if err := p.api.Call(ctx, rawapi.Patch, policiesPath+"/"+rawapi.Escape(identifier), body, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// get returns nil when the policy no longer exists. The client cannot tell a
// 404 from other failures, so a failed read is confirmed against the list.
func (p *policyAPI) get(ctx context.Context, identifier string) (*policy, error) {
	var out policyEnvelope
	getErr := p.api.Call(ctx, rawapi.Get, policiesPath+"/"+rawapi.Escape(identifier), nil, &out)
	if getErr == nil {
		if out.Data.State == "removing" {
			return nil, nil
		}
		return &out.Data, nil
	}
	all, err := p.list(ctx, "")
	if err != nil {
		return nil, getErr
	}
	for i := range all {
		if all[i].Identifier == identifier {
			return nil, getErr
		}
	}
	return nil, nil
}

// list reads every policy, optionally only those in one state.
func (p *policyAPI) list(ctx context.Context, state string) ([]policy, error) {
	path := policiesPath
	if state != "" {
		path += "?state=" + url.QueryEscape(state)
	}
	return rawapi.Paged(ctx, path, func(ctx context.Context, paged string) ([]policy, error) {
		var out policyListEnvelope
		if err := p.api.Call(ctx, rawapi.Get, paged, nil, &out); err != nil {
			return nil, err
		}
		return out.Data.Items, nil
	})
}

// remove deletes an empty policy; the platform wants its name typed back.
func (p *policyAPI) remove(ctx context.Context, identifier, name string) error {
	body := map[string]string{"confirmName": name}
	if err := p.api.Call(ctx, rawapi.Delete, policiesPath+"/"+rawapi.Escape(identifier), body, nil); err != nil {
		return fmt.Errorf("%w (a policy with protected servers cannot be removed until they are removed from it)", err)
	}
	return nil
}
