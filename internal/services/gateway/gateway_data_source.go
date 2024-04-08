package gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type gatewayDataSource struct {
	client *govpsie.Client
}

type gatewayDataSourceModel struct {
	Gateways []gatewaysModel `tfsdk:"gateways"`
	ID       types.String    `tfsdk:"id"`
}

type gatewaysModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	DatacenterID         types.Int64  `tfsdk:"datacenter_id"`
	IPPropertiesID       types.Int64  `tfsdk:"ip_properties_id"`
	IP                   types.String `tfsdk:"ip"`
	IsReserved           types.Int64  `tfsdk:"is_reserved"`
	IPVersion            types.String `tfsdk:"ip_version"`
	BoxID                types.Int64  `tfsdk:"box_id,omitempty"`
	IsPrimary            types.Int64  `tfsdk:"is_primary"`
	Notes                types.String `tfsdk:"notes,omitempty"`
	UserID               types.Int64  `tfsdk:"user_id"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
	IsGatewayReserved    types.Int64  `tfsdk:"is_gateway_reserved"`
	IsUserAccountGateway types.Int64  `tfsdk:"is_user_account_gateway"`
	DatacenterName       types.String `tfsdk:"datacenter_name"`
	State                types.String `tfsdk:"state"`
	DcIdentifier         types.String `tfsdk:"dc_identifier"`
	CreatedBy            types.String `tfsdk:"created_by"`
	AttachedVms          []AttachedVM `tfsdk:"attached_vms"`
}

type AttachedVM struct {
	Identifier       types.String `tfsdk:"identifier"`
	GatewayMappingID types.Int64  `tfsdk:"gateway_mapping_id"`
}

// NewGatewayDataSource is a helper function to create the data source.
func NewGatewayDataSource() datasource.DataSource {
	return &gatewayDataSource{}
}

// Metadata returns the data source type name.
func (g *gatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateways"
}

// Schema defines the schema for the data source.
func (g *gatewayDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"gateways": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"datacenter_id": schema.Int64Attribute{
							Computed: true,
						},
						"ip_properties_id": schema.Int64Attribute{
							Computed: true,
						},
						"ip": schema.StringAttribute{
							Computed: true,
						},
						"is_reserved": schema.Int64Attribute{
							Computed: true,
						},
						"ip_version": schema.StringAttribute{
							Computed: true,
						},
						"box_id": schema.Int64Attribute{
							Computed: true,
						},
						"is_primary": schema.Int64Attribute{
							Computed: true,
						},
						"notes": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"updated_at": schema.StringAttribute{
							Computed: true,
						},
						"is_gateway_reserved": schema.Int64Attribute{
							Computed: true,
						},
						"is_user_account_gateway": schema.Int64Attribute{
							Computed: true,
						},
						"datacenter_name": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
						"attached_vms": schema.ListNestedAttribute{
							Computed: true,
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"identifier": schema.StringAttribute{
										Computed: true,
									},
									"gateway_mapping_id": schema.Int64Attribute{
										Computed: true,
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

// Read refreshes the Terraform state with the latest data.
func (g *gatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state gatewayDataSourceModel

	gateways, err := g.client.Gateway.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting gateways",
			"Could not get gateways, unexpected error: "+err.Error(),
		)

		return
	}

	for _, gateway := range gateways {
		attached := []AttachedVM{}
		if gateway.AttachedVms != nil && len(gateway.AttachedVms) > 0 {
			for _, vm := range gateway.AttachedVms {
				attached = append(attached, AttachedVM{
					Identifier:       types.StringValue(vm.Identifier),
					GatewayMappingID: types.Int64Value(vm.GatewayMappingID),
				})
			}
		}
		gatewayState := gatewaysModel{
			ID:                   types.Int64Value(gateway.ID),
			DatacenterID:         types.Int64Value(gateway.DatacenterID),
			IPPropertiesID:       types.Int64Value(gateway.IPPropertiesID),
			IP:                   types.StringValue(gateway.IP),
			IsReserved:           types.Int64Value(gateway.IsReserved),
			IPVersion:            types.StringValue(gateway.IPVersion),
			BoxID:                types.Int64PointerValue(gateway.BoxID),
			IsPrimary:            types.Int64Value(gateway.IsPrimary),
			Notes:                types.StringPointerValue(gateway.Notes),
			UserID:               types.Int64Value(gateway.UserID),
			UpdatedAt:            types.StringValue(gateway.UpdatedAt.String()),
			IsGatewayReserved:    types.Int64Value(gateway.IsGatewayReserved),
			IsUserAccountGateway: types.Int64Value(gateway.IsUserAccountGateway),
			DatacenterName:       types.StringValue(gateway.DatacenterName),
			State:                types.StringValue(gateway.State),
			DcIdentifier:         types.StringValue(gateway.DcIdentifier),
			CreatedBy:            types.StringValue(gateway.CreatedBy),
			AttachedVms:          attached,
		}

		state.Gateways = append(state.Gateways, gatewayState)
	}
}
func (g *gatewayDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	g.client = client
}
