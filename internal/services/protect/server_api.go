package protect

import (
	"context"
	"errors"
	"fmt"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

const serversPath = "/apps/v2/protect/vms"

// protectedServer is a protected server as the API presents it.
type protectedServer struct {
	Identifier string   `json:"identifier"`
	Name       string   `json:"name"`
	Tier       string   `json:"tier"`
	State      string   `json:"state"`
	PausedBy   *string  `json:"pausedBy"`
	Policy     namedRef `json:"policy"`
	Source     namedRef `json:"source"`
	CreatedAt  string   `json:"createdAt"`
}

type serverEnvelope struct {
	Data protectedServer `json:"data"`
}

type serverListEnvelope struct {
	Data struct {
		Items []protectedServer `json:"items"`
	} `json:"data"`
}

type vmRef struct {
	Ref string `json:"ref"`
}

type addRequest struct {
	Policy string  `json:"policy"`
	Source string  `json:"source"`
	VMs    []vmRef `json:"vms"`
}

type serverAPI struct {
	api *rawapi.API
}

// add protects one server of a connected source under a policy.
func (s *serverAPI) add(ctx context.Context, policy, source, ref string) (*protectedServer, error) {
	var out serverListEnvelope
	body := addRequest{Policy: policy, Source: source, VMs: []vmRef{{Ref: ref}}}
	if err := s.api.Call(ctx, rawapi.Post, serversPath, body, &out); err != nil {
		return nil, err
	}
	if len(out.Data.Items) != 1 {
		return nil, fmt.Errorf("the platform answered %d protected servers for one added server", len(out.Data.Items))
	}
	return &out.Data.Items[0], nil
}

// gone reports whether a server no longer protects anything: it is being
// removed or already removed.
func (p *protectedServer) gone() bool {
	return p.State == "removing" || p.State == "removed"
}

// get returns nil when the protected server no longer exists. The client cannot
// tell a 404 from other failures, so a failed read is confirmed against the list.
func (s *serverAPI) get(ctx context.Context, identifier string) (*protectedServer, error) {
	var out serverEnvelope
	getErr := s.api.Call(ctx, rawapi.Get, serversPath+"/"+rawapi.Escape(identifier), nil, &out)
	if getErr == nil {
		if out.Data.gone() {
			return nil, nil
		}
		return &out.Data, nil
	}
	all, err := s.list(ctx)
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

func (s *serverAPI) list(ctx context.Context) ([]protectedServer, error) {
	return rawapi.Paged(ctx, serversPath, func(ctx context.Context, paged string) ([]protectedServer, error) {
		var out serverListEnvelope
		if err := s.api.Call(ctx, rawapi.Get, paged, nil, &out); err != nil {
			return nil, err
		}
		return out.Data.Items, nil
	})
}

// setPaused pauses or resumes the protection. A pause the platform made itself
// cannot be ended by the tenant, which the platform refuses.
func (s *serverAPI) setPaused(ctx context.Context, identifier string, paused bool) (*protectedServer, error) {
	action := "resume"
	if paused {
		action = "pause"
	}
	var out serverEnvelope
	if err := s.api.Call(ctx, rawapi.Post, serversPath+"/"+rawapi.Escape(identifier)+"/"+action, map[string]string{}, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// remove ends the protection; the platform wants the server name typed back and
// keeps the recovery points until their locks end.
func (s *serverAPI) remove(ctx context.Context, identifier, name string) error {
	if name == "" {
		return errors.New("the server name is missing from the state")
	}
	body := map[string]string{"confirmName": name}
	return s.api.Call(ctx, rawapi.Delete, serversPath+"/"+rawapi.Escape(identifier), body, nil)
}
