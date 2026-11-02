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
	_ datasource.DataSource              = &policiesDataSource{}
	_ datasource.DataSourceWithConfigure = &policiesDataSource{}
)

type policiesDataSource struct {
	api *policyAPI
}

type policiesDataSourceModel struct {
	ID       types.String      `tfsdk:"id"`
	State    types.String      `tfsdk:"state"`
	Policies []policyListModel `tfsdk:"policies"`
}

type policyListModel struct {
	Identifier       types.String `tfsdk:"identifier"`
	Name             types.String `tfsdk:"name"`
	Tier             types.String `tfsdk:"tier"`
	State            types.String `tfsdk:"state"`
	Datacenter       types.String `tfsdk:"datacenter"`
	RPOMinutes       types.Int64  `tfsdk:"rpo_minutes"`
	ImmutabilityDays types.Int64  `tfsdk:"immutability_days"`
	Consistency      types.String `tfsdk:"consistency"`
	VerifyMode       types.String `tfsdk:"verify_mode"`
	Version          types.Int64  `tfsdk:"version"`
}

// NewPoliciesDataSource creates the vpsie_protect_policies data source.
func NewPoliciesDataSource() datasource.DataSource {
	return &policiesDataSource{}
}

func (d *policiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protect_policies"
}

func (d *policiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	num := func(desc string) schema.Int64Attribute {
		return schema.Int64Attribute{MarkdownDescription: desc, Computed: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the account's Backup and Protect policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"state": schema.StringAttribute{
				MarkdownDescription: "Only policies in this state: `active`, `paused` or `pending_approval`.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf("active", "paused", "pending_approval")},
			},
			"policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":        str("The policy identifier."),
						"name":              str("The policy name."),
						"tier":              str("`backup` or `dr`."),
						"state":             str("The policy state."),
						"datacenter":        str("The datacenter identifier."),
						"rpo_minutes":       num("The recovery point objective in minutes."),
						"immutability_days": num("Days a recovery point cannot be deleted."),
						"consistency":       str("`app` or `crash`."),
						"verify_mode":       str("`off`, `latest-daily` or `latest-weekly`."),
						"version":           num("The policy version."),
					},
				},
			},
		},
	}
}

func (d *policiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Data Source Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return
	}
	d.api = &policyAPI{api: api}
}

func (d *policiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg policiesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.api.list(ctx, cfg.State.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing protection policies", err.Error())
		return
	}
	cfg.ID = types.StringValue("protect_policies")
	cfg.Policies = make([]policyListModel, 0, len(items))
	for i := range items {
		p := &items[i]
		dc := types.StringNull()
		if p.Datacenter.Identifier != nil {
			dc = types.StringValue(*p.Datacenter.Identifier)
		}
		cfg.Policies = append(cfg.Policies, policyListModel{
			Identifier:       types.StringValue(p.Identifier),
			Name:             types.StringValue(p.Name),
			Tier:             types.StringValue(p.Tier),
			State:            types.StringValue(p.State),
			Datacenter:       dc,
			RPOMinutes:       types.Int64Value(p.RPOMinutes),
			ImmutabilityDays: types.Int64Value(p.ImmutabilityDays),
			Consistency:      types.StringValue(p.Consistency),
			VerifyMode:       types.StringValue(p.VerifyMode),
			Version:          types.Int64Value(p.Version),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
