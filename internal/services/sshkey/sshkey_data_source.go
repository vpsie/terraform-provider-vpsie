package sshkey

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
)

type sshKeyDataSource struct {
	client *govpsie.Client
}

type sshKeyDataSourceModel struct {
	SshKeys []sshKeysModel `tfsdk:"sshkeys"`
	ID      types.String   `tfsdk:"id"`
}

type sshKeysModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Id         types.Int64  `tfsdk:"id"`
	UserId     types.Int64  `tfsdk:"user_id"`
	Name       types.String `tfsdk:"name"`
	PrivateKey types.String `tfsdk:"private_key"`
	CreatedOn  types.String `tfsdk:"created_on"`
	CreatedBy  types.String `tfsdk:"created_by"`
}

// NewSshKeyDataSource is a helper function to create the data source.
func NewSshKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

// Metadata returns the data source type name.
func (s *sshKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sshkeys"
}

// Schema defines the schema for the data source.
func (s *sshKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"sshkeys": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"private_key": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sshKeyDataSourceModel

	sshKeys, err := s.client.SShKey.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get sshKeys",
			"An unexpected error occurred when getting sshKeys: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Got sshKeys", map[string]interface{}{"sshKeys": sshKeys})

	for _, sshKey := range sshKeys {

		sshKeyState := sshKeysModel{
			Identifier: types.StringValue(sshKey.Identifier),
			CreatedOn:  types.StringValue(sshKey.CreatedOn),
			CreatedBy:  types.StringValue(sshKey.CreatedBy),
			Name:       types.StringValue(sshKey.Name),
			PrivateKey: types.StringValue(sshKey.PrivateKey),
			UserId:     types.Int64Value(sshKey.UserId),
			Id:         types.Int64Value(sshKey.Id),
		}

		state.SshKeys = append(state.SshKeys, sshKeyState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *sshKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	s.client = client
}
