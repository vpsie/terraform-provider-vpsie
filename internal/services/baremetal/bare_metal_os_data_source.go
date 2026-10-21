package baremetal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &bareMetalOSDataSource{}
	_ datasource.DataSourceWithConfigure = &bareMetalOSDataSource{}
)

type bareMetalOSDataSource struct {
	client *govpsie.Client
}

type bareMetalOSDataSourceModel struct {
	ID               types.String       `tfsdk:"id"`
	DcIdentifier     types.String       `tfsdk:"dc_identifier"`
	PlanIdentifier   types.String       `tfsdk:"plan_identifier"`
	OperatingSystems []bareMetalOSModel `tfsdk:"operating_systems"`
}

type bareMetalOSModel struct {
	Identifier               types.String `tfsdk:"identifier"`
	Name                     types.String `tfsdk:"name"`
	Distribution             types.String `tfsdk:"distribution"`
	Category                 types.String `tfsdk:"category"`
	Description              types.String `tfsdk:"description"`
	OSFamily                 types.String `tfsdk:"os_family"`
	GPU                      types.Bool   `tfsdk:"gpu"`
	RemoteAccessConfigurable types.Bool   `tfsdk:"remote_access_configurable"`
	IsDefault                types.Bool   `tfsdk:"is_default"`
}

// NewBareMetalOSDataSource returns the vpsie_bare_metal_os data source.
func NewBareMetalOSDataSource() datasource.DataSource {
	return &bareMetalOSDataSource{}
}

func (d *bareMetalOSDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bare_metal_os"
}

func (d *bareMetalOSDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	computedString := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, MarkdownDescription: desc}
	}
	computedBool := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{Computed: true, MarkdownDescription: desc}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the operating systems a bare-metal server can be installed with in a datacenter, " +
			"optionally for one hardware plan (GPU plans add GPU-ready images).",
		Attributes: map[string]schema.Attribute{
			"id": computedString("The datacenter identifier, followed by the plan identifier when set."),
			"dc_identifier": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identifier of the datacenter.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
			},
			"plan_identifier": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Identifier of a hardware plan, to list the systems that plan can run.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
			},
			"operating_systems": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The installable systems, one entry per version.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":                 computedString("OS identifier (`bmos_...`), for `vpsie_bare_metal.os_identifier`."),
						"name":                       computedString("Full name, for example `Ubuntu 24.04 LTS`."),
						"distribution":               computedString("Distribution the version belongs to, for example `Ubuntu`."),
						"category":                   computedString("Category of the image."),
						"description":                computedString("Description of the image."),
						"os_family":                  computedString("`linux` or `windows`."),
						"gpu":                        computedBool("Whether the image carries GPU drivers."),
						"remote_access_configurable": computedBool("Whether `remote_access` applies to this system."),
						"is_default":                 computedBool("Whether this is the default version of its distribution."),
					},
				},
			},
		},
	}
}

func (d *bareMetalOSDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bareMetalOSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bareMetalOSDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.client.BareMetal.ListOS(ctx, state.DcIdentifier.ValueString(), state.PlanIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bare-metal operating systems", err.Error())
		return
	}

	state.OperatingSystems = []bareMetalOSModel{}
	for _, g := range groups {
		for _, v := range g.Versions {
			state.OperatingSystems = append(state.OperatingSystems, bareMetalOSModel{
				Identifier:               types.StringValue(v.Identifier),
				Name:                     types.StringValue(v.FullName),
				Distribution:             types.StringValue(g.Name),
				Category:                 types.StringValue(v.Category),
				Description:              types.StringValue(v.Description),
				OSFamily:                 types.StringValue(v.OSFamily),
				GPU:                      types.BoolValue(v.GPU),
				RemoteAccessConfigurable: types.BoolValue(v.RemoteAccessConfigurable),
				IsDefault:                types.BoolValue(bool(v.IsDefault)),
			})
		}
	}

	id := state.DcIdentifier.ValueString()
	if plan := state.PlanIdentifier.ValueString(); plan != "" {
		id += "/" + plan
	}
	state.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
