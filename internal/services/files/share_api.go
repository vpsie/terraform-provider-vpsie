// Package files holds the Terraform resources of VPSie Files (NFS file shares).
package files

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

const sharesPath = "/apps/v2/files/shares"

// Poll settings of the asynchronous share changes; tests shorten them.
var (
	pollInterval = 5 * time.Second
	waitTimeout  = 30 * time.Minute
)

type accessEntry struct {
	VMIdentifier string `json:"vmIdentifier,omitempty"`
	CIDR         string `json:"cidr,omitempty"`
	Mode         string `json:"mode,omitempty"`
}

type gateway struct {
	Identifier string `json:"identifier"`
	VPCIP      string `json:"vpcIp"`
	State      string `json:"state"`
}

type mount struct {
	NFS4 string `json:"nfs4"`
}

// share is a file share as the API presents it to a customer. The VPC id is
// not part of it.
type share struct {
	Identifier   string        `json:"identifier"`
	Name         string        `json:"name"`
	DcIdentifier string        `json:"dcIdentifier"`
	SizeGb       int64         `json:"sizeGb"`
	ResizingToGb *int64        `json:"resizingToGb"`
	UsedBytes    int64         `json:"usedBytes"`
	State        string        `json:"state"`
	ErrorCode    string        `json:"errorCode"`
	CreatedAt    string        `json:"createdAt"`
	Gateway      *gateway      `json:"gateway"`
	Access       []accessEntry `json:"access"`
	Mount        *mount        `json:"mount"`
}

type shareEnvelope struct {
	Data share `json:"data"`
}

type shareListEnvelope struct {
	Data []share `json:"data"`
}

type createRequest struct {
	Name         string        `json:"name"`
	DcIdentifier string        `json:"dcIdentifier"`
	VPCID        int64         `json:"vpcId"`
	SizeGb       int64         `json:"sizeGb"`
	Access       []accessEntry `json:"access"`
}

type shareAPI struct {
	api *rawapi.API
}

func (s *shareAPI) create(ctx context.Context, body createRequest) (*share, error) {
	var out shareEnvelope
	if err := s.api.Call(ctx, rawapi.Post, sharesPath, body, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// get returns nil when the share no longer exists (a missing share or one in
// state deleted). The client cannot tell a 404 from other failures, so a
// failed read is confirmed against the list.
func (s *shareAPI) get(ctx context.Context, identifier string) (*share, error) {
	var out shareEnvelope
	getErr := s.api.Call(ctx, rawapi.Get, sharesPath+"/"+rawapi.Escape(identifier), nil, &out)
	if getErr == nil {
		if out.Data.State == "deleted" {
			return nil, nil
		}
		return &out.Data, nil
	}
	all, err := s.list(ctx, "")
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

func (s *shareAPI) list(ctx context.Context, dcIdentifier string) ([]share, error) {
	path := sharesPath
	if dcIdentifier != "" {
		path += "?dcIdentifier=" + url.QueryEscape(dcIdentifier)
	}
	return rawapi.Paged(ctx, path, func(ctx context.Context, paged string) ([]share, error) {
		var out shareListEnvelope
		if err := s.api.Call(ctx, rawapi.Get, paged, nil, &out); err != nil {
			return nil, err
		}
		return out.Data, nil
	})
}

func (s *shareAPI) resize(ctx context.Context, identifier string, sizeGb int64) error {
	return s.api.Call(ctx, rawapi.Patch, sharesPath+"/"+rawapi.Escape(identifier), map[string]int64{"sizeGb": sizeGb}, nil)
}

func (s *shareAPI) setAccess(ctx context.Context, identifier string, access []accessEntry) error {
	return s.api.Call(ctx, rawapi.Put, sharesPath+"/"+rawapi.Escape(identifier)+"/access", map[string][]accessEntry{"access": access}, nil)
}

func (s *shareAPI) remove(ctx context.Context, identifier, name string) error {
	body := map[string]any{"confirmName": name, "force": false}
	return s.api.Call(ctx, rawapi.Delete, sharesPath+"/"+rawapi.Escape(identifier), body, nil)
}

// waitReady polls until the share is ready. A share in state error ends the
// wait with its error code; a share that disappears ends it too.
func (s *shareAPI) waitReady(ctx context.Context, identifier string) (*share, error) {
	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	for {
		sh, err := s.get(ctx, identifier)
		if err != nil {
			return nil, err
		}
		switch {
		case sh == nil:
			return nil, fmt.Errorf("file share %s disappeared while it was changing", identifier)
		case sh.State == "ready":
			return sh, nil
		case sh.State == "error":
			return nil, fmt.Errorf("file share %s is in state error (%s)", identifier, sh.ErrorCode)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("file share %s did not become ready: %w", identifier, ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}

// waitGone polls until the share is deleted.
func (s *shareAPI) waitGone(ctx context.Context, identifier string) error {
	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	for {
		sh, err := s.get(ctx, identifier)
		if err != nil {
			return err
		}
		if sh == nil {
			return nil
		}
		if sh.State == "error" {
			return fmt.Errorf("file share %s could not be deleted (%s)", identifier, sh.ErrorCode)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("file share %s was not deleted: %w", identifier, ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}
