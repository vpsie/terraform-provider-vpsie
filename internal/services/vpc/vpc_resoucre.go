package vpc

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
	_ resource.Resource                = &vpcResource{}
	_ resource.ResourceWithConfigure   = &vpcResource{}
	_ resource.ResourceWithImportState = &vpcResource{}
)

type vpcResource struct {
	client *govpsie.Client
}

type vpcResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	UserID           types.Int64  `tfsdk:"user_id"`
	OwnerID          types.Int64  `tfsdk:"owner_id"`
	DatacenterID     types.Int64  `tfsdk:"datacenter_id"`
	Description      types.String `tfsdk:"description"`
	InterfaceNumber  types.Int64  `tfsdk:"interface_number"`
	NetworkTagNumber types.Int64  `tfsdk:"network_tag_number"`
	NetworkRange     types.String `tfsdk:"network_range"`
	NetworkSize      types.Int64  `tfsdk:"network_size"`
	IsDefault        types.Int64  `tfsdk:"is_default"`
	CreatedBy        types.Int64  `tfsdk:"created_by"`
	UpdatedBy        types.Int64  `tfsdk:"updated_by"`
	CreatedOn        types.String `tfsdk:"created_on"`
	LastUpdated      types.String `tfsdk:"last_updated"`
	LowIPNum         types.Int64  `tfsdk:"low_ip_num"`
	HightIPNum       types.Int64  `tfsdk:"hight_ip_num"`
	IsUpcNetwork     types.Int64  `tfsdk:"is_upc_network"`
	Firstname        types.String `tfsdk:"firstname"`
	Lastname         types.String `tfsdk:"lastname"`
	Username         types.String `tfsdk:"username"`
	State            types.String `tfsdk:"state"`
	DcName           types.String `tfsdk:"dc_name"`
	DcIdentifier     types.String `tfsdk:"dc_identifier"`
}

func NewVpcResource() resource.Resource {
	return &vpcResource{}
}

func (v *vpcResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (v *vpcResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"dc_identifier": schema.StringAttribute{
				Required: true,
			},
			"auto_generate": schema.Int64Attribute{
				Required: true,
			},
			"network_size": schema.StringAttribute{
				Required: true,
			},
			"network_range": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"created_on": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"firstname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lastname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dc_name": schema.StringAttribute{
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
			"owner_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"datacenter_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"interface_number": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"network_tag_number": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"low_ip_num": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"hight_ip_num": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_upc_network": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_default": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"updated_by": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (v *vpcResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	v.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (v *vpcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := v.client.VPC.CreateVpc(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating VPC", err.Error())
		return
	}

	vpc, err := v.GetVpcByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting VPC by name", err.Error())
		return
	}

	plan.ID = types.Int64Value(int64(vpc.ID))
	plan.NetworkRange = types.StringValue(vpc.NetworkRange)
	plan.CreatedOn = types.StringValue(vpc.CreatedOn.String())
	plan.LastUpdated = types.StringValue(vpc.LastUpdated.String())
	plan.Firstname = types.StringValue(vpc.Firstname)
	plan.Lastname = types.StringValue(vpc.Lastname)
	plan.Username = types.StringValue(vpc.Username)
	plan.State = types.StringValue(vpc.State)
	plan.DcName = types.StringValue(vpc.DcName)
	plan.DcIdentifier = types.StringValue(vpc.DcIdentifier)
	plan.UserID = types.Int64Value(int64(vpc.UserID))
	plan.OwnerID = types.Int64Value(int64(vpc.OwnerID))
	plan.DatacenterID = types.Int64Value(int64(vpc.DatacenterID))
	plan.InterfaceNumber = types.Int64Value(int64(vpc.InterfaceNumber))
	plan.NetworkTagNumber = types.Int64Value(int64(vpc.NetworkTagNumber))
	plan.NetworkSize = types.Int64Value(int64(vpc.NetworkSize))
	plan.LowIPNum = types.Int64Value(int64(vpc.LowIPNum))
	plan.HightIPNum = types.Int64Value(int64(vpc.HightIPNum))
	plan.IsUpcNetwork = types.Int64Value(int64(vpc.IsUpcNetwork))
	plan.IsDefault = types.Int64Value(int64(vpc.IsDefault))
	plan.CreatedBy = types.Int64Value(int64(vpc.CreatedBy))
	plan.UpdatedBy = types.Int64Value(int64(vpc.UpdatedBy))
	plan.Description = types.StringValue(vpc.Description)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (v *vpcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := v.client.VPC.Get(ctx, state.ID.String())
	if err != nil {
		resp.Diagnostics.AddError("Error getting VPC", err.Error())
		return
	}

	state.Name = types.StringValue(vpc.Name)
	state.Description = types.StringValue(vpc.Description)
	state.NetworkRange = types.StringValue(vpc.NetworkRange)
	state.CreatedOn = types.StringValue(vpc.CreatedOn.String())
	state.LastUpdated = types.StringValue(vpc.LastUpdated.String())
	state.Firstname = types.StringValue(vpc.Firstname)
	state.Lastname = types.StringValue(vpc.Lastname)
	state.Username = types.StringValue(vpc.Username)
	state.State = types.StringValue(vpc.State)
	state.DcName = types.StringValue(vpc.DcName)
	state.DcIdentifier = types.StringValue(vpc.DcIdentifier)
	state.UserID = types.Int64Value(int64(vpc.UserID))
	state.OwnerID = types.Int64Value(int64(vpc.OwnerID))
	state.DatacenterID = types.Int64Value(int64(vpc.DatacenterID))
	state.InterfaceNumber = types.Int64Value(int64(vpc.InterfaceNumber))
	state.NetworkTagNumber = types.Int64Value(int64(vpc.NetworkTagNumber))
	state.NetworkSize = types.Int64Value(int64(vpc.NetworkSize))
	state.LowIPNum = types.Int64Value(int64(vpc.LowIPNum))
	state.HightIPNum = types.Int64Value(int64(vpc.HightIPNum))
	state.IsUpcNetwork = types.Int64Value(int64(vpc.IsUpcNetwork))
	state.IsDefault = types.Int64Value(int64(vpc.IsDefault))
	state.CreatedBy = types.Int64Value(int64(vpc.CreatedBy))
	state.UpdatedBy = types.Int64Value(int64(vpc.UpdatedBy))
	state.Description = types.StringValue(vpc.Description)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (v *vpcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (v *vpcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := v.client.VPC.DeleteVpc(ctx, state.ID.String(), "terraform-provider", "terraform-provider")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting vpc",
			"couldn't delete vpc, unexpected error: "+err.Error(),
		)

		return
	}
}

func (v *vpcResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (v *vpcResource) GetVpcByName(ctx context.Context, name string) (*govpsie.VPC, error) {
	return nil, nil
}
