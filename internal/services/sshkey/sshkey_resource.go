package sshkey

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
	_ resource.Resource                = &sshkeyResource{}
	_ resource.ResourceWithConfigure   = &sshkeyResource{}
	_ resource.ResourceWithImportState = &sshkeyResource{}
)

type sshkeyResource struct {
	client *govpsie.Client
}

type sshkeyResourceModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Id         types.Int64  `tfsdk:"id"`
	UserId     types.Int64  `tfsdk:"user_id"`
	Name       types.String `tfsdk:"name"`
	PrivateKey types.String `tfsdk:"private_key"`
	CreatedOn  types.String `tfsdk:"created_on"`
	CreatedBy  types.String `tfsdk:"created_by"`
}

func NewSshkeyResource() resource.Resource {
	return &sshkeyResource{}
}

func (s *sshkeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sshkey"
}

func (s *sshkeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"private_key": schema.StringAttribute{
				Required: true,
			},
			"created_on": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (s *sshkeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (s *sshkeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sshkeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.SShKey.Create(ctx, plan.PrivateKey.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating sshkey",
			"Cloudn't create sshkey, unexpected error: "+err.Error(),
		)

		return
	}

	sshkey, err := s.GetSshkeyByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating sshkey",
			"Cloudn't create sshkey, unexpected error: "+err.Error(),
		)

		return
	}

	plan.Name = types.StringValue(sshkey.Name)
	plan.Id = types.Int64Value(sshkey.Id)
	plan.CreatedOn = types.StringValue(sshkey.CreatedOn)
	plan.CreatedBy = types.StringValue(sshkey.CreatedBy)
	plan.UserId = types.Int64Value(sshkey.UserId)
	plan.Identifier = types.StringValue(sshkey.Name)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *sshkeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sshkeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshkey, err := s.client.SShKey.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading vpsie sshkeys",
			"Cloudn't read vpsie sshkeys identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state

	state.Name = types.StringValue(sshkey.Name)
	state.Id = types.Int64Value(sshkey.Id)
	state.CreatedOn = types.StringValue(sshkey.CreatedOn)
	state.CreatedBy = types.StringValue(sshkey.CreatedBy)
	state.PrivateKey = types.StringValue(sshkey.PrivateKey)
	state.UserId = types.Int64Value(sshkey.UserId)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (s *sshkeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (s *sshkeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sshkeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.SShKey.Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting sshkey",
			"Cloudn't delete sshkey, unexpected error: "+err.Error(),
		)

		return
	}
}

func (s *sshkeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (s *sshkeyResource) GetSshkeyByName(ctx context.Context, sshkeyName string) (*govpsie.SShKey, error) {
	sshkeys, err := s.client.SShKey.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, sshkey := range sshkeys {
		if sshkey.Name == sshkeyName {
			return &sshkey, nil
		}
	}

	return nil, fmt.Errorf("sshkey %s not found", sshkeyName)
}
