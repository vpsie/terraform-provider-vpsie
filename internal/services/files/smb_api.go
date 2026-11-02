package files

import (
	"context"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

const smbUsersPath = "/apps/v2/files/smb-users"

// smbUser is an SMB user as the API presents it to a customer: never a
// password or a hash.
type smbUser struct {
	Identifier        string `json:"identifier"`
	Username          string `json:"username"`
	DcIdentifier      string `json:"dcIdentifier"`
	ShareCount        *int64 `json:"shareCount"`
	PasswordChangedAt string `json:"passwordChangedAt"`
	CreatedAt         string `json:"createdAt"`
}

type smbUserEnvelope struct {
	Data smbUser `json:"data"`
}

type smbUserListEnvelope struct {
	Data []smbUser `json:"data"`
}

// smbUserRequest creates a user in a VPC, named by datacenter and VPC id or as
// the VPC of a share. The two forms exclude each other.
type smbUserRequest struct {
	DcIdentifier    string `json:"dcIdentifier,omitempty"`
	VPCID           int64  `json:"vpcId,omitempty"`
	ShareIdentifier string `json:"shareIdentifier,omitempty"`
	Username        string `json:"username"`
	Password        string `json:"password"`
}

// smbGrant is one SMB user a share grants.
type smbGrant struct {
	Identifier string `json:"identifier"`
	Mode       string `json:"mode"`
}

type smbShareRequest struct {
	Enabled bool       `json:"enabled"`
	Users   []smbGrant `json:"users"`
}

func (s *shareAPI) createSmbUser(ctx context.Context, body smbUserRequest) (*smbUser, error) {
	var out smbUserEnvelope
	if err := s.api.Call(ctx, rawapi.Post, smbUsersPath, body, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

func (s *shareAPI) listSmbUsers(ctx context.Context) ([]smbUser, error) {
	var out smbUserListEnvelope
	if err := s.api.Call(ctx, rawapi.Get, smbUsersPath, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// getSmbUser returns nil when the account no longer has the user. The API has
// no read by identifier, so the account's list is searched.
func (s *shareAPI) getSmbUser(ctx context.Context, identifier string) (*smbUser, error) {
	var out smbUserListEnvelope
	if err := s.api.Call(ctx, rawapi.Get, smbUsersPath, nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Data {
		if out.Data[i].Identifier == identifier {
			return &out.Data[i], nil
		}
	}
	return nil, nil
}

func (s *shareAPI) setSmbPassword(ctx context.Context, identifier, password string) error {
	return s.api.Call(ctx, rawapi.Put, smbUsersPath+"/"+rawapi.Escape(identifier)+"/password", map[string]string{"password": password}, nil)
}

func (s *shareAPI) removeSmbUser(ctx context.Context, identifier string) error {
	return s.api.Call(ctx, rawapi.Delete, smbUsersPath+"/"+rawapi.Escape(identifier), nil, nil)
}

func (s *shareAPI) setShareSmb(ctx context.Context, identifier string, body smbShareRequest) error {
	return s.api.Call(ctx, rawapi.Put, sharesPath+"/"+rawapi.Escape(identifier)+"/smb", body, nil)
}
