package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type vpcDataSource struct {
	client *govpsie.Client
}

type vpcDataSourceModel struct {
	Vpcs []vpcsModel  `tfsdk:"vpcs"`
	ID   types.String `tfsdk:"id"`
}

type vpcsModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	UserID           types.Int64  `tfsdk:"user_id"`
	OwnerID          types.Int64  `tfsdk:"owner_id"`
	DatacenterID     types.Int64  `tfsdk:"datacenter_id"`
	Description      types.String `tfsdk:"description"`
	InterfaceNumber  types.Int64  `tfsdk:"interface_number"`
	NetworkTagNumber types.Int64  `tfsdk:"network_tag_number"`
	NetworkRange     types.String `tfsdk:"network_range"`
	NetworkSize      types.Int64  `tfsdk:"network_size"`
	IsDefault        types.Int64  `tfsdk:"is_default"`
	CreatedBy        types.Int64  `tfsdk:"created_by"`
	UpdatedBy        types.Int64  `tfsdk:"updated_by"`
	CreatedOn        types.String `tfsdk:"created_on"`
	LastUpdated      types.String `tfsdk:"last_updated"`
	LowIPNum         types.Int64  `tfsdk:"low_ip_num"`
	HightIPNum       types.Int64  `tfsdk:"hight_ip_num"`
	IsUpcNetwork     types.Int64  `tfsdk:"is_upc_network"`
	Firstname        types.String `tfsdk:"firstname"`
	Lastname         types.String `tfsdk:"lastname"`
	Username         types.String `tfsdk:"username"`
	State            types.String `tfsdk:"state"`
	DcName           types.String `tfsdk:"dc_name"`
	DcIdentifier     types.String `tfsdk:"dc_identifier"`
}

// NewVpcDataSource is a helper function to create the data source.
func NewVpcDataSource() datasource.DataSource {
	return &vpcDataSource{}
}

// Metadata returns the data source type name.
func (v *vpcDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

// Schema defines the schema for the data source.
func (v *vpcDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"vpcs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"owner_id": schema.Int64Attribute{
							Computed: true,
						},
						"datacenter_id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"interface_number": schema.Int64Attribute{
							Computed: true,
						},
						"network_tag_number": schema.Int64Attribute{
							Computed: true,
						},
						"network_range": schema.StringAttribute{
							Computed: true,
						},
						"network_size": schema.Int64Attribute{
							Computed: true,
						},
						"is_default": schema.Int64Attribute{
							Computed: true,
						},
						"created_by": schema.Int64Attribute{
							Computed: true,
						},
						"updated_by": schema.Int64Attribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
						"low_ip_num": schema.Int64Attribute{
							Computed: true,
						},
						"hight_ip_num": schema.Int64Attribute{
							Computed: true,
						},
						"is_upc_network": schema.Int64Attribute{
							Computed: true,
						},
						"firstname": schema.StringAttribute{
							Computed: true,
						},
						"lastname": schema.StringAttribute{
							Computed: true,
						},
						"username": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"dc_name": schema.StringAttribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (v *vpcDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vpcDataSourceModel

	vpcs, err := v.client.VPC.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting VPCs",
			"Could not get VPCs, unexpected error: "+err.Error(),
		)

		return
	}

	for _, vpc := range vpcs {
		vpcState := vpcsModel{
			ID:               types.Int64Value(int64(vpc.ID)),
			UserID:           types.Int64Value(int64(vpc.UserID)),
			OwnerID:          types.Int64Value(int64(vpc.OwnerID)),
			DatacenterID:     types.Int64Value(int64(vpc.DatacenterID)),
			Name:             types.StringValue(vpc.Name),
			Description:      types.StringValue(vpc.Description),
			InterfaceNumber:  types.Int64Value(int64(vpc.InterfaceNumber)),
			NetworkTagNumber: types.Int64Value(int64(vpc.NetworkTagNumber)),
			NetworkRange:     types.StringValue(vpc.NetworkRange),
			NetworkSize:      types.Int64Value(int64(vpc.NetworkSize)),
			IsDefault:        types.Int64Value(int64(vpc.IsDefault)),
			CreatedBy:        types.Int64Value(int64(vpc.CreatedBy)),
			UpdatedBy:        types.Int64Value(int64(vpc.UpdatedBy)),
			CreatedOn:        types.StringValue(vpc.CreatedOn.String()),
			LastUpdated:      types.StringValue(vpc.LastUpdated.String()),
			LowIPNum:         types.Int64Value(int64(vpc.LowIPNum)),
			HightIPNum:       types.Int64Value(int64(vpc.HightIPNum)),
			IsUpcNetwork:     types.Int64Value(int64(vpc.IsUpcNetwork)),
			Firstname:        types.StringValue(vpc.Firstname),
			Lastname:         types.StringValue(vpc.Lastname),
			Username:         types.StringValue(vpc.Username),
			State:            types.StringValue(vpc.State),
			DcName:           types.StringValue(vpc.DcName),
			DcIdentifier:     types.StringValue(vpc.DcIdentifier),
		}

		state.Vpcs = append(state.Vpcs, vpcState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (v *vpcDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	v.client = client
}
