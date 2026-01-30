package datacenter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

type datacenterDataSource struct {
	client *govpsie.Client
}

type datacenterDataSourceModel struct {
	ID          types.String      `tfsdk:"id"`
	Datacenters []datacenterModel `tfsdk:"datacenters"`
}

type datacenterModel struct {
	DcName            types.String `tfsdk:"dc_name"`
	DcImage           types.String `tfsdk:"dc_image"`
	State             types.String `tfsdk:"state"`
	Country           types.String `tfsdk:"country"`
	IsActive          types.Int64  `tfsdk:"is_active"`
	Identifier        types.String `tfsdk:"identifier"`
	DefaultSelected   types.Int64  `tfsdk:"default_selected"`
	IsFipAvailable    types.Int64  `tfsdk:"is_fip_available"`
	IsBucketAvailable types.Int64  `tfsdk:"is_bucket_available"`
	IsPrivate         types.Int64  `tfsdk:"is_private"`
}

func NewDatacenterDataSource() datasource.DataSource {
	return &datacenterDataSource{}
}

func (d *datacenterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenters"
}

func (d *datacenterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"datacenters": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dc_name": schema.StringAttribute{
							Computed: true,
						},
						"dc_image": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"country": schema.StringAttribute{
							Computed: true,
						},
						"is_active": schema.Int64Attribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"default_selected": schema.Int64Attribute{
							Computed: true,
						},
						"is_fip_available": schema.Int64Attribute{
							Computed: true,
						},
						"is_bucket_available": schema.Int64Attribute{
							Computed: true,
						},
						"is_private": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *datacenterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *datacenterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datacenterDataSourceModel

	datacenters, err := d.client.DataCenter.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Datacenters",
			"Could not get datacenters, unexpected error: "+err.Error(),
		)
		return
	}

	for _, dc := range datacenters {
		state.Datacenters = append(state.Datacenters, datacenterModel{
			DcName:            types.StringValue(dc.DcName),
			DcImage:           types.StringValue(dc.DcImage),
			State:             types.StringValue(dc.State),
			Country:           types.StringValue(dc.Country),
			IsActive:          types.Int64Value(int64(dc.IsActive)),
			Identifier:        types.StringValue(dc.Identifier),
			DefaultSelected:   types.Int64Value(int64(dc.DefaultSelected)),
			IsFipAvailable:    types.Int64Value(int64(dc.IsFipAvailable)),
			IsBucketAvailable: types.Int64Value(int64(dc.IsBucketAvailable)),
			IsPrivate:         types.Int64Value(int64(dc.IsPrivate)),
		})
	}

	state.ID = types.StringValue("datacenters")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
