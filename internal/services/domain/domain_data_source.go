package domain

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type domainDataSource struct {
	client *govpsie.Client
}

type domainDataSourceModel struct {
	Domains []domainsModel `tfsdk:"domains"`
	ID      types.String   `tfsdk:"id"`
}

type domainsModel struct {
	Identifier  types.String `tfsdk:"identifier"`
	DomainName  types.String `tfsdk:"domain_name"`
	NsValidated types.Int64  `tfsdk:"ns_validated"`
	CreatedOn   types.String `tfsdk:"created_on"`
	LastCheck   types.String `tfsdk:"last_check"`
}

// NewDomainDataSource is a helper function to create the data source.
func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

// Metadata returns the data source type name.
func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

// Schema defines the schema for the data source.
func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"domains": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"domain_name": schema.StringAttribute{
							Computed: true,
						},
						"ns_validated": schema.Int64Attribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"last_check": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state domainDataSourceModel

	domains, err := d.client.Domain.ListDomains(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting domains",
			"Could not get domains, unexpected error: "+err.Error(),
		)

		return
	}

	for _, domain := range domains {
		domainState := domainsModel{
			Identifier:  types.StringValue(domain.Identifier),
			DomainName:  types.StringValue(domain.DomainName),
			NsValidated: types.Int64Value(int64(domain.NsValidated)),
			CreatedOn:   types.StringValue(domain.CreatedOn),
			LastCheck:   types.StringValue(domain.LastCheck),
		}

		state.Domains = append(state.Domains, domainState)
	}
}
func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}
