package monitoring

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

type monitoringRuleDataSource struct {
	client *govpsie.Client
}

type monitoringRuleDataSourceModel struct {
	ID    types.String           `tfsdk:"id"`
	Rules []monitoringRuleDSItem `tfsdk:"rules"`
}

type monitoringRuleDSItem struct {
	Identifier types.String         `tfsdk:"identifier"`
	RuleName   types.String         `tfsdk:"rule_name"`
	Status     types.Int64          `tfsdk:"status"`
	Frequency  types.Int64          `tfsdk:"frequency"`
	CreatedOn  types.String         `tfsdk:"created_on"`
	CreatedBy  types.String         `tfsdk:"created_by"`
	Metrics    []monitoringMetricDS `tfsdk:"metrics"`
}

type monitoringMetricDS struct {
	MetricType    types.String         `tfsdk:"metric_type"`
	Condition     types.String         `tfsdk:"condition"`
	Threshold     types.Int64          `tfsdk:"threshold"`
	ThresholdType types.String         `tfsdk:"threshold_type"`
	Period        types.Int64          `tfsdk:"period"`
	Status        types.Int64          `tfsdk:"status"`
	Actions       []monitoringActionDS `tfsdk:"actions"`
}

type monitoringActionDS struct {
	ActionName types.String `tfsdk:"action_name"`
	ActionKey  types.String `tfsdk:"action_key"`
	Email      types.String `tfsdk:"email"`
	Value      types.String `tfsdk:"value"`
}

func NewMonitoringRuleDataSource() datasource.DataSource {
	return &monitoringRuleDataSource{}
}

func (d *monitoringRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_rules"
}

func (d *monitoringRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{Computed: true},
						"rule_name":  schema.StringAttribute{Computed: true},
						"status":     schema.Int64Attribute{Computed: true},
						"frequency":  schema.Int64Attribute{Computed: true},
						"created_on": schema.StringAttribute{Computed: true},
						"created_by": schema.StringAttribute{Computed: true},
						"metrics": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"metric_type":    schema.StringAttribute{Computed: true},
									"condition":      schema.StringAttribute{Computed: true},
									"threshold":      schema.Int64Attribute{Computed: true},
									"threshold_type": schema.StringAttribute{Computed: true},
									"period":         schema.Int64Attribute{Computed: true},
									"status":         schema.Int64Attribute{Computed: true},
									"actions": schema.ListNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"action_name": schema.StringAttribute{Computed: true},
												"action_key":  schema.StringAttribute{Computed: true},
												"email":       schema.StringAttribute{Computed: true},
												"value":       schema.StringAttribute{Computed: true},
											},
										},
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

func (d *monitoringRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *monitoringRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state monitoringRuleDataSourceModel

	rules, err := d.client.Monitoring.ListMonitoringRule(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Monitoring Rules",
			"Could not get monitoring rules, unexpected error: "+err.Error(),
		)
		return
	}

	for _, r := range rules {
		item := monitoringRuleDSItem{
			Identifier: types.StringValue(r.Identifier),
			RuleName:   types.StringValue(r.RuleName),
			Status:     types.Int64Value(int64(r.Status)),
			Frequency:  types.Int64Value(int64(r.Frequency)),
			CreatedOn:  types.StringValue(r.CreatedOn),
			CreatedBy:  types.StringValue(r.CreatedBy),
		}
		for _, m := range r.Metrics {
			metric := monitoringMetricDS{
				MetricType:    types.StringValue(m.MetricType),
				Condition:     types.StringValue(m.Condition),
				Threshold:     types.Int64Value(int64(m.Threshold)),
				ThresholdType: types.StringValue(m.ThresholdType),
				Period:        types.Int64Value(int64(m.Period)),
				Status:        types.Int64Value(int64(m.Status)),
			}
			for _, a := range m.Actions {
				metric.Actions = append(metric.Actions, monitoringActionDS{
					ActionName: types.StringValue(a.ActionName),
					ActionKey:  types.StringValue(a.ActionKey),
					Email:      types.StringValue(a.Email),
					Value:      types.StringValue(a.Value),
				})
			}
			item.Metrics = append(item.Metrics, metric)
		}
		state.Rules = append(state.Rules, item)
	}

	state.ID = types.StringValue("monitoring_rules")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
