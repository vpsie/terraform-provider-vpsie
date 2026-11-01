package files

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

var (
	_ datasource.DataSource              = &sharesDataSource{}
	_ datasource.DataSourceWithConfigure = &sharesDataSource{}
)

type sharesDataSource struct {
	api *shareAPI
}

type sharesDataSourceModel struct {
	ID           types.String     `tfsdk:"id"`
	DcIdentifier types.String     `tfsdk:"dc_identifier"`
	Shares       []shareListModel `tfsdk:"shares"`
}

type shareListModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	DcID       types.String `tfsdk:"dc_identifier"`
	SizeGb     types.Int64  `tfsdk:"size_gb"`
	UsedBytes  types.Int64  `tfsdk:"used_bytes"`
	State      types.String `tfsdk:"state"`
	GatewayIP  types.String `tfsdk:"gateway_vpc_ip"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

// NewSharesDataSource creates the vpsie_file_shares data source.
func NewSharesDataSource() datasource.DataSource {
	return &sharesDataSource{}
}

func (d *sharesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_shares"
}

func (d *sharesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	num := func(desc string) schema.Int64Attribute {
		return schema.Int64Attribute{MarkdownDescription: desc, Computed: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the account's file shares.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"dc_identifier": schema.StringAttribute{
				MarkdownDescription: "Only shares in this datacenter.",
				Optional:            true,
			},
			"shares": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":     str("The share identifier."),
						"name":           str("The share name."),
						"dc_identifier":  str("The datacenter identifier."),
						"size_gb":        num("The share size in GB."),
						"used_bytes":     num("Bytes used, as last measured."),
						"state":          str("The share state."),
						"gateway_vpc_ip": str("The address inside the VPC that serves the share."),
						"created_at":     str("When the share was created (RFC 3339)."),
					},
				},
			},
		},
	}
}

func (d *sharesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Data Source Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return
	}
	d.api = &shareAPI{api: api}
}

func (d *sharesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg sharesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.api.list(ctx, cfg.DcIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing file shares", err.Error())
		return
	}
	cfg.ID = types.StringValue("file_shares")
	cfg.Shares = make([]shareListModel, 0, len(items))
	for i := range items {
		s := &items[i]
		gw := types.StringNull()
		if s.Gateway != nil && s.Gateway.VPCIP != "" {
			gw = types.StringValue(s.Gateway.VPCIP)
		}
		cfg.Shares = append(cfg.Shares, shareListModel{
			Identifier: types.StringValue(s.Identifier),
			Name:       types.StringValue(s.Name),
			DcID:       types.StringValue(s.DcIdentifier),
			SizeGb:     types.Int64Value(s.SizeGb),
			UsedBytes:  types.Int64Value(s.UsedBytes),
			State:      types.StringValue(s.State),
			GatewayIP:  gw,
			CreatedAt:  types.StringValue(s.CreatedAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
