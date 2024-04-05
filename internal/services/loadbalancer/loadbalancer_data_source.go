package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type loadbalancerDataSource struct {
	client *govpsie.Client
}

type loadbalancerDataSourceModel struct {
	Loadbalancers []loadbalancersModel `tfsdk:"loadbalancers"`
	ID            types.String         `tfsdk:"id"`
}

type loadbalancersModel struct {
	Cpu        types.Int64  `tfsdk:"cpu"`
	Ssd        types.Int64  `tfsdk:"ssd"`
	Ram        types.Int64  `tfsdk:"ram"`
	LBName     types.String `tfsdk:"lb_name"`
	Traffic    types.Int64  `tfsdk:"traffic"`
	BoxsizeID  types.Int64  `tfsdk:"boxsize_id"`
	DefaultIP  types.String `tfsdk:"default_ip"`
	DCName     types.String `tfsdk:"dc_name"`
	Identifier types.String `tfsdk:"identifier"`
	CreatedOn  types.String `tfsdk:"created_on"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	Package    types.String `tfsdk:"package"`
	CreatedBy  types.String `tfsdk:"created_by"`
	UserID     types.Int64  `tfsdk:"user_id"`
}

// NewLoadbalancerDataSource is a helper function to create the data source.
func NewLoadbalancerDataSource() datasource.DataSource {
	return &loadbalancerDataSource{}
}

// Metadata returns the data source type name.
func (l *loadbalancerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancers"
}

// Schema defines the schema for the data source.
func (l *loadbalancerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"loadbalancers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"cpu": schema.Int64Attribute{
							Computed: true,
						},
						"ssd": schema.Int64Attribute{
							Computed: true,
						},
						"ram": schema.Int64Attribute{
							Computed: true,
						},
						"lb_name": schema.StringAttribute{
							Computed: true,
						},
						"traffic": schema.Int64Attribute{
							Computed: true,
						},
						"boxsize_id": schema.Int64Attribute{
							Computed: true,
						},
						"default_ip": schema.StringAttribute{
							Computed: true,
						},
						"dc_name": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"updated_at": schema.StringAttribute{
							Computed: true,
						},
						"package": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (k *loadbalancerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state loadbalancerDataSourceModel

	loadbalancers, err := k.client.LB.ListLBs(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Loadbalancers",
			"Could not get Loadbalancers, unexpected error: "+err.Error(),
		)

		return
	}

	for _, lb := range loadbalancers {
		lbState := loadbalancersModel{
			Identifier: types.StringValue(lb.Identifier),
			UserID:     types.Int64Value(int64(lb.UserID)),
			CreatedOn:  types.StringValue(lb.CreatedOn),
			CreatedBy:  types.StringValue(lb.CreatedBy),
			UpdatedAt:  types.StringValue(lb.UpdatedAt),
			Package:    types.StringValue(lb.Package),
			Cpu:        types.Int64Value(int64(lb.Cpu)),
			Ram:        types.Int64Value(int64(lb.Ram)),
			Ssd:        types.Int64Value(int64(lb.Ssd)),
			LBName:     types.StringValue(lb.LBName),
			Traffic:    types.Int64Value(int64(lb.Traffic)),
			BoxsizeID:  types.Int64Value(int64(lb.BoxsizeID)),
			DefaultIP:  types.StringValue(lb.DefaultIP),
			DCName:     types.StringValue(lb.DCName),
		}

		state.Loadbalancers = append(state.Loadbalancers, lbState)
	}
}
func (l *loadbalancerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report  this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	l.client = client
}
