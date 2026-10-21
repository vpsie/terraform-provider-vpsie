package baremetal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &bareMetalPlansDataSource{}
	_ datasource.DataSourceWithConfigure = &bareMetalPlansDataSource{}
)

type bareMetalPlansDataSource struct {
	client *govpsie.Client
}

type bareMetalPlansDataSourceModel struct {
	ID           types.String         `tfsdk:"id"`
	DcIdentifier types.String         `tfsdk:"dc_identifier"`
	MinRAMGB     types.Int64          `tfsdk:"min_ram_gb"`
	MinDiskGB    types.Int64          `tfsdk:"min_disk_gb"`
	GPU          types.Bool           `tfsdk:"gpu"`
	Plans        []bareMetalPlanModel `tfsdk:"plans"`
}

type bareMetalPlanModel struct {
	Identifier       types.String   `tfsdk:"identifier"`
	Name             types.String   `tfsdk:"name"`
	ProcessorModel   types.String   `tfsdk:"processor_model"`
	CPUCount         types.Int64    `tfsdk:"cpu_count"`
	RAMGB            types.Int64    `tfsdk:"ram_gb"`
	DiskGB           types.Int64    `tfsdk:"disk_gb"`
	DiskTypeSummary  types.String   `tfsdk:"disk_type_summary"`
	GPUModel         types.String   `tfsdk:"gpu_model"`
	GPUCount         types.Int64    `tfsdk:"gpu_count"`
	GPUMemoryGB      types.Int64    `tfsdk:"gpu_memory_gb"`
	Price            types.String   `tfsdk:"price"`
	Currency         types.String   `tfsdk:"currency"`
	SetupFee         types.String   `tfsdk:"setup_fee"`
	Delivery         types.String   `tfsdk:"delivery"`
	DeliveryHours    types.Int64    `tfsdk:"delivery_hours"`
	ContractTerms    []types.String `tfsdk:"contract_terms"`
	LinkSpeedOptions []types.Int64  `tfsdk:"link_speed_options"`
	AvailableServers types.Int64    `tfsdk:"available_servers"`
}

// NewBareMetalPlansDataSource returns the vpsie_bare_metal_plans data source.
func NewBareMetalPlansDataSource() datasource.DataSource {
	return &bareMetalPlansDataSource{}
}

func (d *bareMetalPlansDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bare_metal_plans"
}

func (d *bareMetalPlansDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	computedString := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, MarkdownDescription: desc}
	}
	computedInt := func(desc string) schema.Int64Attribute {
		return schema.Int64Attribute{Computed: true, MarkdownDescription: desc}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the bare-metal hardware plans a datacenter offers. A datacenter that sells no bare metal returns an empty list.",
		Attributes: map[string]schema.Attribute{
			"id": computedString("The datacenter identifier."),
			"dc_identifier": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identifier of the datacenter.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
			},
			"min_ram_gb": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Only plans with at least this much RAM, in GB.",
				Validators:          []validator.Int64{int64validator.AtLeast(1)},
			},
			"min_disk_gb": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Only plans with at least this much disk, in GB.",
				Validators:          []validator.Int64{int64validator.AtLeast(1)},
			},
			"gpu": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "`true` for plans with GPUs only, `false` for plans without.",
			},
			"plans": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The matching plans.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":        computedString("Plan identifier, for `vpsie_bare_metal.plan_identifier`."),
						"name":              computedString("Plan name."),
						"processor_model":   computedString("Processor model."),
						"cpu_count":         computedInt("Number of CPU cores."),
						"ram_gb":            computedInt("RAM in GB."),
						"disk_gb":           computedInt("Total disk in GB."),
						"disk_type_summary": computedString("Disk layout, for example `2 x 960 GB NVMe`."),
						"gpu_model":         computedString("GPU model, empty without GPUs."),
						"gpu_count":         computedInt("Number of GPUs."),
						"gpu_memory_gb":     computedInt("Memory per GPU in GB."),
						"price":             computedString("Monthly price, as a decimal string."),
						"currency":          computedString("Currency of the price."),
						"setup_fee":         computedString("One-time setup fee, as a decimal string."),
						"delivery":          computedString("`instant` (in stock) or `build_to_order`."),
						"delivery_hours":    computedInt("Hours a build-to-order server takes to deliver."),
						"contract_terms": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Billing terms the plan can be ordered on (`hourly`, `monthly`).",
						},
						"link_speed_options": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.Int64Type,
							MarkdownDescription: "Uplink speeds in Mbit/s the plan can be ordered with.",
						},
						"available_servers": computedInt("Servers in stock for the plan."),
					},
				},
			},
		},
	}
}

func (d *bareMetalPlansDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *bareMetalPlansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bareMetalPlansDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := &govpsie.BareMetalPlanFilter{
		MinRAMGB:  int(state.MinRAMGB.ValueInt64()),
		MinDiskGB: int(state.MinDiskGB.ValueInt64()),
		GPU:       state.GPU.ValueBoolPointer(),
	}
	plans, err := d.client.BareMetal.ListPlans(ctx, state.DcIdentifier.ValueString(), filter)
	if err != nil {
		resp.Diagnostics.AddError("Error reading bare-metal plans", err.Error())
		return
	}

	state.Plans = make([]bareMetalPlanModel, 0, len(plans))
	for _, p := range plans {
		terms := make([]types.String, 0, len(p.ContractTerms))
		for _, t := range p.ContractTerms {
			terms = append(terms, types.StringValue(t.Term))
		}
		speeds := make([]types.Int64, 0, len(p.LinkSpeedOptions))
		for _, s := range p.LinkSpeedOptions {
			speeds = append(speeds, types.Int64Value(int64(s)))
		}

		state.Plans = append(state.Plans, bareMetalPlanModel{
			Identifier:       types.StringValue(p.Identifier),
			Name:             types.StringValue(p.Name),
			ProcessorModel:   types.StringValue(p.ProcessorModel),
			CPUCount:         types.Int64Value(int64(p.CPUCount)),
			RAMGB:            types.Int64Value(int64(p.RAMGB)),
			DiskGB:           types.Int64Value(int64(p.DiskGB)),
			DiskTypeSummary:  types.StringValue(p.DiskTypeSummary),
			GPUModel:         types.StringValue(p.GPUModel),
			GPUCount:         types.Int64Value(int64(p.GPUCount)),
			GPUMemoryGB:      types.Int64Value(int64(p.GPUMemoryGB)),
			Price:            types.StringValue(p.Price.String()),
			Currency:         types.StringValue(p.Currency),
			SetupFee:         types.StringValue(p.SetupFee.String()),
			Delivery:         types.StringValue(p.Delivery),
			DeliveryHours:    types.Int64Value(int64(p.DeliveryHours)),
			ContractTerms:    terms,
			LinkSpeedOptions: speeds,
			AvailableServers: types.Int64Value(int64(p.AvailableServers)),
		})
	}
	state.ID = state.DcIdentifier

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
