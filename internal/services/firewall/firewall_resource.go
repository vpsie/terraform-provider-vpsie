package firewall

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &firewallResource{}
	_ resource.ResourceWithConfigure   = &firewallResource{}
	_ resource.ResourceWithImportState = &firewallResource{}
)

type firewallResource struct {
	client *govpsie.Client
}

type firewallResourceModel struct {
	ID            types.Int64         `tfsdk:"id"`
	Identifier    types.String        `tfsdk:"identifier"`
	GroupName     types.String        `tfsdk:"group_name"`
	UserName      types.String        `tfsdk:"user_name"`
	CreatedOn     types.String        `tfsdk:"created_on"`
	UpdatedOn     types.String        `tfsdk:"updated_on"`
	CreatedBy     types.Int64         `tfsdk:"created_by"`
	InboundCount  types.Int64         `tfsdk:"inbound_count"`
	OutboundCount types.Int64         `tfsdk:"outbound_count"`
	VmsCount      types.Int64         `tfsdk:"vms_count"`
	Rules         []firewallRuleModel `tfsdk:"rule"`
}

type firewallRuleModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Action     types.String `tfsdk:"action"`
	Type       types.String `tfsdk:"type"`
	Proto      types.String `tfsdk:"proto"`
	Source     types.List   `tfsdk:"source"`
	Dest       types.List   `tfsdk:"dest"`
	Dport      types.String `tfsdk:"dport"`
	Sport      types.String `tfsdk:"sport"`
	Comment    types.String `tfsdk:"comment"`
	Enable     types.Int64  `tfsdk:"enable"`
	Macro      types.String `tfsdk:"macro"`
	Log        types.String `tfsdk:"log"`
}

func NewFirewallResource() resource.Resource {
	return &firewallResource{}
}

func (f *firewallResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (f *firewallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VPSie firewall (security) group and its rules. Attach the group " +
			"to servers with `vpsie_firewall_attachment`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"identifier": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"group_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the firewall group (5-50 characters).",
				Validators:          []validator.String{stringvalidator.LengthBetween(5, 50)},
			},
			"user_name": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_on": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_on": schema.StringAttribute{
				Computed: true,
			},
			"created_by": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"inbound_count": schema.Int64Attribute{
				Computed: true,
			},
			"outbound_count": schema.Int64Attribute{
				Computed: true,
			},
			"vms_count": schema.Int64Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				MarkdownDescription: "A firewall rule. Rules are applied in the order given.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"action": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Rule action: `ACCEPT` or `DROP`.",
							Validators:          []validator.String{stringvalidator.OneOf("ACCEPT", "DROP")},
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Rule direction: `in` (inbound) or `out` (outbound).",
							Validators:          []validator.String{stringvalidator.OneOf("in", "out")},
						},
						"proto": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Protocol: `tcp`, `udp`, or `icmp`.",
							Validators:          []validator.String{stringvalidator.OneOf("tcp", "udp", "icmp")},
						},
						"source": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Source IPs or CIDRs.",
						},
						"dest": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Destination IPs or CIDRs.",
						},
						"dport": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Destination port or range (e.g. `80` or `8000:8100`). Ignored when `macro` is set.",
						},
						"sport": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Source port or range. Ignored when `macro` is set.",
						},
						"comment": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "An optional label. Within a single direction (`type`), comments must be unique; omit it to let the provider use a shared default.",
						},
						"enable": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(1),
							MarkdownDescription: "Whether the rule is enabled (`1`) or disabled (`0`). Defaults to `1`.",
						},
						"macro": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "A predefined Proxmox macro (e.g. `SSH`, `HTTP`). Setting it clears `dport`/`sport`.",
						},
						"log": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Log level: one of emerge, alert, crit, err, warning, notice, info, debug, nolog.",
						},
					},
				},
			},
		},
	}
}

func (f *firewallResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	f.client = client
}

