package ip

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

type ipDataSource struct {
	client *govpsie.Client
}

type ipDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
	IPs  []ipModel    `tfsdk:"ips"`
}

type ipModel struct {
	ID            types.Int64  `tfsdk:"id"`
	DcName        types.String `tfsdk:"dc_name"`
	DcIdentifier  types.String `tfsdk:"dc_identifier"`
	IP            types.String `tfsdk:"ip"`
	IPVersion     types.String `tfsdk:"ip_version"`
	IsPrimary     types.Int64  `tfsdk:"is_primary"`
	Hostname      types.String `tfsdk:"hostname"`
	BoxIdentifier types.String `tfsdk:"box_identifier"`
	FullName      types.String `tfsdk:"full_name"`
	Category      types.String `tfsdk:"category"`
	Type          types.String `tfsdk:"type"`
	CreatedBy     types.String `tfsdk:"created_by"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewIPDataSource() datasource.DataSource {
	return &ipDataSource{}
}

func (d *ipDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ips"
}

func (d *ipDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by IP type: `all`, `public`, or `private`. Defaults to `all`.",
			},
			"ips": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.Int64Attribute{Computed: true},
						"dc_name":        schema.StringAttribute{Computed: true},
						"dc_identifier":  schema.StringAttribute{Computed: true},
						"ip":             schema.StringAttribute{Computed: true},
						"ip_version":     schema.StringAttribute{Computed: true},
						"is_primary":     schema.Int64Attribute{Computed: true},
						"hostname":       schema.StringAttribute{Computed: true},
						"box_identifier": schema.StringAttribute{Computed: true},
						"full_name":      schema.StringAttribute{Computed: true},
						"category":       schema.StringAttribute{Computed: true},
						"type":           schema.StringAttribute{Computed: true},
						"created_by":     schema.StringAttribute{Computed: true},
						"updated_at":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ipDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ipDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ipDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipType := "all"
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		ipType = config.Type.ValueString()
	}

	var ips []govpsie.IP
	var err error

	switch ipType {
	case "public":
		ips, err = d.client.IP.ListPublicIPs(ctx, nil)
	case "private":
		ips, err = d.client.IP.ListPrivateIPs(ctx, nil)
	default:
		ips, err = d.client.IP.ListAllIPs(ctx, nil)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting IPs",
			"Could not get IPs, unexpected error: "+err.Error(),
		)
		return
	}

	var state ipDataSourceModel
	state.Type = config.Type

	for _, ipItem := range ips {
		state.IPs = append(state.IPs, ipModel{
			ID:            types.Int64Value(int64(ipItem.ID)),
			DcName:        types.StringValue(ipItem.DcName),
			DcIdentifier:  types.StringValue(ipItem.DcIdentifier),
			IP:            types.StringValue(ipItem.IP),
			IPVersion:     types.StringValue(ipItem.IPVersion),
			IsPrimary:     types.Int64Value(int64(ipItem.IsPrimary)),
			Hostname:      types.StringValue(ipItem.Hostname),
			BoxIdentifier: types.StringValue(ipItem.BoxIdentifier),
			FullName:      types.StringValue(ipItem.FullName),
			Category:      types.StringValue(ipItem.Category),
			Type:          types.StringValue(ipItem.Type),
			CreatedBy:     types.StringValue(ipItem.CreatedBy),
			UpdatedAt:     types.StringValue(ipItem.UpdatedAt),
		})
	}

	state.ID = types.StringValue("ips")

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
