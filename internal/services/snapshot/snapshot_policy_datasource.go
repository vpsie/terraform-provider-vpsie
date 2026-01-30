package snapshot

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

type snapshotPolicyDataSource struct {
	client *govpsie.Client
}

type snapshotPolicyDataSourceModel struct {
	ID       types.String                `tfsdk:"id"`
	Policies []snapshotPolicyListModel `tfsdk:"policies"`
}

type snapshotPolicyListModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	BackupPlan types.String `tfsdk:"backup_plan"`
	PlanEvery  types.Int64  `tfsdk:"plan_every"`
	Keep       types.Int64  `tfsdk:"keep"`
	Disabled   types.Int64  `tfsdk:"disabled"`
	VmsCount   types.Int64  `tfsdk:"vms_count"`
	CreatedOn  types.String `tfsdk:"created_on"`
	CreatedBy  types.String `tfsdk:"created_by"`
}

func NewSnapshotPolicyDataSource() datasource.DataSource {
	return &snapshotPolicyDataSource{}
}

func (d *snapshotPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot_policies"
}

func (d *snapshotPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"backup_plan": schema.StringAttribute{
							Computed: true,
						},
						"plan_every": schema.Int64Attribute{
							Computed: true,
						},
						"keep": schema.Int64Attribute{
							Computed: true,
						},
						"disabled": schema.Int64Attribute{
							Computed: true,
						},
						"vms_count": schema.Int64Attribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *snapshotPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *snapshotPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state snapshotPolicyDataSourceModel

	policies, err := d.client.Snapshot.ListSnapShotPolicies(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Snapshot Policies",
			"Could not get snapshot policies, unexpected error: "+err.Error(),
		)
		return
	}

	for _, p := range policies {
		state.Policies = append(state.Policies, snapshotPolicyListModel{
			Identifier: types.StringValue(p.Identifier),
			Name:       types.StringValue(p.Name),
			BackupPlan: types.StringValue(p.BackupPlan),
			PlanEvery:  types.Int64Value(p.PlanEvery),
			Keep:       types.Int64Value(p.Keep),
			Disabled:   types.Int64Value(p.Disabled),
			VmsCount:   types.Int64Value(p.VmsCount),
			CreatedOn:  types.StringValue(p.CreatedOn),
			CreatedBy:  types.StringValue(p.CreatedBy),
		})
	}

	state.ID = types.StringValue("snapshot_policies")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