func (f *firewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rulesReq := buildRuleReqs(ctx, plan.Rules)

	if err := f.client.FirewallGroup.Create(ctx, plan.GroupName.ValueString(), rulesReq); err != nil {
		resp.Diagnostics.AddError("Error creating firewall", "couldn't create firewall, unexpected error: "+err.Error())
		return
	}

	group, err := f.getGroupByName(ctx, plan.GroupName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall after creation", err.Error())
		return
	}

	detail, err := f.client.FirewallGroup.Get(ctx, group.Identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall after creation", err.Error())
		return
	}

	plan.Rules = assignRuleIdentifiers(plan.Rules, detail)
	applyGroupComputed(&plan, detail)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (f *firewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := f.client.FirewallGroup.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall", "couldn't read firewall, unexpected error: "+err.Error())
		return
	}

	// Compare the set of rule identifiers rather than their positions: the API
	// does not guarantee rule ordering, so per-rule identifier pairing is only
	// best-effort. When the identifier set is unchanged the managed rule set is
	// intact, so keep the config-owned values the user wrote (the API normalizes
	// CIDRs, clears ports when a macro is set, etc.). When the set differs, a
	// rule was added or removed out of band, so rebuild from the API to surface
	// the drift.
	apiIDs := map[string]struct{}{}
	for _, in := range detail.Rules.InBound {
		apiIDs[in.Identifier] = struct{}{}
	}
	for _, out := range detail.Rules.OutBound {
		apiIDs[out.Identifier] = struct{}{}
	}

	stateIDs := map[string]struct{}{}
	for _, rule := range state.Rules {
		if id := rule.Identifier.ValueString(); id != "" {
			stateIDs[id] = struct{}{}
		}
	}

	if !sameIDSet(apiIDs, stateIDs) {
		rebuilt := make([]firewallRuleModel, 0, len(detail.Rules.InBound)+len(detail.Rules.OutBound))
		for _, in := range detail.Rules.InBound {
			rebuilt = append(rebuilt, apiRuleToModel(ctx, firewallAPIRuleFromIn(in)))
		}
		for _, out := range detail.Rules.OutBound {
			rebuilt = append(rebuilt, apiRuleToModel(ctx, firewallAPIRuleFromOut(out)))
		}
		if len(rebuilt) == 0 {
			state.Rules = nil
		} else {
			state.Rules = rebuilt
		}
	}
	applyGroupComputed(&state, detail)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (f *firewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state firewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()

	if plan.GroupName.ValueString() != state.GroupName.ValueString() {
		if err := f.client.FirewallGroup.RenameGroup(ctx, identifier, plan.GroupName.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error renaming firewall group", err.Error())
			return
		}
	}

	planReqs := buildRuleReqs(ctx, plan.Rules)
	stateReqs := buildRuleReqs(ctx, state.Rules)

	// Rules have no stable user-facing key, so reconcile the whole set: drop the
	// existing rules and re-add the desired ones. This keeps the group (and its
	// VM attachments) intact, unlike recreating the whole resource.
	if !reflect.DeepEqual(planReqs, stateReqs) {
		var existingIDs []string
		for _, rule := range state.Rules {
			if id := rule.Identifier.ValueString(); id != "" {
				existingIDs = append(existingIDs, id)
			}
		}
		if len(existingIDs) > 0 {
			if err := f.client.FirewallGroup.DeleteRules(ctx, identifier, existingIDs); err != nil {
				resp.Diagnostics.AddError("Error updating firewall rules", "couldn't remove old rules: "+err.Error())
				return
			}
		}
		if len(planReqs) > 0 {
			if err := f.client.FirewallGroup.UpdateRules(ctx, identifier, planReqs); err != nil {
				resp.Diagnostics.AddError("Error updating firewall rules", "couldn't add rules: "+err.Error())
				return
			}
		}
	}

	detail, err := f.client.FirewallGroup.Get(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall after update", err.Error())
		return
	}

	plan.Rules = assignRuleIdentifiers(plan.Rules, detail)
	applyGroupComputed(&plan, detail)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (f *firewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := f.client.FirewallGroup.Delete(ctx, state.Identifier.ValueString(), "Deleted by Terraform", "")
	if err != nil {
		resp.Diagnostics.AddError("Error deleting firewall", "couldn't delete firewall, unexpected error: "+err.Error())
	}
}

func (f *firewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// getGroupByName finds a firewall group by name, retrying briefly because the
// create endpoint returns no identifier and the group can lag the list.
func (f *firewallResource) getGroupByName(ctx context.Context, name string) (*govpsie.FirewallGroupListData, error) {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
		groups, err := f.client.FirewallGroup.List(ctx, nil)
		if err != nil {
			lastErr = err
			continue
		}
		for i := range groups {
			if groups[i].GroupName == name {
				return &groups[i], nil
			}
		}
		lastErr = fmt.Errorf("firewall group %q not found after creation", name)
	}
	return nil, lastErr
}

// applyGroupComputed copies the server-computed group fields into the model.
func applyGroupComputed(m *firewallResourceModel, detail *govpsie.FirewallGroupDetailData) {
	m.ID = types.Int64Value(detail.Group.ID)
	m.Identifier = types.StringValue(detail.Group.Identifier)
	m.GroupName = types.StringValue(detail.Group.GroupName)
	m.UserName = types.StringValue(detail.Group.UserName)
	m.CreatedOn = types.StringValue(detail.Group.CreatedOn)
	m.UpdatedOn = types.StringValue(detail.Group.UpdatedOn)
	m.CreatedBy = types.Int64Value(detail.Group.CreatedBy)
	m.InboundCount = types.Int64Value(detail.Group.InboundCount)
	m.OutboundCount = types.Int64Value(detail.Group.OutboundCount)
	m.VmsCount = types.Int64Value(detail.Group.Vms)
}

// assignRuleIdentifiers stamps each plan rule with an identifier the API
// assigned, matching by (type, position). The backend does not contractually
// guarantee rule ordering, so an individual pairing may be off; that is benign
// here because identifiers are only ever used as a set — Update deletes every
// stored identifier before re-adding, and Read compares the identifier set (see
// sameIDSet), never a single rule's id.
func assignRuleIdentifiers(rules []firewallRuleModel, detail *govpsie.FirewallGroupDetailData) []firewallRuleModel {
	inIdx, outIdx := 0, 0
	for i := range rules {
		switch rules[i].Type.ValueString() {
		case "in":
			if inIdx < len(detail.Rules.InBound) {
				rules[i].Identifier = types.StringValue(detail.Rules.InBound[inIdx].Identifier)
				inIdx++
			}
		case "out":
			if outIdx < len(detail.Rules.OutBound) {
				rules[i].Identifier = types.StringValue(detail.Rules.OutBound[outIdx].Identifier)
				outIdx++
			}
		}
	}
	return rules
}

// buildRuleReqs converts the model rules into SDK request objects. The
// identifier is intentionally omitted so the resulting slices compare equal
// regardless of server-assigned ids.
func buildRuleReqs(ctx context.Context, rules []firewallRuleModel) []govpsie.FirewallUpdateReq {
	out := make([]govpsie.FirewallUpdateReq, 0, len(rules))
	for _, rule := range rules {
		var source, dest []string
		if !rule.Source.IsNull() && !rule.Source.IsUnknown() {
			_ = rule.Source.ElementsAs(ctx, &source, false)
		}
		if !rule.Dest.IsNull() && !rule.Dest.IsUnknown() {
			_ = rule.Dest.ElementsAs(ctx, &dest, false)
		}
		enable := int64(1)
		if !rule.Enable.IsNull() && !rule.Enable.IsUnknown() {
			enable = rule.Enable.ValueInt64()
		}
		// The API rejects a rule with no comment and treats a comment as a
		// per-direction unique key unless it is the literal "Custom"; send
		// "Custom" when the user omits a comment so uncommented rules work and
		// several may coexist in the same direction.
		comment := rule.Comment.ValueString()
		if comment == "" {
			comment = "Custom"
		}
		out = append(out, govpsie.FirewallUpdateReq{
			Action:  rule.Action.ValueString(),
			Type:    rule.Type.ValueString(),
			Proto:   rule.Proto.ValueString(),
			Source:  source,
			Dest:    dest,
			Dport:   rule.Dport.ValueString(),
			Sport:   rule.Sport.ValueString(),
			Comment: comment,
			Enable:  enable,
			Macro:   rule.Macro.ValueString(),
			Log:     rule.Log.ValueString(),
		})
	}
	return out
}

// firewallAPIRule is the common subset of the SDK in/out rule structs.
type firewallAPIRule struct {
	Identifier string
	Action     string
	Type       string
	Proto      string
	Source     []string
	Dest       []string
	Dport      string
	Sport      string
	Comment    string
	Enable     int64
	Macro      string
	Log        string
}

func firewallAPIRuleFromIn(in govpsie.InBoundFirewallRules) firewallAPIRule {
	return firewallAPIRule{
		Identifier: in.Identifier, Action: in.Action, Type: in.Type, Proto: in.Proto,
		Source: in.Source, Dest: in.Dest, Dport: in.Dport, Sport: in.Sport,
		Comment: in.Comment, Enable: in.Enable, Macro: in.Macro, Log: in.Log,
	}
}

func firewallAPIRuleFromOut(out govpsie.OutBoundFirewallRules) firewallAPIRule {
	return firewallAPIRule{
		Identifier: out.Identifier, Action: out.Action, Type: out.Type, Proto: out.Proto,
		Source: out.Source, Dest: out.Dest, Dport: out.Dport, Sport: out.Sport,
		Comment: out.Comment, Enable: out.Enable, Macro: out.Macro, Log: out.Log,
	}
}

// apiRuleToModel converts a live API rule into a model rule for out-of-band
// rules the state does not yet track.
func apiRuleToModel(ctx context.Context, r firewallAPIRule) firewallRuleModel {
	source := optionalStringList(ctx, r.Source)
	dest := optionalStringList(ctx, r.Dest)
	return firewallRuleModel{
		Identifier: types.StringValue(r.Identifier),
		Action:     types.StringValue(r.Action),
		Type:       types.StringValue(r.Type),
		Proto:      optionalString(r.Proto),
		Source:     source,
		Dest:       dest,
		Dport:      optionalString(r.Dport),
		Sport:      optionalString(r.Sport),
		Comment:    optionalString(r.Comment),
		Enable:     types.Int64Value(r.Enable),
		Macro:      optionalString(r.Macro),
		Log:        optionalString(r.Log),
	}
}

func sameIDSet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for id := range a {
		if _, ok := b[id]; !ok {
			return false
		}
	}
	return true
}

func optionalString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func optionalStringList(ctx context.Context, values []string) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, values)
	return list
}
