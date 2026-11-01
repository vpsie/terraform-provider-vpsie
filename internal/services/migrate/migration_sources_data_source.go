package migrate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &migrationSourcesDataSource{}
	_ datasource.DataSourceWithConfigure = &migrationSourcesDataSource{}
)

type migrationSourcesDataSource struct {
	api *sourceAPI
}

type migrationSourcesModel struct {
	ID      types.String           `tfsdk:"id"`
	Type    types.String           `tfsdk:"type"`
	Sources []migrationSourceEntry `tfsdk:"sources"`
}

type migrationSourceEntry struct {
	Identifier           types.String `tfsdk:"identifier"`
	Type                 types.String `tfsdk:"type"`
	Name                 types.String `tfsdk:"name"`
	DatacenterIdentifier types.String `tfsdk:"datacenter_identifier"`
	Host                 types.String `tfsdk:"host"`
	Username             types.String `tfsdk:"username"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
}

// NewMigrationSourcesDataSource creates the vpsie_migration_sources data source.
func NewMigrationSourcesDataSource() datasource.DataSource {
	return &migrationSourcesDataSource{}
}

func (d *migrationSourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_migration_sources"
}

func (d *migrationSourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the account's Migrate sources. Credentials are never returned.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"type": schema.StringAttribute{
				MarkdownDescription: "Only sources of this type, for example `vmware-vcenter`.",
				Optional:            true,
			},
			"sources": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"identifier":            str("The source identifier."),
					"type":                  str("The source type."),
					"name":                  str("The source name."),
					"datacenter_identifier": str("The datacenter whose nodes run the source's jobs."),
					"host":                  str("The host the source connects to."),
					"username":              str("The user the source signs in as."),
					"status":                str("The source status."),
					"created_at":            str("When the source was created (RFC 3339)."),
				}},
			},
		},
	}
}

func (d *migrationSourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	d.api = &sourceAPI{client: client}
}

func (d *migrationSourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg migrationSourcesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.api.listAll(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing migration sources", err.Error())
		return
	}
	cfg.ID = types.StringValue("migration_sources")
	cfg.Sources = make([]migrationSourceEntry, 0, len(items))
	for _, s := range items {
		if t := cfg.Type.ValueString(); t != "" && s.Type != t {
			continue
		}
		cfg.Sources = append(cfg.Sources, migrationSourceEntry{
			Identifier: types.StringValue(s.Identifier), Type: types.StringValue(s.Type), Name: types.StringValue(s.Name),
			DatacenterIdentifier: types.StringValue(s.DatacenterIdentifier), Host: types.StringValue(s.Host),
			Username: types.StringValue(s.Username), Status: types.StringValue(s.Status), CreatedAt: types.StringValue(s.CreatedAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
