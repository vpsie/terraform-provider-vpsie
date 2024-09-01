package kubernetes

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
	_ resource.Resource                = &kubernetesGroupResource{}
	_ resource.ResourceWithConfigure   = &kubernetesGroupResource{}
	_ resource.ResourceWithImportState = &kubernetesGroupResource{}
)

type kubernetesGroupResource struct {
	client *govpsie.Client
}

type kubernetesGroupResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	GroupName         types.String `tfsdk:"group_name"`
	UserID            types.Int64  `tfsdk:"user_id"`
	BoxsizeID         types.Int64  `tfsdk:"boxsize_id"`
	DatacenterID      types.Int64  `tfsdk:"datacenter_id"`
	RAM               types.Int64  `tfsdk:"ram"`
	CPU               types.Int64  `tfsdk:"cpu"`
	Ssd               types.Int64  `tfsdk:"ssd"`
	Traffic           types.Int64  `tfsdk:"traffic"`
	Notes             types.String `tfsdk:"notes"`
	CreatedOn         types.String `tfsdk:"created_on"`
	LastUpdated       types.String `tfsdk:"last_updated"`
	DroppedOn         types.String `tfsdk:"dropped_on"`
	IsActive          types.Int64  `tfsdk:"is_active"`
	IsDeleted         types.Int64  `tfsdk:"is_deleted"`
	Identifier        types.String `tfsdk:"identifier"`
	ProjectID         types.Int64  `tfsdk:"project_id"`
	ClusterID         types.Int64  `tfsdk:"cluster_id"`
	NodesCount        types.Int64  `tfsdk:"nodes_count"`
	DcIdentifier      types.String `tfsdk:"dc_identifier"`
	ClusterIdentifier types.String `tfsdk:"cluster_identifier"`
}

// NewKubernetesGroupDataSource is a helper function to create the data source.
func NewKubernetesGroupResource() resource.Resource {
	return &kubernetesGroupResource{}
}

func (k *kubernetesGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_group"
}

// Schema defines the schema for the resource.
func (k *kubernetesGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"group_name": schema.StringAttribute{
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
			"boxsize_id": schema.Int64Attribute{
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
			"ram": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ssd": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"traffic": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"notes": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"dropped_on": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_active": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_deleted": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nodes_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dc_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_identifier": schema.StringAttribute{
				Required: true,
			},
		},
	}

}

