package certificate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &certificateDataSource{}
	_ datasource.DataSourceWithConfigure = &certificateDataSource{}
)

type certificateDataSource struct {
	client *govpsie.Client
}

type certificateDataSourceModel struct {
	Certificates []certificateModel `tfsdk:"certificates"`
	ID           types.String       `tfsdk:"id"`
}

type certificateModel struct {
	Identifier      types.String `tfsdk:"identifier"`
	CertificateName types.String `tfsdk:"certificate_name"`
	DomainName      types.String `tfsdk:"domain_name"`
	Issuer          types.String `tfsdk:"issuer"`
	Serial          types.String `tfsdk:"serial"`
	CertType        types.String `tfsdk:"cert_type"`
	ValidFrom       types.String `tfsdk:"valid_from"`
	ValidTo         types.String `tfsdk:"valid_to"`
	CreatedOn       types.String `tfsdk:"created_on"`
}

// NewCertificateDataSource is a helper function to create the data source.
func NewCertificateDataSource() datasource.DataSource {
	return &certificateDataSource{}
}

func (c *certificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (c *certificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all VPSie TLS certificates.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"certificates": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":       schema.StringAttribute{Computed: true},
						"certificate_name": schema.StringAttribute{Computed: true},
						"domain_name":      schema.StringAttribute{Computed: true},
						"issuer":           schema.StringAttribute{Computed: true},
						"serial":           schema.StringAttribute{Computed: true},
						"cert_type":        schema.StringAttribute{Computed: true},
						"valid_from":       schema.StringAttribute{Computed: true},
						"valid_to":         schema.StringAttribute{Computed: true},
						"created_on":       schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (c *certificateDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state certificateDataSourceModel

	certs, err := c.client.Certificate.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get certificates",
			"An unexpected error occurred when getting certificates: "+err.Error(),
		)

		return
	}

	for _, cert := range certs {
		state.Certificates = append(state.Certificates, certificateModel{
			Identifier:      types.StringValue(cert.Identifier),
			CertificateName: types.StringValue(cert.CertificateName),
			DomainName:      types.StringValue(cert.DomainName),
			Issuer:          types.StringValue(cert.Issuer),
			Serial:          types.StringValue(cert.Serial),
			CertType:        types.StringValue(cert.CertType),
			ValidFrom:       types.StringValue(cert.ValidFrom),
			ValidTo:         types.StringValue(cert.ValidTo),
			CreatedOn:       types.StringValue(cert.CreatedOn),
		})
	}

	state.ID = types.StringValue("certificates")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (c *certificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	c.client = client
}
