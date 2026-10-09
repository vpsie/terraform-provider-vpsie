package loadbalancer

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

func TestCorsHeadersAttribute(t *testing.T) {
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
	attr, ok := domain.NestedObject.Attributes["cors_headers"].(schema.BoolAttribute)
	if !ok {
		t.Fatalf("cors_headers attribute missing or of an unexpected type")
	}
	// Optional only: a Computed attribute that an older API (prod NY) never returns stays unknown in
	// every plan, and an unknown domain attribute replaces every HTTP rule on each apply.
	if !attr.Optional || attr.Computed {
		t.Errorf("cors_headers must be Optional and not Computed")
	}
	if len(attr.PlanModifiers) != 0 {
		t.Errorf("cors_headers needs no plan modifier, got %d", len(attr.PlanModifiers))
	}
	if attr.Default != nil {
		t.Errorf("cors_headers must have no default: an unset attribute is not sent")
	}
}

func TestDomainsToAPISendsCorsHeadersOnlyWhenSet(t *testing.T) {
	cases := []struct {
		name  string
		value types.Bool
		want  *bool
	}{
		{"unset (null) is not sent", types.BoolNull(), nil},
		{"unknown is not sent", types.BoolUnknown(), nil},
		{"false is sent", types.BoolValue(false), boolPtr(false)},
		{"true is sent", types.BoolValue(true), boolPtr(true)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domainsToAPI([]lbDomainModel{{CorsHeaders: tc.value}})[0].CorsHeaders
			if (got == nil) != (tc.want == nil) || (got != nil && *got != *tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCorsHeadersFromAPI(t *testing.T) {
	on := govpsie.FlexBool(true)
	off := govpsie.FlexBool(false)
	cases := []struct {
		name    string
		current types.Bool
		api     *govpsie.FlexBool
		want    types.Bool
	}{
		{"not set in config: stays null whatever the API returns", types.BoolNull(), &on, types.BoolNull()},
		{"not set in config, older API without the field: stays null", types.BoolNull(), nil, types.BoolNull()},
		{"set in config: the API value is read back (drift shows in the plan)", types.BoolValue(true), &off, types.BoolValue(false)},
		{"set in config, older API without the field: the configured value is kept", types.BoolValue(false), nil, types.BoolValue(false)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := corsHeadersFromAPI(tc.current, tc.api); !got.Equal(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func boolPtr(v bool) *bool { return &v }
