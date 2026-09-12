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
	ID    types.String          `tfsdk:"id"`
	Rules []monitoringRuleModel `tfsdk:"rules"`
}

type monitoringRuleModel struct {
	Identifier    types.String `tfsdk:"identifier"`
	RuleName      types.String `tfsdk:"rule_name"`
	MetricType    types.String `tfsdk:"metric_type"`
	Condition     types.String `tfsdk:"condition"`
	ThresholdType types.String `tfsdk:"threshold_type"`
	Threshold     types.Int64  `tfsdk:"threshold"`
	Period        types.Int64  `tfsdk:"period"`
	Frequency     types.Int64  `tfsdk:"frequency"`
	Status        types.Int64  `tfsdk:"status"`
	Email         types.String `tfsdk:"email"`
	CreatedOn     types.String `tfsdk:"created_on"`
	CreatedBy     types.String `tfsdk:"created_by"`
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
			"id": schema.StringAttribute{
				Computed: true,
			},
			"rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":     schema.StringAttribute{Computed: true},
						"rule_name":      schema.StringAttribute{Computed: true},
						"metric_type":    schema.StringAttribute{Computed: true},
						"condition":      schema.StringAttribute{Computed: true},
						"threshold_type": schema.StringAttribute{Computed: true},
						"threshold":      schema.Int64Attribute{Computed: true},
						"period":         schema.Int64Attribute{Computed: true},
						"frequency":      schema.Int64Attribute{Computed: true},
						"status":         schema.Int64Attribute{Computed: true},
						"email":          schema.StringAttribute{Computed: true},
						"created_on":     schema.StringAttribute{Computed: true},
						"created_by":     schema.StringAttribute{Computed: true},
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
		state.Rules = append(state.Rules, monitoringRuleModel{
			Identifier:    types.StringValue(r.Identifier),
			RuleName:      types.StringValue(r.RuleName),
			MetricType:    types.StringValue(r.MetricType),
			Condition:     types.StringValue(r.Condition),
			ThresholdType: types.StringValue(r.ThresholdType),
			Threshold:     types.Int64Value(int64(r.Threshold)),
			Period:        types.Int64Value(int64(r.Period)),
			Frequency:     types.Int64Value(int64(r.Frequency)),
			Status:        types.Int64Value(int64(r.Status)),
			Email:         types.StringValue(r.Email),
			CreatedOn:     types.StringValue(r.CreatedOn),
			CreatedBy:     types.StringValue(r.CreatedBY),
		})
	}

	state.ID = types.StringValue("monitoring_rules")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
