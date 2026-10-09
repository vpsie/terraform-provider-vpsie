package accesstoken

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// The API generates the token value and returns it once; it refuses a value
// chosen by the client. access_token is therefore computed, never configured.
func TestAccessTokenIsServerGenerated(t *testing.T) {
	resp := &resource.SchemaResponse{}
	NewAccessTokenResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	attr, ok := resp.Schema.Attributes["access_token"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("access_token attribute missing or of an unexpected type")
	}
	if attr.Required || attr.Optional {
		t.Errorf("access_token must not be configurable: the API refuses a client-chosen value")
	}
	if !attr.Computed || !attr.Sensitive {
		t.Errorf("access_token must be computed and sensitive")
	}
}
