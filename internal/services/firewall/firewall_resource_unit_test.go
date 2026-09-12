package firewall

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

func stringList(values ...string) types.List {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elems)
}

func TestBuildRuleReqs(t *testing.T) {
	ctx := context.Background()
	rules := []firewallRuleModel{
		{
			Action:  types.StringValue("ACCEPT"),
			Type:    types.StringValue("in"),
			Proto:   types.StringValue("tcp"),
			Macro:   types.StringValue("SSH"),
			Source:  stringList("10.0.0.0/8"),
			Comment: types.StringValue("ssh"),
			Enable:  types.Int64Null(), // exercise the default
		},
		{
			Action: types.StringValue("DROP"),
			Type:   types.StringValue("out"),
			Dport:  types.StringValue("53"),
			Enable: types.Int64Value(0),
		},
	}

	reqs := buildRuleReqs(ctx, rules)
	if len(reqs) != 2 {
		t.Fatalf("expected 2 reqs, got %d", len(reqs))
	}
	if reqs[0].Action != "ACCEPT" || reqs[0].Type != "in" || reqs[0].Macro != "SSH" {
		t.Errorf("unexpected first req: %+v", reqs[0])
	}
	if len(reqs[0].Source) != 1 || reqs[0].Source[0] != "10.0.0.0/8" {
		t.Errorf("unexpected source: %+v", reqs[0].Source)
	}
	if reqs[0].Enable != 1 {
		t.Errorf("expected enable to default to 1, got %d", reqs[0].Enable)
	}
	if reqs[0].Identifier != "" {
		t.Errorf("identifier must be omitted so reqs compare by content, got %q", reqs[0].Identifier)
	}
	if reqs[1].Enable != 0 {
		t.Errorf("expected explicit enable 0, got %d", reqs[1].Enable)
	}
	if reqs[0].Comment != "ssh" {
		t.Errorf("explicit comment should be preserved, got %q", reqs[0].Comment)
	}
	if reqs[1].Comment != "Custom" {
		t.Errorf("an omitted comment must become \"Custom\" (the API rejects empty), got %q", reqs[1].Comment)
	}
}

func TestAssignRuleIdentifiers(t *testing.T) {
	rules := []firewallRuleModel{
		{Type: types.StringValue("in")},
		{Type: types.StringValue("out")},
		{Type: types.StringValue("in")},
	}
	detail := &govpsie.FirewallGroupDetailData{
		Rules: govpsie.FirewallRules{
			InBound: []govpsie.InBoundFirewallRules{
				{Identifier: "in-1"},
				{Identifier: "in-2"},
			},
			OutBound: []govpsie.OutBoundFirewallRules{
				{Identifier: "out-1"},
			},
		},
	}

	got := assignRuleIdentifiers(rules, detail)
	if got[0].Identifier.ValueString() != "in-1" {
		t.Errorf("rule 0: got %q want in-1", got[0].Identifier.ValueString())
	}
	if got[1].Identifier.ValueString() != "out-1" {
		t.Errorf("rule 1: got %q want out-1", got[1].Identifier.ValueString())
	}
	if got[2].Identifier.ValueString() != "in-2" {
		t.Errorf("rule 2: got %q want in-2", got[2].Identifier.ValueString())
	}
}

func TestOptionalStringAndList(t *testing.T) {
	if !optionalString("").IsNull() {
		t.Error("empty string should map to null")
	}
	if optionalString("x").ValueString() != "x" {
		t.Error("non-empty string should be preserved")
	}
	if !optionalStringList(context.Background(), nil).IsNull() {
		t.Error("nil slice should map to a null list")
	}
	list := optionalStringList(context.Background(), []string{"1.1.1.1"})
	if list.IsNull() || len(list.Elements()) != 1 {
		t.Errorf("non-empty slice should map to a one-element list, got %+v", list)
	}
}
