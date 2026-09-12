package manageddb

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
	_ resource.Resource                = &managedDBResource{}
	_ resource.ResourceWithConfigure   = &managedDBResource{}
	_ resource.ResourceWithImportState = &managedDBResource{}
)

type managedDBResource struct {
	client *govpsie.Client
}

type managedDBResourceModel struct {
	Identifier           types.String `tfsdk:"identifier"`
	Name                 types.String `tfsdk:"name"`
	DBType               types.String `tfsdk:"db_type"`
	DatacenterIdentifier types.String `tfsdk:"datacenter_identifier"`
	PlanID               types.Int64  `tfsdk:"plan_id"`
	ProjectID            types.String `tfsdk:"project_id"`
	NodeCount            types.Int64  `tfsdk:"node_count"`
	ClusterName          types.String `tfsdk:"cluster_name"`
	CPU                  types.Int64  `tfsdk:"cpu"`
	RAM                  types.Int64  `tfsdk:"ram"`
	Traffic              types.Int64  `tfsdk:"traffic"`
	AdminPassword        types.String `tfsdk:"admin_password"`
	CreatedOn            types.String `tfsdk:"created_on"`
}

// NewManagedDatabaseResource is a helper function to create the resource.
func NewManagedDatabaseResource() resource.Resource {
	return &managedDBResource{}
}

func (m *managedDBResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database"
}

func (m *managedDBResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VPSie managed database cluster.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the managed database cluster.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the managed database cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_type": schema.StringAttribute{
				MarkdownDescription: "The database engine type (for example `mysql`, `postgresql`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datacenter_identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier of the datacenter that hosts the cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan_id": schema.Int64Attribute{
				MarkdownDescription: "The plan identifier that determines node size.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the project the cluster belongs to.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"node_count": schema.Int64Attribute{
				MarkdownDescription: "The number of nodes in the cluster. Changing this scales the cluster up or down one node at a time.",
				Required:            true,
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "The cluster name as reported by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: "The number of vCPUs per node.",
				Computed:            true,
			},
			"ram": schema.Int64Attribute{
				MarkdownDescription: "The amount of RAM per node.",
				Computed:            true,
			},
			"traffic": schema.Int64Attribute{
				MarkdownDescription: "The traffic allowance for the cluster.",
				Computed:            true,
			},
			"admin_password": schema.StringAttribute{
				MarkdownDescription: "The generated administrator password for the cluster.",
				Computed:            true,
				Sensitive:           true,
			},
			"created_on": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the cluster.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (m *managedDBResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	m.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (m *managedDBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedDBResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &govpsie.CreateManagedDBRequest{
		Name:         plan.Name.ValueString(),
		DBType:       plan.DBType.ValueString(),
		DcIdentifier: plan.DatacenterIdentifier.ValueString(),
		NodeCount:    plan.NodeCount.ValueInt64(),
		PlanID:       plan.PlanID.ValueInt64(),
		ProjectID:    plan.ProjectID.ValueString(),
	}

	if err := m.client.ManagedDB.Create(ctx, createReq); err != nil {
		resp.Diagnostics.AddError(
			"Error creating managed database",
			"couldn't create managed database, unexpected error: "+err.Error(),
		)

		return
	}

	summary, err := m.getClusterByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating managed database",
			"managed database created but couldn't resolve its identifier: "+err.Error(),
		)

		return
	}

	plan.Identifier = types.StringValue(summary.Identifier)
	details, err := m.client.ManagedDB.Get(ctx, summary.Identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating managed database",
			"managed database created but couldn't read its details: "+err.Error(),
		)

		return
	}

	applyManagedDBDetails(&plan, details)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (m *managedDBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedDBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Confirm the cluster still exists via the list endpoint first, so an
	// externally deleted cluster is removed from state instead of surfacing a
	// hard error from the detail endpoint.
	exists, err := m.clusterExists(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading managed database",
			"couldn't list managed databases: "+err.Error(),
		)

		return
	}

	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}

	details, err := m.client.ManagedDB.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading managed database",
			"couldn't read managed database "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}

	if details == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	applyManagedDBDetails(&state, details)
	state.NodeCount = types.Int64Value(details.NodesCount)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (m *managedDBResource) clusterExists(ctx context.Context, identifier string) (bool, error) {
	clusters, err := m.client.ManagedDB.List(ctx)
	if err != nil {
		return false, err
	}

	for i := range clusters {
		if clusters[i].Identifier == identifier {
			return true, nil
		}
	}

	return false, nil
}

// Update scales the cluster node count up or down; other attributes force replacement.
func (m *managedDBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state managedDBResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()
	current := state.NodeCount.ValueInt64()
	desired := plan.NodeCount.ValueInt64()

	for current < desired {
		if err := m.client.ManagedDB.AddNode(ctx, identifier); err != nil {
			resp.Diagnostics.AddError(
				"Error scaling managed database",
				"couldn't add a node to managed database "+identifier+": "+err.Error(),
			)

			return
		}
		current++
	}

	for current > desired {
		if err := m.client.ManagedDB.ReduceNode(ctx, identifier); err != nil {
			resp.Diagnostics.AddError(
				"Error scaling managed database",
				"couldn't remove a node from managed database "+identifier+": "+err.Error(),
			)

			return
		}
		current--
	}

	details, err := m.client.ManagedDB.Get(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error scaling managed database",
			"managed database scaled but couldn't read its details: "+err.Error(),
		)

		return
	}

	plan.Identifier = state.Identifier
	applyManagedDBDetails(&plan, details)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (m *managedDBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state managedDBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := m.client.ManagedDB.Delete(ctx, state.Identifier.ValueString(), "Deleted via Terraform")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting managed database",
			"couldn't delete managed database "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}
}

func (m *managedDBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// applyManagedDBDetails maps API detail fields onto the model. It deliberately
// does not touch node_count, which is a required, config-driven attribute:
// overwriting it with the API's (possibly still-provisioning) node count would
// produce "inconsistent result after apply" errors. Read updates node_count
// separately for drift detection.
func applyManagedDBDetails(model *managedDBResourceModel, details *govpsie.ManagedDBDetails) {
	if details.Identifier != "" {
		model.Identifier = types.StringValue(details.Identifier)
	}
	model.ClusterName = types.StringValue(details.ClusterName)
	model.CPU = types.Int64Value(details.CPU)
	model.RAM = types.Int64Value(details.RAM)
	model.Traffic = types.Int64Value(details.Traffic)
	model.AdminPassword = types.StringValue(details.AdminPassword)
	model.CreatedOn = types.StringValue(details.CreatedOn)
}

func (m *managedDBResource) getClusterByName(ctx context.Context, name string) (*govpsie.ManagedDB, error) {
	clusters, err := m.client.ManagedDB.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range clusters {
		if clusters[i].ClusterName == name || clusters[i].Nickname == name {
			return &clusters[i], nil
		}
	}

	return nil, fmt.Errorf("managed database %q not found", name)
}
