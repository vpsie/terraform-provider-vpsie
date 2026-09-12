package monitoring

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &monitoringRuleResource{}
	_ resource.ResourceWithConfigure   = &monitoringRuleResource{}
	_ resource.ResourceWithImportState = &monitoringRuleResource{}
)

type monitoringRuleResource struct {
	client *govpsie.Client
}

type monitoringRuleResourceModel struct {
	Identifier types.String            `tfsdk:"identifier"`
	RuleName   types.String            `tfsdk:"rule_name"`
	Status     types.String            `tfsdk:"status"`
	Frequency  types.String            `tfsdk:"frequency"`
	CreatedOn  types.String            `tfsdk:"created_on"`
	Vms        types.List              `tfsdk:"vms"`
	Rules      []monitoringMetricModel `tfsdk:"rule"`
}

type monitoringMetricModel struct {
	MetricType    types.String            `tfsdk:"metric_type"`
	Condition     types.String            `tfsdk:"condition"`
	Threshold     types.String            `tfsdk:"threshold"`
	ThresholdType types.String            `tfsdk:"threshold_type"`
	Period        types.String            `tfsdk:"period"`
	Status        types.String            `tfsdk:"status"`
	Actions       []monitoringActionModel `tfsdk:"action"`
}

type monitoringActionModel struct {
	ActionName types.String `tfsdk:"action_name"`
	ActionKey  types.String `tfsdk:"action_key"`
	Email      types.String `tfsdk:"email"`
	Value      types.String `tfsdk:"value"`
}

func NewMonitoringRuleResource() resource.Resource {
	return &monitoringRuleResource{}
}

func (m *monitoringRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_rule"
}

func (m *monitoringRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VPSie monitoring rule: one or more metric conditions, each with " +
			"alert actions, optionally attached to VMs. Changing the metrics or their actions replaces " +
			"the rule; `status` and `vms` are updated in place.\n\n" +
			"The API rejects some rule/action combinations (surfaced as an error on apply): the same " +
			"`metric_type`+`condition` may not repeat across rule blocks, a single rule may not combine " +
			"`power_off` and `reboot` actions, and actions within a rule must be distinct.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"rule_name": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"frequency": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "How often (in minutes) the rule is evaluated.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1"),
				MarkdownDescription: "Rule status: `1` (enabled) or `0` (disabled).",
				Validators:          []validator.String{stringvalidator.OneOf("0", "1")},
			},
			"vms": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Identifiers of VMs the rule applies to.",
			},
			"created_on": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				MarkdownDescription: "A metric condition to evaluate. Changing any rule replaces the resource.",
				PlanModifiers:       []planmodifier.List{listplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"metric_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "One of cpu, ram, disk_percentage, disk_read, disk_write, net_in, net_out.",
							Validators: []validator.String{stringvalidator.OneOf(
								"cpu", "ram", "disk_percentage", "disk_read", "disk_write", "net_in", "net_out",
							)},
						},
						"condition": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "`greater_than` or `less_than`.",
							Validators:          []validator.String{stringvalidator.OneOf("greater_than", "less_than")},
						},
						"threshold": schema.StringAttribute{
							Required: true,
						},
						"threshold_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "`percentage` or `mbps`.",
							Validators:          []validator.String{stringvalidator.OneOf("percentage", "mbps")},
						},
						"period": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Sustained period, in minutes, before the rule triggers.",
						},
						"status": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("1"),
							MarkdownDescription: "Metric status: `1` (enabled) or `0` (disabled); defaults to `1`. Set it explicitly when importing a disabled metric, otherwise the default plans a replacement.",
						},
					},
					Blocks: map[string]schema.Block{
						"action": schema.ListNestedBlock{
							MarkdownDescription: "An action to take when the condition is met.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"action_name": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "`send_alert`, `power_controls`, or `auto_scale`.",
										Validators:          []validator.String{stringvalidator.OneOf("send_alert", "power_controls", "auto_scale")},
									},
									"action_key": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "`send_alert`, `reboot`, `power_off`, `cpu`, `ram`, or `ssd`.",
										Validators:          []validator.String{stringvalidator.OneOf("send_alert", "reboot", "power_off", "cpu", "ram", "ssd")},
									},
									"email": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Alert recipient email. Required: the API rejects an action with no email.",
									},
									"value": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: "Adjustment value for auto_scale actions (e.g. `+2`).",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (m *monitoringRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	m.client = client
}

