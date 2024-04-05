package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	_ resource.Resource                = &kubernetesResource{}
	_ resource.ResourceWithConfigure   = &kubernetesResource{}
	_ resource.ResourceWithImportState = &kubernetesResource{}
)

type kubernetesResource struct {
	client *govpsie.Client
}

type kubernetesResourceModel struct {
	ClusterName types.String  `tfsdk:"cluster_name"`
	Identifier  types.String  `tfsdk:"identifier"`
	Count       types.Int64   `tfsdk:"count"`
	CreatedOn   types.String  `tfsdk:"created_on"`
	UpdatedOn   types.String  `tfsdk:"updated_on"`
	CreatedBy   types.String  `tfsdk:"created_by"`
	NickName    types.String  `tfsdk:"nickname"`
	Cpu         types.Int64   `tfsdk:"cpu"`
	Ram         types.Int64   `tfsdk:"ram"`
	Traffic     types.Int64   `tfsdk:"traffic"`
	Color       types.String  `tfsdk:"color"`
	Price       types.Float64 `tfsdk:"price"`
	Nodes       []Node        `tfsdk:"nodes"`

	DcIdentifier       types.String   `tfsdk:"dc_identifier"`
	ResourceIdentifier types.String   `tfsdk:"resource_identifier"`
	ManagerCount       types.Int64    `tfsdk:"manager_count"`
	SlaveCount         types.Int64    `tfsdk:"slave_count"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	VpcId              types.Int64    `tfsdk:"vpc_id"`
	KuberVer           types.Int64    `tfsdk:"kuber_ver"`
	ProjectIdentifier  types.String   `tfsdk:"project_identifier"`
}

type Node struct {
	Id           types.Int64  `tfsdk:"id"`
	UserId       types.Int64  `tfsdk:"user_id"`
	HostName     types.String `tfsdk:"hostname"`
	DefaultIP    types.String `tfsdk:"default_ip"`
	PrivateIP    types.String `tfsdk:"private_ip"`
	NodeType     types.Int64  `tfsdk:"node_type"`
	NodeId       types.Int64  `tfsdk:"node_id"`
	DatacenterId types.Int64  `tfsdk:"datacenter_id"`
	CreatedOn    types.String `tfsdk:"created_on"`
}

func NewKubernetesResource() resource.Resource {
	return &kubernetesResource{}
}

func (k *kubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes"
}

func (k *kubernetesResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"cluster_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"count": schema.Int64Attribute{
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
			"updated_on": schema.StringAttribute{
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
			"nickname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.Int64Attribute{
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
			"traffic": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"price": schema.Float64Attribute{
				Computed: true,
			},
			"nodes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"hostname": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"default_ip": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"private_ip": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"node_type": schema.Int64Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"node_id": schema.Int64Attribute{
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
						"created_on": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),

			"dc_identifier": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_identifier": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_identifier": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"manager_count": schema.Int64Attribute{
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"slave_count": schema.Int64Attribute{
				Optional: true,
			},
			"vpc_id": schema.Int64Attribute{
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (k *kubernetesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (k *kubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kubernetesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := govpsie.CreateK8sReq{
		ClusterName:        plan.ClusterName.ValueString(),
		DcIdentifier:       plan.DcIdentifier.ValueString(),
		NodesCountMaster:   int(plan.ManagerCount.ValueInt64()),
		NodesCountSlave:    int(plan.SlaveCount.ValueInt64()),
		VpcId:              int(plan.VpcId.ValueInt64()),
		KuberVer:           int(plan.KuberVer.ValueInt64()),
		ResourceIdentifier: plan.ResourceIdentifier.ValueString(),
		ProjectIdentifier:  plan.ProjectIdentifier.ValueString(),
	}

	err := k.client.K8s.Create(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating kubernetes", err.Error())
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 20*time.Minute)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	for {
		// Check if the context has expired
		if ctx.Err() != nil {
			resp.Diagnostics.AddError("error waiting for resource to become ready", ctx.Err().Error())
			return
		}

		// Check resource status here using provider's API or SDK
		k8s, ready, err := k.checkResourceStatus(ctx, plan.ClusterName.ValueString())
		if err != nil {
			//  return fmt.Errorf("Timeout waiting for resource to become ready")
			resp.Diagnostics.AddError("Error cheking status of resource", err.Error())
			return
		}

		if ready {
			plan.Identifier = types.StringValue(k8s.Identifier)
			plan.ClusterName = types.StringValue(k8s.ClusterName)
			plan.Color = types.StringValue(k8s.Color)
			plan.Count = types.Int64Value(int64(k8s.Count))
			plan.CreatedOn = types.StringValue(k8s.CreatedOn)
			plan.UpdatedOn = types.StringValue(k8s.UpdatedOn)
			plan.CreatedBy = types.StringValue(k8s.CreatedBy)
			plan.NickName = types.StringValue(k8s.NickName)
			plan.Cpu = types.Int64Value(int64(k8s.Cpu))
			plan.Ram = types.Int64Value(int64(k8s.Ram))
			plan.Traffic = types.Int64Value(int64(k8s.Traffic))
			plan.Price = types.Float64Value(k8s.Price)

			nodes := []Node{}
			for _, node := range k8s.Nodes {
				nodes = append(nodes, Node{
					Id:           types.Int64Value(int64(node.Id)),
					UserId:       types.Int64Value(int64(node.UserId)),
					HostName:     types.StringValue(node.HostName),
					DefaultIP:    types.StringValue(node.DefaultIP),
					PrivateIP:    types.StringValue(node.PrivateIP),
					NodeType:     types.Int64Value(int64(node.NodeType)),
					NodeId:       types.Int64Value(int64(node.NodeId)),
					DatacenterId: types.Int64Value(int64(node.DatacenterId)),
					CreatedOn:    types.StringValue(node.CreatedOn),
				})
			}
			plan.Nodes = nodes

			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			return
		}

		// Wait for a delay before retrying
		time.Sleep(5 * time.Second)
	}

}

// Read refreshes the Terraform state with the latest data.
func (k *kubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state kubernetesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	k8s, err := k.client.K8s.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		if err.Error() == "kubernetes not found" {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading vpsie kubernetes",
			"Could't read vpsie kubernetes identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Identifier = types.StringValue(k8s.Identifier)
	state.Identifier = types.StringValue(k8s.Identifier)
	state.ClusterName = types.StringValue(k8s.ClusterName)
	state.Color = types.StringValue(k8s.Color)
	state.Count = types.Int64Value(int64(k8s.Count))
	state.CreatedOn = types.StringValue(k8s.CreatedOn)
	state.UpdatedOn = types.StringValue(k8s.UpdatedOn)
	state.CreatedBy = types.StringValue(k8s.CreatedBy)
	state.NickName = types.StringValue(k8s.NickName)
	state.Cpu = types.Int64Value(int64(k8s.Cpu))
	state.Ram = types.Int64Value(int64(k8s.Ram))
	state.Traffic = types.Int64Value(int64(k8s.Traffic))
	state.Price = types.Float64Value(k8s.Price)

	nodes := []Node{}
	for _, node := range k8s.Nodes {
		nodes = append(nodes, Node{
			Id:           types.Int64Value(int64(node.Id)),
			UserId:       types.Int64Value(int64(node.UserId)),
			HostName:     types.StringValue(node.HostName),
			DefaultIP:    types.StringValue(node.DefaultIP),
			PrivateIP:    types.StringValue(node.PrivateIP),
			NodeType:     types.Int64Value(int64(node.NodeType)),
			NodeId:       types.Int64Value(int64(node.NodeId)),
			DatacenterId: types.Int64Value(int64(node.DatacenterId)),
			CreatedOn:    types.StringValue(node.CreatedOn),
		})
	}
	state.Nodes = nodes

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (k *kubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state kubernetesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan kubernetesResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SlaveCount.ValueInt64() > plan.SlaveCount.ValueInt64() {
		for i := plan.SlaveCount.ValueInt64(); i < state.SlaveCount.ValueInt64(); i++ {
			err := k.client.K8s.AddSlave(ctx, state.Identifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating kubernetes",
					"Couldn't update kubernetes, unexpected error: "+err.Error(),
				)

				return
			}
		}

	} else if state.SlaveCount.ValueInt64() < plan.SlaveCount.ValueInt64() {
		for i := state.SlaveCount.ValueInt64(); i < plan.SlaveCount.ValueInt64(); i++ {
			err := k.client.K8s.RemoveSlave(ctx, state.Identifier.ValueString())
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
func (k *kubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kubernetesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := k.client.K8s.Delete(ctx, state.Identifier.ValueString(), "terraform", "terraform")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting kubernetes",
			"Couldn't delete kubernetes, unexpected error: "+err.Error(),
		)

		return
	}
}

func (k *kubernetesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (k *kubernetesResource) checkResourceStatus(ctx context.Context, cluster_name string) (*govpsie.K8s, bool, error) {
	kubernetes, err := k.client.K8s.List(ctx, nil)
	if err != nil {
		return nil, false, err
	}

	for _, k8s := range kubernetes {
		if k8s.ClusterName == cluster_name {
			newK8s, err := k.client.K8s.Get(ctx, k8s.Identifier)
			if err != nil {
				return nil, false, err
			}

			return newK8s, true, nil
		}
	}

	return nil, false, nil
}
