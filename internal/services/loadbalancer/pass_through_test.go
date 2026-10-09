package loadbalancer

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// passThroughAttribute digs the domain-level pass_through attribute out of the resource schema.
func passThroughAttribute(t *testing.T) schema.BoolAttribute {
	t.Helper()
	resp := &resource.SchemaResponse{}
	NewLoadbalancerResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	rule, ok := resp.Schema.Blocks["rule"].(schema.ListNestedBlock)
	if !ok {
		t.Fatalf("rule block missing or of an unexpected type")
	}
	domain, ok := rule.NestedObject.Blocks["domain"].(schema.ListNestedBlock)
	if !ok {
		t.Fatalf("domain block missing or of an unexpected type")
	}
	attr, ok := domain.NestedObject.Attributes["pass_through"].(schema.BoolAttribute)
	if !ok {
		t.Fatalf("pass_through attribute missing or of an unexpected type")
	}
	return attr
}

func TestPassThroughIsDeprecated(t *testing.T) {
	attr := passThroughAttribute(t)
	if !strings.Contains(attr.DeprecationMessage, "backend_scheme") {
		t.Errorf("deprecation message should point to backend_scheme, got %q", attr.DeprecationMessage)
	}
	if !strings.Contains(attr.MarkdownDescription, "Deprecated") {
		t.Errorf("description should say the attribute is deprecated, got %q", attr.MarkdownDescription)
	}
}

func TestPassThroughMustBeFalse(t *testing.T) {
	attr := passThroughAttribute(t)
	if len(attr.Validators) == 0 {
		t.Fatalf("pass_through has no validator")
	}
	cases := []struct {
		name    string
		value   types.Bool
		wantErr bool
	}{
		{"true is refused (the API answers 400)", types.BoolValue(true), true},
		{"false is accepted", types.BoolValue(false), false},
		{"null is accepted", types.BoolNull(), false},
		{"unknown is accepted", types.BoolUnknown(), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.BoolResponse{}
			for _, v := range attr.Validators {
				v.ValidateBool(context.Background(), validator.BoolRequest{
					Path:        path.Root("pass_through"),
					ConfigValue: tc.value,
				}, resp)
			}
			if resp.Diagnostics.HasError() != tc.wantErr {
				t.Errorf("HasError() = %v, want %v (%v)", resp.Diagnostics.HasError(), tc.wantErr, resp.Diagnostics)
			}
		})
	}
}

func TestDomainsToAPINeverSendsPassThrough(t *testing.T) {
	domains := domainsToAPI([]lbDomainModel{{PassThrough: types.BoolValue(true)}})
	if len(domains) != 1 || domains[0].PassThrough {
		t.Errorf("passThrough must always be sent as false, got %+v", domains)
	}
}