func (m *monitoringRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan monitoringRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vms []string
	if !plan.Vms.IsNull() && !plan.Vms.IsUnknown() {
		resp.Diagnostics.Append(plan.Vms.ElementsAs(ctx, &vms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := &govpsie.CreateMonitoringRuleReq{
		RuleName:  plan.RuleName.ValueString(),
		Status:    plan.Status.ValueString(),
		Frequency: plan.Frequency.ValueString(),
		Rules:     buildMetricReqs(plan.Rules),
		Vms:       vms,
	}

	if err := m.client.Monitoring.CreateRule(ctx, createReq); err != nil {
		resp.Diagnostics.AddError("Error creating monitoring rule", err.Error())
		return
	}

	rule, err := m.getRuleByName(ctx, plan.RuleName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading monitoring rule after creation", err.Error())
		return
	}

	plan.Identifier = types.StringValue(rule.Identifier)
	plan.CreatedOn = types.StringValue(rule.CreatedOn)
	plan.Status = types.StringValue(strconv.Itoa(rule.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (m *monitoringRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state monitoringRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := m.client.Monitoring.GetMonitoringRule(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading monitoring rule", err.Error())
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.RuleName = types.StringValue(rule.RuleName)
	state.Status = types.StringValue(strconv.Itoa(rule.Status))
	state.Frequency = types.StringValue(strconv.Itoa(rule.Frequency))
	state.CreatedOn = types.StringValue(rule.CreatedOn)

	// The metric/action blocks are config-owned and replace-on-change; keep the
	// values the user wrote to avoid string/int normalization drift. On import
	// (no prior rules in state) populate them from the API instead.
	if len(state.Rules) == 0 {
		state.Rules = metricsFromAPI(rule.Metrics)
	}

	// Refresh the VM set only when it is tracked or non-empty (avoids null->[] drift).
	if len(rule.Vms) > 0 || !state.Vms.IsNull() {
		vmIdentifiers := make([]string, 0, len(rule.Vms))
		for _, vm := range rule.Vms {
			vmIdentifiers = append(vmIdentifiers, vm.Identifier)
		}
		vmsList, diags := types.ListValueFrom(ctx, types.StringType, vmIdentifiers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Vms = vmsList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (m *monitoringRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state monitoringRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()

	if !plan.Status.Equal(state.Status) {
		if err := m.client.Monitoring.ToggleMonitoringRuleStatus(ctx, plan.Status.ValueString(), identifier); err != nil {
			resp.Diagnostics.AddError("Error updating monitoring rule status", err.Error())
			return
		}
	}

	var planVms, stateVms []string
	if !plan.Vms.IsNull() && !plan.Vms.IsUnknown() {
		resp.Diagnostics.Append(plan.Vms.ElementsAs(ctx, &planVms, false)...)
	}
	if !state.Vms.IsNull() && !state.Vms.IsUnknown() {
		resp.Diagnostics.Append(state.Vms.ElementsAs(ctx, &stateVms, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	toAttach := stringsDifference(planVms, stateVms)
	toDetach := stringsDifference(stateVms, planVms)
	if len(toAttach) > 0 {
		if err := m.client.Monitoring.AttachVms(ctx, identifier, toAttach); err != nil {
			resp.Diagnostics.AddError("Error attaching VMs to monitoring rule", err.Error())
			return
		}
	}
	if len(toDetach) > 0 {
		if err := m.client.Monitoring.DetachVms(ctx, identifier, toDetach); err != nil {
			resp.Diagnostics.AddError("Error detaching VMs from monitoring rule", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (m *monitoringRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state monitoringRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := m.client.Monitoring.DeleteMonitoringRule(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting monitoring rule",
			"couldn't delete monitoring rule, unexpected error: "+err.Error(),
		)
	}
}

func (m *monitoringRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// getRuleByName recovers a freshly-created rule by name, since the create
// endpoint returns no identifier. rule_name is RequiresReplace so it is stable
// within a Terraform configuration; duplicate names created out of band would
// be ambiguous.
func (m *monitoringRuleResource) getRuleByName(ctx context.Context, name string) (*govpsie.MonitoringRule, error) {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
		rules, err := m.client.Monitoring.ListMonitoringRule(ctx, nil)
		if err != nil {
			lastErr = err
			continue
		}
		for i := range rules {
			if rules[i].RuleName == name {
				return &rules[i], nil
			}
		}
		lastErr = fmt.Errorf("monitoring rule %q not found after creation", name)
	}
	return nil, lastErr
}

// buildMetricReqs converts the model metric blocks into create request metrics.
func buildMetricReqs(metrics []monitoringMetricModel) []govpsie.CreateMonitoringMetricReq {
	out := make([]govpsie.CreateMonitoringMetricReq, 0, len(metrics))
	for _, metric := range metrics {
		status := metric.Status.ValueString()
		if status == "" {
			status = "1"
		}
		actions := make([]govpsie.CreateMonitoringActionReq, 0, len(metric.Actions))
		for _, a := range metric.Actions {
			actions = append(actions, govpsie.CreateMonitoringActionReq{
				ActionName: a.ActionName.ValueString(),
				ActionKey:  a.ActionKey.ValueString(),
				Email:      a.Email.ValueString(),
				Value:      a.Value.ValueString(),
			})
		}
		out = append(out, govpsie.CreateMonitoringMetricReq{
			MetricType:    metric.MetricType.ValueString(),
			Condition:     metric.Condition.ValueString(),
			Threshold:     metric.Threshold.ValueString(),
			ThresholdType: metric.ThresholdType.ValueString(),
			Period:        metric.Period.ValueString(),
			Status:        status,
			Actions:       actions,
		})
	}
	return out
}

// metricsFromAPI converts API metrics into model blocks, used on import.
func metricsFromAPI(metrics []govpsie.MonitoringMetric) []monitoringMetricModel {
	out := make([]monitoringMetricModel, 0, len(metrics))
	for _, metric := range metrics {
		model := monitoringMetricModel{
			MetricType:    types.StringValue(metric.MetricType),
			Condition:     types.StringValue(metric.Condition),
			Threshold:     types.StringValue(strconv.Itoa(metric.Threshold)),
			ThresholdType: types.StringValue(metric.ThresholdType),
			Period:        types.StringValue(strconv.Itoa(metric.Period)),
			Status:        types.StringValue(strconv.Itoa(metric.Status)),
		}
		for _, a := range metric.Actions {
			action := monitoringActionModel{
				ActionName: types.StringValue(a.ActionName),
				ActionKey:  types.StringValue(a.ActionKey),
				Email:      types.StringValue(a.Email),
			}
			// value is optional; represent an absent value as null.
			if a.Value != "" {
				action.Value = types.StringValue(a.Value)
			} else {
				action.Value = types.StringNull()
			}
			model.Actions = append(model.Actions, action)
		}
		out = append(out, model)
	}
	return out
}

func stringsDifference(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	var diff []string
	for _, v := range a {
		if _, ok := set[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}
