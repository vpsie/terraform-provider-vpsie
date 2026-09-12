package certificate

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &certificateResource{}
	_ resource.ResourceWithConfigure   = &certificateResource{}
	_ resource.ResourceWithImportState = &certificateResource{}
)

type certificateResource struct {
	client *govpsie.Client
}

type certificateResourceModel struct {
	Identifier types.String `tfsdk:"identifier"`
	CertName   types.String `tfsdk:"cert_name"`
	DomainID   types.String `tfsdk:"domain_id"`
	DomainName types.String `tfsdk:"domain_name"`
	Issuer     types.String `tfsdk:"issuer"`
	Serial     types.String `tfsdk:"serial"`
	CertType   types.String `tfsdk:"cert_type"`
	ValidFrom  types.String `tfsdk:"valid_from"`
	ValidTo    types.String `tfsdk:"valid_to"`
	CreatedOn  types.String `tfsdk:"created_on"`
}

// NewCertificateResource is a helper function to create the resource.
func NewCertificateResource() resource.Resource {
	return &certificateResource{}
}

func (c *certificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (c *certificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VPSie TLS certificate issued for a domain.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the certificate.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cert_name": schema.StringAttribute{
				MarkdownDescription: "The name of the certificate.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the domain the certificate is issued for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name covered by the certificate.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: "The certificate issuer.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: "The certificate serial number.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cert_type": schema.StringAttribute{
				MarkdownDescription: "The certificate type.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"valid_from": schema.StringAttribute{
				MarkdownDescription: "The start of the certificate validity period.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"valid_to": schema.StringAttribute{
				MarkdownDescription: "The end of the certificate validity period.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_on": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the certificate.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (c *certificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	c.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (c *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan certificateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := c.client.Certificate.Add(ctx, plan.CertName.ValueString(), plan.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating certificate",
			"couldn't create certificate, unexpected error: "+err.Error(),
		)

		return
	}

	cert, err := c.getCertificateByName(ctx, plan.CertName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating certificate",
			"certificate created but couldn't resolve it: "+err.Error(),
		)

		return
	}

	applyCertificate(&plan, cert)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (c *certificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state certificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := c.getCertificateByIdentifier(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading certificate",
			"couldn't read certificate "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}

	if cert == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	applyCertificate(&state, cert)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is a no-op: all configurable attributes force replacement.
func (c *certificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan certificateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (c *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state certificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := c.client.Certificate.Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting certificate",
			"couldn't delete certificate "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}
}

func (c *certificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// applyCertificate maps API fields onto the model. It deliberately does not set
// cert_name: that is a Required, config-owned attribute, and the API may store a
// suffixed variant of the requested name — overwriting it would cause an
// "inconsistent result after apply" error and perpetual replacement.
func applyCertificate(model *certificateResourceModel, cert *govpsie.Certificate) {
	model.Identifier = types.StringValue(cert.Identifier)
	model.DomainName = types.StringValue(cert.DomainName)
	model.Issuer = types.StringValue(cert.Issuer)
	model.Serial = types.StringValue(cert.Serial)
	model.CertType = types.StringValue(cert.CertType)
	model.ValidFrom = types.StringValue(cert.ValidFrom)
	model.ValidTo = types.StringValue(cert.ValidTo)
	model.CreatedOn = types.StringValue(cert.CreatedOn)
}

// getCertificateByName resolves a certificate created for the given name. The
// API may store the name verbatim or append a suffix (for example
// "my-cert-4cb5"), so an exact match is preferred and a prefix match is used as
// a fallback.
func (c *certificateResource) getCertificateByName(ctx context.Context, name string) (*govpsie.Certificate, error) {
	certs, err := c.client.Certificate.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range certs {
		if certs[i].CertificateName == name {
			return &certs[i], nil
		}
	}

	for i := range certs {
		if strings.HasPrefix(certs[i].CertificateName, name+"-") {
			return &certs[i], nil
		}
	}

	return nil, fmt.Errorf("certificate %q not found", name)
}

func (c *certificateResource) getCertificateByIdentifier(ctx context.Context, identifier string) (*govpsie.Certificate, error) {
	certs, err := c.client.Certificate.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range certs {
		if certs[i].Identifier == identifier {
			return &certs[i], nil
		}
	}

	return nil, nil
}
