package domain

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &domainResource{}
	_ resource.ResourceWithConfigure   = &domainResource{}
	_ resource.ResourceWithImportState = &domainResource{}
)

type domainResource struct {
	client *govpsie.Client
}

type domainResourceModel struct {
	Identifier        types.String `tfsdk:"identifier"`
	DomainName        types.String `tfsdk:"domain_name"`
	NsValidated       types.Int64  `tfsdk:"ns_validated"`
	CreatedOn         types.String `tfsdk:"created_on"`
	LastCheck         types.String `tfsdk:"last_check"`
	ProjectIdentifier types.String `tfsdk:"project_identifier"`
}

func NewDomainResource() resource.Resource {
	return &domainResource{}
}

func (s *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (s *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_name": schema.StringAttribute{
				Required: true,
			},
			"ns_validated": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_on": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_check": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_identifier": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (d *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createDomainReq := govpsie.CreateDomainRequest{
		ProjectIdentifier: plan.ProjectIdentifier.ValueString(),
		Domain:            plan.DomainName.ValueString(),
	}

	err := d.client.Domain.CreateDomain(ctx, &createDomainReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain", err.Error())
		return
	}

	domain, err := d.GetDomainByName(ctx, plan.DomainName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting domain", err.Error())
		return
	}

	// Overwrite items with refreshed state

	plan.Identifier = types.StringValue(domain.Identifier)
	plan.DomainName = types.StringValue(domain.DomainName)
	plan.NsValidated = types.Int64Value(int64(domain.NsValidated))
	plan.CreatedOn = types.StringValue(domain.CreatedOn)
	plan.LastCheck = types.StringValue(domain.LastCheck)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := d.GetDomainByName(ctx, state.DomainName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading vpsie domains",
			"couldn't read vpsie domains identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state

	state.DomainName = types.StringValue(domain.DomainName)
	state.CreatedOn = types.StringValue(domain.CreatedOn)
	state.NsValidated = types.Int64Value(int64(domain.NsValidated))
	state.LastCheck = types.StringValue(domain.LastCheck)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (d *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (d *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := d.client.Domain.DeleteDomain(ctx, state.Identifier.ValueString(), "terraform delete", "terraform delete")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting domain",
			"couldn't delete domain, unexpected error: "+err.Error(),
		)

		return
	}
}

func (d *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (d *domainResource) GetDomainByName(ctx context.Context, domainName string) (*govpsie.Domain, error) {
	domains, err := d.client.Domain.ListDomains(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if domainName == domain.DomainName {
			return &domain, nil
		}
	}

	return nil, fmt.Errorf("domain with name %s not found", domainName)
}
