package loadbalancer

import (
	"context"
	"fmt"
	"strconv"
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
	_ resource.Resource                = &loadbalancerResource{}
	_ resource.ResourceWithConfigure   = &loadbalancerResource{}
	_ resource.ResourceWithImportState = &loadbalancerResource{}
)

type loadbalancerResource struct {
	client *govpsie.Client
}

type loadbalancerResourceModel struct {
	LBName     types.String `tfsdk:"lb_name"`
	Identifier types.String `tfsdk:"identifier"`
	Traffic    types.Int64  `tfsdk:"traffic"`
	BoxsizeID  types.Int64  `tfsdk:"boxsize_id"`
	DefaultIP  types.String `tfsdk:"default_ip"`
	DcName     types.String `tfsdk:"dc_name"`
	DcID       types.String `tfsdk:"dcId"`
	CreatedBy  types.String `tfsdk:"created_by"`
	UserID     types.Int64  `tfsdk:"user_id"`
	Rules      []LBRule     `tfsdk:"rules"`

	Algorithm          types.String   `tfsdk:"algorithm"`
	CookieName         types.String   `tfsdk:"cookie_name"`
	HealthCheckPath    types.String   `tfsdk:"health_checkPath"`
	CookieCheck        types.Bool     `tfsdk:"cookie_check"`
	RedirectHTTP       types.Int64    `tfsdk:"redirect_http"`
	ResourceIdentifier types.String   `tfsdk:"resource_identifier"`
	CheckInterval      types.Int64    `tfsdk:"check_interval"`
	FastInterval       types.Int64    `tfsdk:"fast_interval"`
	Rise               types.Int64    `tfsdk:"rise"`
	Fall               types.Int64    `tfsdk:"fall"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

type LBRule struct {
	Scheme    types.String `tfsdk:"scheme"`
	FrontPort types.Int64  `tfsdk:"front_port"`
	BackPort  types.Int64  `tfsdk:"backPort"`
	CreatedOn types.String `tfsdk:"created_on"`
	RuleID    types.String `tfsdk:"rule_id"`
	Domains   []LBDomain   `tfsdk:"domains,omitempty"`
	Backends  []Backend    `tfsdk:"backends,omitempty"`

	DomainName types.String `tfsdk:"domain_ame"`
}

type LBDomain struct {
	DomainID      types.String `tfsdk:"domainId"`
	Backends      []Backend    `tfsdk:"backends"`
	BackPort      types.String `tfsdk:"back_port"`
	BackendScheme types.String `tfsdk:"backend_scheme"`
	DomainName    types.String `tfsdk:"domain_name"`
	Subdomain     types.String `tfsdk:"subdomain,omitempty"`

	Algorithm       types.String `tfsdk:"algorithm"`
	RedirectHTTP    types.Int64  `tfsdk:"redirect_http"`
	HealthCheckPath types.String `tfsdk:"health_check_path"`
	CookieCheck     types.Int64  `tfsdk:"cookie_check"`
	CookieName      types.String `tfsdk:"cookie_name"`
	CreatedOn       types.String `tfsdk:"created_on"`
	CheckInterval   types.Int64  `tfsdk:"check_interval"`
	FastInterval    types.Int64  `tfsdk:"fast_interval"`
	Rise            types.Int64  `tfsdk:"rise"`
	Fall            types.Int64  `tfsdk:"fall"`
}

type Backend struct {
	IP           types.String `tfsdk:"ip"`
	Identifier   types.String `tfsdk:"identifier"`
	VMIdentifier types.String `tfsdk:"vm_identifier,omitempty"`
	CreatedOn    types.String `tfsdk:"created_on"`
}

func NewLoadbalancerResource() resource.Resource {
	return &loadbalancerResource{}
}

func (l *loadbalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

var backendScheme = map[string]schema.Attribute{
	"ip": schema.StringAttribute{
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
	"vm_identifier": schema.StringAttribute{
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
}

func (l *loadbalancerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"lb_name": schema.StringAttribute{
				Required: true,
			},
			"traffic": schema.Int64Attribute{
				Required: true,
			},
			"boxsize_id": schema.Int64Attribute{
				Required: true,
			},
			"default_ip": schema.StringAttribute{
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
			"dcId": schema.StringAttribute{
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
			"user_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"algorithm": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"redirect_http": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"health_check_path": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cookie_check": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cookie_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"check_interval": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"fast_interval": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"rise": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"fall": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"rule_id": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"front_port": schema.Int64Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"back_port": schema.Int64Attribute{
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
						"scheme": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"domains": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"domain_id": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"backend_scheme": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"subdomain": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"algorithm": schema.StringAttribute{
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
									"back_port": schema.Int64Attribute{
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"domain_name": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"redirect_http": schema.Int64Attribute{
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"health_check_path": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"cookie_check": schema.Int64Attribute{
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"cookie_name": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"check_interval": schema.Int64Attribute{
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"fast_interval": schema.Int64Attribute{
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"rise": schema.Int64Attribute{
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"fall": schema.Int64Attribute{
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"backends": schema.ListNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: backendScheme,
										},
									},
								},
							},
						},
						"backends": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: backendScheme,
							},
						},
					},
				},
			},
			"timeout": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (l *loadbalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	l.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (l *loadbalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadbalancerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rules []govpsie.Rule
	for _, rule := range plan.Rules {
		var newRule = govpsie.Rule{}

		newRule.BackPort = rule.BackPort.String()
		newRule.DomainName = rule.DomainName.ValueString()
		newRule.FrontPort = rule.FrontPort.String()
		newRule.Scheme = rule.Scheme.ValueString()

		domains := []govpsie.LBDomain{}
		for _, dns := range rule.Domains {
			domain := govpsie.LBDomain{}
			domain.DomainName = dns.DomainName.ValueString()
			domain.BackPort = dns.BackPort.String()
			domain.BackendScheme = dns.BackendScheme.ValueString()
			domain.BackPort = dns.BackPort.ValueString()

			dnsBackends := []govpsie.Backend{}
			for _, backend := range dns.Backends {
				dnsBackends = append(dnsBackends, govpsie.Backend{
					Ip:           backend.IP.ValueString(),
					VmIdentifier: backend.VMIdentifier.ValueString(),
				})
			}
			domain.Backends = dnsBackends
			domains = append(domains, domain)
		}

		newRule.Domains = domains

		backends := []govpsie.Backend{}
		for _, backend := range rule.Backends {
			backends = append(backends, govpsie.Backend{
				Ip:           backend.IP.ValueString(),
				VmIdentifier: backend.VMIdentifier.ValueString(),
			})
		}
		newRule.Backends = backends

		rules = append(rules, newRule)
	}

	createLb := &govpsie.CreateLBReq{
		LBName:             plan.LBName.ValueString(),
		Algorithm:          plan.Algorithm.ValueString(),
		CookieName:         plan.CookieName.ValueString(),
		CookieCheck:        plan.CookieCheck.ValueBool(),
		RedirectHTTP:       int(plan.RedirectHTTP.ValueInt64()),
		ResourceIdentifier: plan.ResourceIdentifier.ValueString(),
		DcIdentifier:       plan.DcID.ValueString(),
		Rule:               rules,
	}

	err := l.client.LB.CreateLB(ctx, createLb)
	if err != nil {
		resp.Diagnostics.AddError("Error creating image", err.Error())
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
		lb, ready, err := l.checkResourceStatus(ctx, plan.LBName.ValueString())
		if err != nil {
			//  return fmt.Errorf("Timeout waiting for resource to become ready")
			resp.Diagnostics.AddError("Error cheking status of resource", err.Error())
			return
		}

		if ready {
			plan.Identifier = types.StringValue(lb.Identifier)
			plan.UserID = types.Int64Value(int64(lb.UserID))
			plan.DcName = types.StringValue(lb.DcName)
			plan.CreatedBy = types.StringValue(lb.CreatedBy)
			plan.LBName = types.StringValue(lb.LBName)
			plan.Traffic = types.Int64Value(int64(lb.Traffic))
			plan.BoxsizeID = types.Int64Value(int64(lb.BoxsizeID))
			plan.DefaultIP = types.StringValue(lb.DefaultIP)
			plan.DcName = types.StringValue(lb.DcName)
			plan.DcID = types.StringValue(lb.DcID)
			plan.CreatedBy = types.StringValue(lb.CreatedBy)
			plan.UserID = types.Int64Value(int64(lb.UserID))

			var rules = []LBRule{}
			for _, rule := range lb.Rules {
				var newRule = LBRule{}
				newRule.BackPort = types.Int64Value(int64(rule.BackPort))
				newRule.FrontPort = types.Int64Value(int64(rule.FrontPort))
				newRule.Scheme = types.StringValue(rule.Scheme)
				newRule.CreatedOn = types.StringValue(rule.CreatedOn.String())
				newRule.RuleID = types.StringValue(rule.RuleID)

				domains := []LBDomain{}
				for _, dns := range rule.Domains {
					domain := LBDomain{}
					domain.DomainName = types.StringValue(dns.DomainName)
					domain.CreatedOn = types.StringValue(dns.CreatedOn.String())

					var subdomain types.String
					if dns.Subdomain != nil && *dns.Subdomain != "" {
						subdomain = types.StringValue(*dns.Subdomain)
					}

					domain.Subdomain = subdomain
					domain.BackendScheme = types.StringValue(dns.BackendScheme)
					domain.Algorithm = types.StringValue(dns.Algorithm)
					domain.RedirectHTTP = types.Int64Value(int64(dns.RedirectHTTP))
					domain.HealthCheckPath = types.StringValue(dns.HealthCheckPath)
					domain.CookieCheck = types.Int64Value(int64(dns.CookieCheck))
					domain.CookieName = types.StringValue(dns.CookieName)
					domain.CreatedOn = types.StringValue(dns.CreatedOn.String())
					domain.BackPort = types.StringValue(fmt.Sprint(dns.BackPort))
					domain.DomainID = types.StringValue(dns.DomainID)
					domain.CheckInterval = types.Int64Value(int64(dns.CheckInterval))
					domain.FastInterval = types.Int64Value(int64(dns.FastInterval))
					domain.Rise = types.Int64Value(int64(dns.Rise))
					domain.Fall = types.Int64Value(int64(dns.Fall))

					dnsBackends := []Backend{}
					for _, backend := range dns.Backends {
						dnsBackends = append(dnsBackends, Backend{
							IP:           types.StringValue(backend.IP),
							Identifier:   types.StringValue(backend.Identifier),
							VMIdentifier: types.StringValue(backend.VMIdentifier),
							CreatedOn:    types.StringValue(backend.CreatedOn.String()),
						})
					}
					domain.Backends = dnsBackends
					domains = append(domains, domain)
				}

				newRule.Domains = domains

				backends := []Backend{}
				for _, backend := range rule.Backends {
					backends = append(backends, Backend{
						IP:           types.StringValue(backend.IP),
						Identifier:   types.StringValue(backend.Identifier),
						VMIdentifier: types.StringValue(backend.VMIdentifier),
						CreatedOn:    types.StringValue(backend.CreatedOn.String()),
					})
				}
				newRule.Backends = backends

				rules = append(rules, newRule)
			}
			plan.Rules = rules

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
func (l *loadbalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadbalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, err := l.client.LB.GetLB(ctx, state.Identifier.ValueString())
	if err != nil {
		if err.Error() == "loadbalancer not found" {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading vpsie loadbalancer",
			"Could't read vpsie loadbalancer identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Identifier = types.StringValue(lb.Identifier)
	state.UserID = types.Int64Value(int64(lb.UserID))
	state.DcName = types.StringValue(lb.DcName)
	state.CreatedBy = types.StringValue(lb.CreatedBy)
	state.LBName = types.StringValue(lb.LBName)
	state.Traffic = types.Int64Value(int64(lb.Traffic))
	state.BoxsizeID = types.Int64Value(int64(lb.BoxsizeID))
	state.DefaultIP = types.StringValue(lb.DefaultIP)
	state.DcName = types.StringValue(lb.DcName)
	state.DcID = types.StringValue(lb.DcID)
	state.CreatedBy = types.StringValue(lb.CreatedBy)
	state.UserID = types.Int64Value(int64(lb.UserID))

	var rules = []LBRule{}
	for _, rule := range lb.Rules {
		var newRule = LBRule{}
		newRule.BackPort = types.Int64Value(int64(rule.BackPort))
		newRule.FrontPort = types.Int64Value(int64(rule.FrontPort))
		newRule.Scheme = types.StringValue(rule.Scheme)
		newRule.CreatedOn = types.StringValue(rule.CreatedOn.String())
		newRule.RuleID = types.StringValue(rule.RuleID)

		domains := []LBDomain{}
		for _, dns := range rule.Domains {
			domain := LBDomain{}
			domain.DomainName = types.StringValue(dns.DomainName)
			domain.CreatedOn = types.StringValue(dns.CreatedOn.String())

			var subdomain types.String
			if dns.Subdomain != nil && *dns.Subdomain != "" {
				subdomain = types.StringValue(*dns.Subdomain)
			}

			domain.Subdomain = subdomain
			domain.BackendScheme = types.StringValue(dns.BackendScheme)
			domain.Algorithm = types.StringValue(dns.Algorithm)
			domain.RedirectHTTP = types.Int64Value(int64(dns.RedirectHTTP))
			domain.HealthCheckPath = types.StringValue(dns.HealthCheckPath)
			domain.CookieCheck = types.Int64Value(int64(dns.CookieCheck))
			domain.CookieName = types.StringValue(dns.CookieName)
			domain.CreatedOn = types.StringValue(dns.CreatedOn.String())
			domain.BackPort = types.StringValue(fmt.Sprint(dns.BackPort))
			domain.DomainID = types.StringValue(dns.DomainID)
			domain.CheckInterval = types.Int64Value(int64(dns.CheckInterval))
			domain.FastInterval = types.Int64Value(int64(dns.FastInterval))
			domain.Rise = types.Int64Value(int64(dns.Rise))
			domain.Fall = types.Int64Value(int64(dns.Fall))

			dnsBackends := []Backend{}
			for _, backend := range dns.Backends {
				dnsBackends = append(dnsBackends, Backend{
					IP:           types.StringValue(backend.IP),
					Identifier:   types.StringValue(backend.Identifier),
					VMIdentifier: types.StringValue(backend.VMIdentifier),
					CreatedOn:    types.StringValue(backend.CreatedOn.String()),
				})
			}
			domain.Backends = dnsBackends
			domains = append(domains, domain)
		}

		newRule.Domains = domains

		backends := []Backend{}
		for _, backend := range rule.Backends {
			backends = append(backends, Backend{
				IP:           types.StringValue(backend.IP),
				Identifier:   types.StringValue(backend.Identifier),
				VMIdentifier: types.StringValue(backend.VMIdentifier),
				CreatedOn:    types.StringValue(backend.CreatedOn.String()),
			})
		}
		newRule.Backends = backends

		rules = append(rules, newRule)
	}
	state.Rules = rules

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (l *loadbalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state loadbalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan loadbalancerResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rules = make(map[string]LBRule)
	for _, rule := range state.Rules {
		rules[rule.RuleID.ValueString()] = rule
	}

	for _, rule := range plan.Rules {
		if rule.RuleID.ValueString() == "" {

			var newRule = govpsie.AddRuleReq{}

			newRule.BackPort = rule.BackPort.String()
			newRule.FrontPort = rule.FrontPort.String()
			newRule.Scheme = rule.Scheme.ValueString()
			newRule.LbId = state.Identifier.String()

			domains := []govpsie.LBDomain{}
			for _, dns := range rule.Domains {
				domain := govpsie.LBDomain{}
				domain.DomainName = dns.DomainName.ValueString()
				domain.BackPort = dns.BackPort.String()
				domain.BackendScheme = dns.BackendScheme.ValueString()

				dnsBackends := []govpsie.Backend{}
				for _, backend := range dns.Backends {
					dnsBackends = append(dnsBackends, govpsie.Backend{
						Ip:           backend.IP.ValueString(),
						VmIdentifier: backend.VMIdentifier.ValueString(),
					})
				}
				domain.Backends = dnsBackends

				domains = append(domains, domain)
			}
			newRule.Domains = domains
			err := l.client.LB.AddLBRule(ctx, &newRule)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating loadbalancer rule",
					"Couldn't create loadbalancer rule, unexpected error: "+err.Error(),
				)

				return
			}

			return
		}

		var newRule = govpsie.RuleUpdateReq{}

		newRule.BackPort = int(rule.BackPort.ValueInt64())
		newRule.FrontPort = int(rule.FrontPort.ValueInt64())
		newRule.Scheme = rule.Scheme.ValueString()

		backends := []govpsie.Backend{}
		for _, backend := range rule.Backends {
			backends = append(backends, govpsie.Backend{
				Ip:           backend.IP.ValueString(),
				VmIdentifier: backend.VMIdentifier.ValueString(),
			})
		}
		newRule.Backends = backends
		newRule.RuleID = rule.RuleID.ValueString()

		for _, dns := range rule.Domains {
			domainUpdateReq := govpsie.DomainUpdateReq{}

			backPort, err := strconv.Atoi(dns.BackPort.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating loadbalancer domain",
					"Couldn't update loadbalancer domain, unexpected error: "+err.Error(),
				)

				return
			}
			domainUpdateReq.BackPort = backPort
			domainUpdateReq.Algorithm = dns.Algorithm.ValueString()
			domainUpdateReq.DomainID = dns.DomainID.ValueString()
			domainUpdateReq.Subdomain = dns.Subdomain.ValueString()
			domainUpdateReq.CheckInterval = int(dns.CheckInterval.ValueInt64())
			domainUpdateReq.CookieName = dns.CookieName.ValueString()
			domainUpdateReq.Fall = int(dns.Fall.ValueInt64())
			domainUpdateReq.FastInterval = int(dns.FastInterval.ValueInt64())
			domainUpdateReq.RedirectHTTP = int(dns.RedirectHTTP.ValueInt64())
			domainUpdateReq.Rise = int(dns.Rise.ValueInt64())

			err = l.client.LB.UpdateLBDomain(ctx, &domainUpdateReq)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating loadbalancer domain",
					"Couldn't update loadbalancer domain, unexpected error: "+err.Error(),
				)

				return
			}

			dnsBackends := []govpsie.Backend{}
			for _, backend := range dns.Backends {
				dnsBackends = append(dnsBackends, govpsie.Backend{
					Ip:           backend.IP.ValueString(),
					VmIdentifier: backend.VMIdentifier.ValueString(),
				})
			}

			err = l.client.LB.UpdateDomainBackend(ctx, dns.DomainID.String(), dnsBackends)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating loadbalancer domain",
					"Couldn't update loadbalancer domain, unexpected error: "+err.Error(),
				)

				return
			}

		}

		err := l.client.LB.UpdateLBRules(ctx, &newRule)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating loadbalancer rule",
				"Couldn't update loadbalancer rule, unexpected error: "+err.Error(),
			)

			return
		}

		delete(rules, rule.RuleID.ValueString())
	}

	for _, rule := range rules {
		err := l.client.LB.DeleteLBRule(ctx, rule.RuleID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting loadbalancer rule",
				"Couldn't delete loadbalancer rule, unexpected error: "+err.Error(),
			)

			return
		}

	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (l *loadbalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadbalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := l.client.LB.DeleteLB(ctx, state.Identifier.ValueString(), "terraform provider", "terraform provider")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting loadbalancer",
			"Couldn't delete loadbalancer, unexpected error: "+err.Error(),
		)

		return
	}
}

func (l *loadbalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (l *loadbalancerResource) checkResourceStatus(ctx context.Context, lbName string) (*govpsie.LBDetails, bool, error) {
	lbs, err := l.client.LB.ListLBs(ctx, nil)
	if err != nil {
		return nil, false, err
	}

	for _, lb := range lbs {
		if lb.LBName == lbName {

			newLb, err := l.client.LB.GetLB(ctx, lb.Identifier)
			if err != nil {
				return nil, false, err
			}

			return newLb, true, nil
		}
	}

	return nil, false, nil
}
