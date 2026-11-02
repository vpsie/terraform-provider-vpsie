package protect

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

var (
	_ datasource.DataSource              = &protectedServersDataSource{}
	_ datasource.DataSourceWithConfigure = &protectedServersDataSource{}
)

type protectedServersDataSource struct {
	api *serverAPI
}

type protectedServersDataSourceModel struct {
	ID      types.String               `tfsdk:"id"`
	State   types.String               `tfsdk:"state"`
	Servers []protectedServerListModel `tfsdk:"servers"`
}

type protectedServerListModel struct {
	Identifier       types.String `tfsdk:"identifier"`
	Name             types.String `tfsdk:"name"`
	Tier             types.String `tfsdk:"tier"`
	State            types.String `tfsdk:"state"`
	PausedBy         types.String `tfsdk:"paused_by"`
	PolicyIdentifier types.String `tfsdk:"policy_identifier"`
	SourceIdentifier types.String `tfsdk:"source_identifier"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

// NewProtectedServersDataSource creates the vpsie_protected_servers data source.
func NewProtectedServersDataSource() datasource.DataSource {
	return &protectedServersDataSource{}
}

func (d *protectedServersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protected_servers"
}

func (d *protectedServersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the servers the account protects, whatever created the protection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"state": schema.StringAttribute{
				MarkdownDescription: "Only servers in this state, as the platform names it (for example `protected` or `paused`).",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"servers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":        str("The protected server identifier."),
						"name":              str("The server name."),
						"tier":              str("`backup` or `dr`."),
						"state":             str("The protection state."),
						"paused_by":         str("Who paused the protection, empty when it is not paused."),
						"policy_identifier": str("The policy that protects the server."),
						"source_identifier": str("The Migrate source the server is read from."),
						"created_at":        str("When the protection started."),
					},
				},
			},
		},
	}
}

func (d *protectedServersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Data Source Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return
	}
	d.api = &serverAPI{api: api}
}

func derefString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func (d *protectedServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg protectedServersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.api.list(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing protected servers", err.Error())
		return
	}
	cfg.ID = types.StringValue("protected_servers")
	cfg.Servers = make([]protectedServerListModel, 0, len(items))
	for i := range items {
		v := &items[i]
		if !cfg.State.IsNull() && v.State != cfg.State.ValueString() {
			continue
		}
		cfg.Servers = append(cfg.Servers, protectedServerListModel{
			Identifier:       types.StringValue(v.Identifier),
			Name:             types.StringValue(v.Name),
			Tier:             types.StringValue(v.Tier),
			State:            types.StringValue(v.State),
			PausedBy:         derefString(v.PausedBy),
			PolicyIdentifier: derefString(v.Policy.Identifier),
			SourceIdentifier: derefString(v.Source.Identifier),
			CreatedAt:        types.StringValue(v.CreatedAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