func (k *kubernetesGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	k.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (k *kubernetesGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kubernetesGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := govpsie.CreateK8sGroupReq{
		ClusterIdentifier: plan.ClusterIdentifier.ValueString(),
		GroupName:         plan.GroupName.ValueString(),
		KubeSizeID:        2,
	}

	err := k.client.K8s.CreateK8sGroup(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating kubernetes", err.Error())
		return
	}

	k8sGroup, err := k.GetKubernetesGroupByName(ctx, plan.GroupName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading kubernetes", err.Error())
		return
	}

	plan.GroupName = types.StringValue(k8sGroup.GroupName)
	plan.ID = types.Int64Value(k8sGroup.ID)
	plan.UserID = types.Int64Value(k8sGroup.UserID)
	plan.BoxsizeID = types.Int64Value(k8sGroup.BoxsizeID)
	plan.DatacenterID = types.Int64Value(k8sGroup.DatacenterID)
	plan.RAM = types.Int64Value(k8sGroup.RAM)
	plan.CPU = types.Int64Value(k8sGroup.CPU)
	plan.Ssd = types.Int64Value(k8sGroup.Ssd)
	plan.Traffic = types.Int64Value(k8sGroup.Traffic)
	plan.Notes = types.StringValue(k8sGroup.Notes)
	plan.CreatedOn = types.StringValue(k8sGroup.CreatedOn.String())
	plan.LastUpdated = types.StringValue(k8sGroup.LastUpdated.String())
	plan.DroppedOn = types.StringValue(k8sGroup.DroppedOn.String())
	plan.IsActive = types.Int64Value(k8sGroup.IsActive)
	plan.IsDeleted = types.Int64Value(k8sGroup.IsDeleted)
	plan.Identifier = types.StringValue(k8sGroup.Identifier)
	plan.ProjectID = types.Int64Value(k8sGroup.ProjectID)
	plan.ClusterID = types.Int64Value(k8sGroup.ClusterID)
	plan.NodesCount = types.Int64Value(k8sGroup.NodesCount)
	plan.DcIdentifier = types.StringValue(k8sGroup.DcIdentifier)

}

// Read refreshes the Terraform state with the latest data.
func (k *kubernetesGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state kubernetesGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	k8sGroup, err := k.GetKubernetesGroupByIdentifier(ctx, state.Identifier.ValueString())
	if err != nil {
		if err.Error() == "kubernetes group not found" {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading vpsie kubernetes group",
			"Could't read vpsie kubernetes group identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.GroupName = types.StringValue(k8sGroup.GroupName)
	state.ID = types.Int64Value(k8sGroup.ID)
	state.UserID = types.Int64Value(k8sGroup.UserID)
	state.BoxsizeID = types.Int64Value(k8sGroup.BoxsizeID)
	state.DatacenterID = types.Int64Value(k8sGroup.DatacenterID)
	state.RAM = types.Int64Value(k8sGroup.RAM)
	state.CPU = types.Int64Value(k8sGroup.CPU)
	state.Ssd = types.Int64Value(k8sGroup.Ssd)
	state.Traffic = types.Int64Value(k8sGroup.Traffic)
	state.Notes = types.StringValue(k8sGroup.Notes)
	state.CreatedOn = types.StringValue(k8sGroup.CreatedOn.String())
	state.LastUpdated = types.StringValue(k8sGroup.LastUpdated.String())
	state.DroppedOn = types.StringValue(k8sGroup.DroppedOn.String())
	state.IsActive = types.Int64Value(k8sGroup.IsActive)
	state.IsDeleted = types.Int64Value(k8sGroup.IsDeleted)
	state.Identifier = types.StringValue(k8sGroup.Identifier)
	state.ProjectID = types.Int64Value(k8sGroup.ProjectID)
	state.ClusterID = types.Int64Value(k8sGroup.ClusterID)
	state.NodesCount = types.Int64Value(k8sGroup.NodesCount)
	state.DcIdentifier = types.StringValue(k8sGroup.DcIdentifier)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (k *kubernetesGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state kubernetesGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan kubernetesGroupResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.NodesCount.ValueInt64() > plan.NodesCount.ValueInt64() {
		for i := plan.NodesCount.ValueInt64(); i < state.NodesCount.ValueInt64(); i++ {
			err := k.client.K8s.AddNode(ctx, state.ClusterIdentifier.ValueString(), "slave", int(state.ID.ValueInt64()))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating kubernetes",
					"Couldn't update kubernetes, unexpected error: "+err.Error(),
				)

				return
			}
		}

	} else if state.NodesCount.ValueInt64() < plan.NodesCount.ValueInt64() {
		for i := state.NodesCount.ValueInt64(); i < plan.NodesCount.ValueInt64(); i++ {
			err := k.client.K8s.RemoveNode(ctx, state.ClusterIdentifier.ValueString(), "slave", int(state.ID.ValueInt64()))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating kubernetes",
					"Couldn't update kubernetes, unexpected error: "+err.Error(),
				)

				return
			}
		}

	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (k *kubernetesGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kubernetesGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := k.client.K8s.DeleteK8sGroup(ctx, state.Identifier.ValueString(), "terraform", "terraform")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting kubernetes group",
			"Couldn't delete kubernetes group, unexpected error: "+err.Error(),
		)

		return
	}
}

func (k *kubernetesGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (k *kubernetesGroupResource) GetKubernetesGroupByName(ctx context.Context, name string) (*govpsie.K8sGroup, error) {
	k8s, err := k.client.K8s.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, k8 := range k8s {
		groups, err := k.client.K8s.ListK8sGroups(ctx, k8.Identifier)
		if err != nil {
			return nil, err
		}

		for _, group := range groups {
			if group.GroupName == name {
				return &group, nil
			}
		}

	}

	return nil, fmt.Errorf("KUBERNETES GROUP NOT FOUND: %s", name)
}

func (k *kubernetesGroupResource) GetKubernetesGroupByIdentifier(ctx context.Context, identifier string) (*govpsie.K8sGroup, error) {

	k8s, err := k.client.K8s.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, k8 := range k8s {
		groups, err := k.client.K8s.ListK8sGroups(ctx, k8.Identifier)
		if err != nil {
			return nil, err
		}

		for _, group := range groups {
			if group.Identifier == identifier {
				return &group, nil
			}
		}

	}

	return nil, fmt.Errorf("KUBERNETES GROUP NOT FOUND: %s", identifier)
}
