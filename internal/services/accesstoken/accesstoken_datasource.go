package accesstoken

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

type accessTokenDataSource struct {
	client *govpsie.Client
}

type accessTokenDataSourceModel struct {
	ID     types.String       `tfsdk:"id"`
	Tokens []accessTokenModel `tfsdk:"tokens"`
}

type accessTokenModel struct {
	Identifier     types.String `tfsdk:"identifier"`
	Name           types.String `tfsdk:"name"`
	ExpirationDate types.String `tfsdk:"expiration_date"`
	CreatedOn      types.String `tfsdk:"created_on"`
}

func NewAccessTokenDataSource() datasource.DataSource {
	return &accessTokenDataSource{}
}

func (d *accessTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_tokens"
}

func (d *accessTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"tokens": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"expiration_date": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *accessTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *accessTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessTokenDataSourceModel

	tokens, err := d.client.AccessToken.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Access Tokens",
			"Could not get access tokens, unexpected error: "+err.Error(),
		)
		return
	}

	for _, t := range tokens {
		state.Tokens = append(state.Tokens, accessTokenModel{
			Identifier:     types.StringValue(t.AccessTokenIdentifier),
			Name:           types.StringValue(t.Name),
			ExpirationDate: types.StringValue(t.ExpirationDate),
			CreatedOn:      types.StringValue(t.CreatedOn),
		})
	}

	state.ID = types.StringValue("access_tokens")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
