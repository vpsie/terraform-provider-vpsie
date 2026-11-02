package files

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

var (
	_ datasource.DataSource              = &smbUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &smbUsersDataSource{}
)

type smbUsersDataSource struct {
	api *shareAPI
}

type smbUsersDataSourceModel struct {
	ID    types.String       `tfsdk:"id"`
	Users []smbUserListModel `tfsdk:"users"`
}

type smbUserListModel struct {
	Identifier        types.String `tfsdk:"identifier"`
	Username          types.String `tfsdk:"username"`
	DcIdentifier      types.String `tfsdk:"dc_identifier"`
	ShareCount        types.Int64  `tfsdk:"share_count"`
	PasswordChangedAt types.String `tfsdk:"password_changed_at"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

// NewSmbUsersDataSource creates the vpsie_file_smb_users data source.
func NewSmbUsersDataSource() datasource.DataSource {
	return &smbUsersDataSource{}
}

func (d *smbUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_smb_users"
}

func (d *smbUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the account's SMB users of the Files service. A password is never part of the answer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":    str("The SMB user identifier."),
						"username":      str("The user name."),
						"dc_identifier": str("The datacenter of the user's VPC."),
						"share_count": schema.Int64Attribute{
							MarkdownDescription: "How many shares grant the user, null when the platform does not say.",
							Computed:            true,
						},
						"password_changed_at": str("When the password was last set."),
						"created_at":          str("When the user was created."),
					},
				},
			},
		},
	}
}

func (d *smbUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *smbUsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.api.listSmbUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing SMB users", err.Error())
		return
	}
	out := smbUsersDataSourceModel{ID: types.StringValue("file_smb_users"), Users: make([]smbUserListModel, 0, len(items))}
	for i := range items {
		u := &items[i]
		count := types.Int64Null()
		if u.ShareCount != nil {
			count = types.Int64Value(*u.ShareCount)
		}
		out.Users = append(out.Users, smbUserListModel{
			Identifier:        types.StringValue(u.Identifier),
			Username:          types.StringValue(u.Username),
			DcIdentifier:      types.StringValue(u.DcIdentifier),
			ShareCount:        count,
			PasswordChangedAt: types.StringValue(u.PasswordChangedAt),
			CreatedAt:         types.StringValue(u.CreatedAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}
