package loadbalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// loadbalancerResourceModel mirrors the resource schema one-for-one. Every
// tfsdk tag here must name an attribute or block that Schema() declares,
// otherwise the framework fails the plan with a conversion error.
type loadbalancerResourceModel struct {
	LBName             types.String `tfsdk:"lb_name"`
	DcIdentifier       types.String `tfsdk:"dc_identifier"`
	ResourceIdentifier types.String `tfsdk:"resource_identifier"`
	PrivateLB          types.Bool   `tfsdk:"private_lb"`
	VpcID              types.Int64  `tfsdk:"vpc_id"`
	ProjectID          types.String `tfsdk:"project_id"`
	Tags               types.List   `tfsdk:"tags"`

	Identifier types.String `tfsdk:"identifier"`
	DefaultIP  types.String `tfsdk:"default_ip"`
	DcName     types.String `tfsdk:"dc_name"`
	Traffic    types.Int64  `tfsdk:"traffic"`
	BoxsizeID  types.Int64  `tfsdk:"boxsize_id"`
	CreatedBy  types.String `tfsdk:"created_by"`
	UserID     types.Int64  `tfsdk:"user_id"`

	Rules    []lbRuleModel  `tfsdk:"rule"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type lbRuleModel struct {
	Scheme    types.String     `tfsdk:"scheme"`
	FrontPort types.Int64      `tfsdk:"front_port"`
	BackPort  types.Int64      `tfsdk:"back_port"`
	ProxyMode types.Bool       `tfsdk:"proxy_mode"`
	RuleID    types.String     `tfsdk:"rule_id"`
	Backends  []lbBackendModel `tfsdk:"backend"`
	Domains   []lbDomainModel  `tfsdk:"domain"`
}

type lbDomainModel struct {
	DomainName      types.String     `tfsdk:"domain_name"`
	Subdomain       types.String     `tfsdk:"subdomain"`
	BackPort        types.Int64      `tfsdk:"back_port"`
	Algorithm       types.String     `tfsdk:"algorithm"`
	RedirectHTTP    types.Int64      `tfsdk:"redirect_http"`
	CookieCheck     types.Bool       `tfsdk:"cookie_check"`
	CookieName      types.String     `tfsdk:"cookie_name"`
	CheckInterval   types.Int64      `tfsdk:"check_interval"`
	FastInterval    types.Int64      `tfsdk:"fast_interval"`
	Rise            types.Int64      `tfsdk:"rise"`
	Fall            types.Int64      `tfsdk:"fall"`
	HealthCheckPath types.String     `tfsdk:"health_check_path"`
	BackendScheme   types.String     `tfsdk:"backend_scheme"`
	PassThrough     types.Bool       `tfsdk:"pass_through"`
	DomainID        types.String     `tfsdk:"domain_id"`
	Backends        []lbBackendModel `tfsdk:"backend"`
}

type lbBackendModel struct {
	IP           types.String `tfsdk:"ip"`
	VMIdentifier types.String `tfsdk:"vm_identifier"`
	Type         types.String `tfsdk:"type"`
	Identifier   types.String `tfsdk:"identifier"`
}

func NewLoadbalancerResource() resource.Resource {
	return &loadbalancerResource{}
}

func (l *loadbalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

// backendBlock is the shared `backend` block used under both `rule` and
// `domain`. A backend is a real target: an IP, optionally tied to a VM.
func backendBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "A backend target receiving traffic for this listener.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"ip": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "IPv4 address of the backend target.",
				},
				"vm_identifier": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
					MarkdownDescription: "Identifier of the VPSie VM serving this backend, when the target is a VM.",
				},
				"type": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Default:  stringdefault.StaticString("vm"),
					Validators: []validator.String{
						stringvalidator.OneOf("vm", "k8s"),
					},
					MarkdownDescription: "Backend type. One of `vm` or `k8s`. Defaults to `vm`.",
				},
				"identifier": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Backend identifier assigned by the API.",
				},
			},
		},
	}
}

func (l *loadbalancerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VPSie load balancer, including its listener rules, virtual-host domains and backends.\n\n" +
			"Changing `dc_identifier`, `resource_identifier`, `private_lb`, `vpc_id` or `project_id` replaces the load balancer, " +
			"which releases its IP address. Renaming and rule changes are applied in place.",
		Attributes: map[string]schema.Attribute{
			"lb_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the load balancer (2-64 characters). Updating it renames the load balancer in place.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 64),
				},
			},
			"dc_identifier": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identifier of the datacenter to deploy into. Changing this forces a new load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_identifier": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identifier of the load balancer offer (size). Changing this forces a new load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_lb": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the load balancer is private (VPC-only). Requires `vpc_id`. Changing this forces a new load balancer.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Numeric VPC id to attach the load balancer to. Required when `private_lb` is true. Changing this forces a new load balancer.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Identifier of the project owning the load balancer. Changing this forces a new load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Tag names to apply at creation time. Changing this forces a new load balancer; use `vpsie_tag` to manage tags over time.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"identifier": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the load balancer assigned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IP address the load balancer listens on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dc_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Human-readable datacenter name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"traffic": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Traffic allowance of the selected offer, in TB.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"boxsize_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Numeric id of the offer backing the load balancer.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Account that created the load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Numeric id of the owning user.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				MarkdownDescription: "A listener on the load balancer. At least one rule is required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"scheme": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("http", "https", "http2", "tcp"),
							},
							MarkdownDescription: "Listener protocol. One of `http`, `https`, `http2` or `tcp`.",
						},
						"front_port": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
							MarkdownDescription: "Port the load balancer listens on.",
						},
						"back_port": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
							MarkdownDescription: "Port on the backend targets. Required when `scheme` is `tcp`.",
						},
						"proxy_mode": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Enable PROXY protocol towards the backends.",
						},
						"rule_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Rule identifier assigned by the API.",
						},
					},
					Blocks: map[string]schema.Block{
						"backend": backendBlock(),
						"domain": schema.ListNestedBlock{
							MarkdownDescription: "A virtual host served by this listener, with its own health checks and backends.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"domain_name": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString(""),
										MarkdownDescription: "Fully-qualified domain name served by this virtual host.",
									},
									"subdomain": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString(""),
										MarkdownDescription: "Subdomain label served by this virtual host.",
									},
									"back_port": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(1, 65535),
										},
										MarkdownDescription: "Port on the backend targets for this virtual host.",
									},
									"algorithm": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Default:  stringdefault.StaticString("roundrobin"),
										Validators: []validator.String{
											stringvalidator.OneOf("roundrobin", "leastconn"),
										},
										MarkdownDescription: "Balancing algorithm. One of `roundrobin` or `leastconn`.",
									},
									"redirect_http": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             int64default.StaticInt64(0),
										Validators:          []validator.Int64{int64validator.OneOf(0, 1)},
										MarkdownDescription: "Set to `1` to redirect plain HTTP to HTTPS.",
									},
									"cookie_check": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             booldefault.StaticBool(false),
										MarkdownDescription: "Enable cookie-based session persistence.",
									},
									"cookie_name": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString(""),
										MarkdownDescription: "Name of the persistence cookie when `cookie_check` is enabled.",
									},
									"check_interval": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             int64default.StaticInt64(1000),
										Validators:          []validator.Int64{int64validator.AtLeast(100)},
										MarkdownDescription: "Health check interval in milliseconds (minimum 100).",
									},
									"fast_interval": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             int64default.StaticInt64(500),
										Validators:          []validator.Int64{int64validator.AtLeast(100)},
										MarkdownDescription: "Health check interval in milliseconds while a target is in transition (minimum 100).",
									},
									"rise": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             int64default.StaticInt64(5),
										Validators:          []validator.Int64{int64validator.AtLeast(1)},
										MarkdownDescription: "Consecutive successful checks before a target is considered healthy.",
									},
									"fall": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             int64default.StaticInt64(2),
										Validators:          []validator.Int64{int64validator.AtLeast(1)},
										MarkdownDescription: "Consecutive failed checks before a target is considered unhealthy.",
									},
									"health_check_path": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString("/"),
										MarkdownDescription: "HTTP path polled for health checks.",
									},
									"backend_scheme": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Default:  stringdefault.StaticString("http"),
										Validators: []validator.String{
											stringvalidator.OneOf("http", "https"),
										},
										MarkdownDescription: "Protocol used towards the backends. One of `http` or `https`.",
									},
									"pass_through": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             booldefault.StaticBool(false),
										MarkdownDescription: "Pass TLS through to the backends without terminating it.",
									},
									"domain_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Domain identifier assigned by the API.",
									},
								},
								Blocks: map[string]schema.Block{
									"backend": backendBlock(),
								},
							},
						},
					},
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
			"Unexpected Resource Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	l.client = client
}

// ruleKey is the natural key used to pair a configured rule with the rule the
// API reports back. The API assigns rule identifiers itself, so listener
// protocol plus front port is the only stable pairing available at create time.
func ruleKey(scheme string, frontPort int64) string {
	return fmt.Sprintf("%s:%d", scheme, frontPort)
}

// domainKey pairs a configured domain with the one the API reports back.
func domainKey(domainName, subdomain string) string {
	return domainName + "|" + subdomain
}

func backendsToAPI(backends []lbBackendModel) []govpsie.Backend {
	out := make([]govpsie.Backend, 0, len(backends))
	for _, backend := range backends {
		out = append(out, govpsie.Backend{
			Ip:           backend.IP.ValueString(),
			VmIdentifier: backend.VMIdentifier.ValueString(),
			Type:         backend.Type.ValueString(),
		})
	}
	return out
}

func domainsToAPI(domains []lbDomainModel) []govpsie.LBDomain {
	out := make([]govpsie.LBDomain, 0, len(domains))
	for _, domain := range domains {
		out = append(out, govpsie.LBDomain{
			DomainID:        domain.DomainID.ValueString(),
			DomainName:      domain.DomainName.ValueString(),
			Subdomain:       domain.Subdomain.ValueString(),
			BackPort:        int(domain.BackPort.ValueInt64()),
			Algorithm:       domain.Algorithm.ValueString(),
			RedirectHTTP:    int(domain.RedirectHTTP.ValueInt64()),
			CookieCheck:     domain.CookieCheck.ValueBool(),
			CookieName:      domain.CookieName.ValueString(),
			CheckInterval:   int(domain.CheckInterval.ValueInt64()),
			FastInterval:    int(domain.FastInterval.ValueInt64()),
			Rise:            int(domain.Rise.ValueInt64()),
			Fall:            int(domain.Fall.ValueInt64()),
			HealthCheckPath: domain.HealthCheckPath.ValueString(),
			BackendScheme:   domain.BackendScheme.ValueString(),
			PassThrough:     domain.PassThrough.ValueBool(),
			Backends:        backendsToAPI(domain.Backends),
		})
	}
	return out
}

func rulesToAPI(rules []lbRuleModel) []govpsie.Rule {
	out := make([]govpsie.Rule, 0, len(rules))
	for _, rule := range rules {
		out = append(out, govpsie.Rule{
			Scheme:    rule.Scheme.ValueString(),
			FrontPort: int(rule.FrontPort.ValueInt64()),
			BackPort:  int(rule.BackPort.ValueInt64()),
			ProxyMode: rule.ProxyMode.ValueBool(),
			Domains:   domainsToAPI(rule.Domains),
			Backends:  backendsToAPI(rule.Backends),
		})
	}
	return out
}

// applyComputed copies the API-owned fields of the load balancer onto the
// model. Config-owned attributes are deliberately left untouched so that a
// value the API normalizes (or omits) cannot produce a perpetual diff.
func applyComputed(model *loadbalancerResourceModel, lb *govpsie.LBDetails) {
	model.Identifier = types.StringValue(lb.Identifier)
	model.DefaultIP = types.StringValue(lb.DefaultIP)
	model.DcName = types.StringValue(lb.DcName)
	model.Traffic = types.Int64Value(int64(lb.Traffic))
	model.BoxsizeID = types.Int64Value(int64(lb.BoxsizeID))
	model.CreatedBy = types.StringValue(lb.CreatedBy)
	model.UserID = types.Int64Value(int64(lb.UserID))
}

// applyIdentifiers fills in the API-assigned identifiers (rule, domain and
// backend) on the configured rules, pairing by natural key. Anything the API
// did not report keeps a known empty value so the state stays fully known.
func applyIdentifiers(model *loadbalancerResourceModel, lb *govpsie.LBDetails) {
	apiRules := map[string]govpsie.LBRuleDetail{}
	for _, apiRule := range lb.Rules {
		apiRules[ruleKey(apiRule.Scheme, int64(apiRule.FrontPort))] = apiRule
	}

	for i := range model.Rules {
		rule := &model.Rules[i]
		apiRule, found := apiRules[ruleKey(rule.Scheme.ValueString(), rule.FrontPort.ValueInt64())]
		if !found {
			rule.RuleID = types.StringValue("")
			clearRuleIdentifiers(rule)
			continue
		}

		rule.RuleID = types.StringValue(apiRule.RuleID)
		applyBackendIdentifiers(rule.Backends, apiRule.Backends)

		apiDomains := map[string]govpsie.LBDomainsDetail{}
		for _, apiDomain := range apiRule.Domains {
			subdomain := ""
			if apiDomain.Subdomain != nil {
				subdomain = *apiDomain.Subdomain
			}
			apiDomains[domainKey(apiDomain.DomainName, subdomain)] = apiDomain
		}

		for j := range rule.Domains {
			domain := &rule.Domains[j]
			apiDomain, ok := apiDomains[domainKey(domain.DomainName.ValueString(), domain.Subdomain.ValueString())]
			if !ok {
				domain.DomainID = types.StringValue("")
				applyBackendIdentifiers(domain.Backends, nil)
				continue
			}
			domain.DomainID = types.StringValue(apiDomain.DomainID)
			applyBackendIdentifiers(domain.Backends, apiDomain.Backends)
		}
	}
}

func clearRuleIdentifiers(rule *lbRuleModel) {
	applyBackendIdentifiers(rule.Backends, nil)
	for j := range rule.Domains {
		rule.Domains[j].DomainID = types.StringValue("")
		applyBackendIdentifiers(rule.Domains[j].Backends, nil)
	}
}

// applyBackendIdentifiers pairs configured backends with API backends by IP,
// which is the only field the caller controls and the API echoes verbatim.
func applyBackendIdentifiers(backends []lbBackendModel, apiBackends []govpsie.LBBackendsDetail) {
	byIP := map[string]govpsie.LBBackendsDetail{}
	for _, apiBackend := range apiBackends {
		byIP[apiBackend.IP] = apiBackend
	}

	for i := range backends {
		apiBackend, ok := byIP[backends[i].IP.ValueString()]
		if !ok {
			backends[i].Identifier = types.StringValue("")
			continue
		}
		backends[i].Identifier = types.StringValue(apiBackend.Identifier)
	}
}

func (l *loadbalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadbalancerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(plan.Rules) == 0 {
		resp.Diagnostics.AddError(
			"Missing load balancer rule",
			"At least one `rule` block is required to create a load balancer.",
		)
		return
	}

	for _, rule := range plan.Rules {
		if rule.Scheme.ValueString() == "tcp" && rule.BackPort.IsNull() {
			resp.Diagnostics.AddError(
				"Missing back_port",
				fmt.Sprintf("Rule %s requires `back_port` because its scheme is `tcp`.", ruleKey(rule.Scheme.ValueString(), rule.FrontPort.ValueInt64())),
			)
			return
		}
	}

	privateLB := 0
	if plan.PrivateLB.ValueBool() {
		privateLB = 1
		if plan.VpcID.IsNull() {
			resp.Diagnostics.AddError(
				"Missing vpc_id",
				"`vpc_id` is required when `private_lb` is true.",
			)
			return
		}
	}

	tags := []string{}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := &govpsie.CreateLBReq{
		LBName:             plan.LBName.ValueString(),
		DcIdentifier:       plan.DcIdentifier.ValueString(),
		ResourceIdentifier: plan.ResourceIdentifier.ValueString(),
		PrivateLB:          privateLB,
		VpcID:              int(plan.VpcID.ValueInt64()),
		ProjectID:          plan.ProjectID.ValueString(),
		Rules:              rulesToAPI(plan.Rules),
		InputTags:          tags,
	}

	if err := l.client.LB.CreateLB(ctx, createReq); err != nil {
		resp.Diagnostics.AddError("Error creating load balancer", err.Error())
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// The create endpoint enqueues the build and returns no identifier, so the
	// load balancer has to be recovered by name. Back off gradually to avoid
	// hammering the API across what can be a multi-minute provision.
	delay := 5 * time.Second
	const maxDelay = 30 * time.Second
	for {
		lb, ready, err := l.checkResourceStatus(ctx, plan.LBName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error checking load balancer status", err.Error())
			return
		}

		if ready {
			// The create endpoint accepts `rules` but does not reliably persist
			// them -- a rule whose backends name a VM is silently dropped, so a
			// load balancer can come up with no listeners at all. Reconcile by
			// adding whatever the API is missing.
			if err := l.reconcileRules(ctx, lb, plan.Rules); err != nil {
				resp.Diagnostics.AddError("Error creating load balancer rules", err.Error())
				return
			}

			lb, err = l.client.LB.GetLB(ctx, lb.Identifier)
			if err != nil {
				resp.Diagnostics.AddError("Error reading load balancer after create", err.Error())
				return
			}
			if lb == nil || lb.Identifier == "" {
				resp.Diagnostics.AddError(
					"Load balancer disappeared during create",
					"The load balancer could not be read back after its rules were applied.",
				)
				return
			}

			applyComputed(&plan, lb)
			applyIdentifiers(&plan, lb)
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			return
		}

		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError(
				"Timed out waiting for load balancer",
				fmt.Sprintf("Load balancer %q did not become ready before the create timeout elapsed. "+
					"It may still be provisioning; import it with `terraform import` once it is ready to avoid an orphaned billable resource.", plan.LBName.ValueString()),
			)
			return
		case <-time.After(delay):
		}

		if delay < maxDelay {
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	}
}

func (l *loadbalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadbalancerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, err := l.client.LB.GetLB(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading load balancer", "couldn't read load balancer, unexpected error: "+err.Error())
		return
	}

	// The API answers a missing load balancer with an empty payload rather than
	// a 404, so treat an absent identifier as a deleted resource.
	if lb == nil || lb.Identifier == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	applyComputed(&state, lb)

	// Pair the configured listeners with what the API reports. When the sets
	// match, keep the config-owned values so API normalisation cannot cause a
	// perpetual diff. When they differ, a listener was added or removed out of
	// band, so rebuild from the API to surface the drift in the next plan.
	if sameRuleKeys(state.Rules, lb.Rules) {
		applyIdentifiers(&state, lb)
	} else {
		state.Rules = rulesFromAPI(lb.Rules)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (l *loadbalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state loadbalancerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := state.Identifier.ValueString()

	if plan.LBName.ValueString() != state.LBName.ValueString() {
		if err := l.client.LB.UpdateLBName(ctx, identifier, plan.LBName.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error renaming load balancer", err.Error())
			return
		}
	}

	// Reconcile listeners by natural key so that unrelated listeners keep
	// serving traffic. A changed listener is replaced rather than patched:
	// rules own their domains and backends, and the API has no single endpoint
	// that rewrites a whole rule tree atomically.
	stateRules := map[string]lbRuleModel{}
	for _, rule := range state.Rules {
		stateRules[ruleKey(rule.Scheme.ValueString(), rule.FrontPort.ValueInt64())] = rule
	}

	planRules := map[string]lbRuleModel{}
	for _, rule := range plan.Rules {
		planRules[ruleKey(rule.Scheme.ValueString(), rule.FrontPort.ValueInt64())] = rule
	}

	for key, stateRule := range stateRules {
		planRule, kept := planRules[key]
		if kept && ruleEqual(stateRule, planRule) {
			continue
		}

		if ruleID := stateRule.RuleID.ValueString(); ruleID != "" {
			if err := l.client.LB.DeleteLBRule(ctx, ruleID); err != nil {
				resp.Diagnostics.AddError("Error deleting load balancer rule", err.Error())
				return
			}
		}
	}

	for key, planRule := range planRules {
		stateRule, existed := stateRules[key]
		if existed && ruleEqual(stateRule, planRule) {
			continue
		}

		apiRule := rulesToAPI([]lbRuleModel{planRule})[0]
		addReq := &govpsie.AddRuleReq{
			LbId:      identifier,
			Scheme:    apiRule.Scheme,
			FrontPort: apiRule.FrontPort,
			BackPort:  apiRule.BackPort,
			ProxyMode: apiRule.ProxyMode,
			Domains:   apiRule.Domains,
			Backends:  apiRule.Backends,
		}

		if err := l.client.LB.AddLBRule(ctx, addReq); err != nil {
			resp.Diagnostics.AddError("Error adding load balancer rule", err.Error())
			return
		}
	}

	lb, err := l.client.LB.GetLB(ctx, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error reading load balancer after update", err.Error())
		return
	}
	if lb == nil || lb.Identifier == "" {
		resp.Diagnostics.AddError(
			"Load balancer disappeared during update",
			fmt.Sprintf("Load balancer %q could not be read back after the update was applied.", identifier),
		)
		return
	}

	applyComputed(&plan, lb)
	applyIdentifiers(&plan, lb)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (l *loadbalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadbalancerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := l.client.LB.DeleteLB(ctx, state.Identifier.ValueString(), "Deleted by Terraform", "")
	if err != nil {
		resp.Diagnostics.AddError("Error deleting load balancer", err.Error())
		return
	}
}

func (l *loadbalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// checkResourceStatus looks the load balancer up by name and reports whether it
// has finished provisioning. The create endpoint returns no identifier, so this
// is the only way to recover the newly created load balancer.
func (l *loadbalancerResource) checkResourceStatus(ctx context.Context, lbName string) (*govpsie.LBDetails, bool, error) {
	lbs, err := l.client.LB.ListLBs(ctx, nil)
	if err != nil {
		return nil, false, err
	}

	for _, lb := range lbs {
		if lb.LBName != lbName {
			continue
		}

		detail, err := l.client.LB.GetLB(ctx, lb.Identifier)
		if err != nil {
			return nil, false, err
		}
		if detail == nil || detail.Identifier == "" {
			return nil, false, nil
		}

		return detail, true, nil
	}

	return nil, false, nil
}

// ruleEqual reports whether two configured rules describe the same listener,
// including their domains and backends. API-assigned identifiers are ignored so
// that a rule read back from the API still compares equal to its configuration.
func ruleEqual(a, b lbRuleModel) bool {
	if !a.Scheme.Equal(b.Scheme) ||
		!a.FrontPort.Equal(b.FrontPort) ||
		!a.BackPort.Equal(b.BackPort) ||
		!a.ProxyMode.Equal(b.ProxyMode) {
		return false
	}

	if !backendsEqual(a.Backends, b.Backends) {
		return false
	}

	if len(a.Domains) != len(b.Domains) {
		return false
	}

	for i := range a.Domains {
		if !domainEqual(a.Domains[i], b.Domains[i]) {
			return false
		}
	}

	return true
}

func domainEqual(a, b lbDomainModel) bool {
	return a.DomainName.Equal(b.DomainName) &&
		a.Subdomain.Equal(b.Subdomain) &&
		a.BackPort.Equal(b.BackPort) &&
		a.Algorithm.Equal(b.Algorithm) &&
		a.RedirectHTTP.Equal(b.RedirectHTTP) &&
		a.CookieCheck.Equal(b.CookieCheck) &&
		a.CookieName.Equal(b.CookieName) &&
		a.CheckInterval.Equal(b.CheckInterval) &&
		a.FastInterval.Equal(b.FastInterval) &&
		a.Rise.Equal(b.Rise) &&
		a.Fall.Equal(b.Fall) &&
		a.HealthCheckPath.Equal(b.HealthCheckPath) &&
		a.BackendScheme.Equal(b.BackendScheme) &&
		a.PassThrough.Equal(b.PassThrough) &&
		backendsEqual(a.Backends, b.Backends)
}

func backendsEqual(a, b []lbBackendModel) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].IP.Equal(b[i].IP) ||
			!a[i].VMIdentifier.Equal(b[i].VMIdentifier) ||
			!a[i].Type.Equal(b[i].Type) {
			return false
		}
	}

	return true
}

// reconcileRules adds any configured listener the API is not already serving.
func (l *loadbalancerResource) reconcileRules(ctx context.Context, lb *govpsie.LBDetails, rules []lbRuleModel) error {
	existing := map[string]struct{}{}
	for _, apiRule := range lb.Rules {
		existing[ruleKey(apiRule.Scheme, int64(apiRule.FrontPort))] = struct{}{}
	}

	for _, rule := range rules {
		if _, ok := existing[ruleKey(rule.Scheme.ValueString(), rule.FrontPort.ValueInt64())]; ok {
			continue
		}

		apiRule := rulesToAPI([]lbRuleModel{rule})[0]
		addReq := &govpsie.AddRuleReq{
			LbId:      lb.Identifier,
			Scheme:    apiRule.Scheme,
			FrontPort: apiRule.FrontPort,
			BackPort:  apiRule.BackPort,
			ProxyMode: apiRule.ProxyMode,
			Domains:   apiRule.Domains,
			Backends:  apiRule.Backends,
		}

		if err := l.client.LB.AddLBRule(ctx, addReq); err != nil {
			return fmt.Errorf("adding rule %s: %w", ruleKey(apiRule.Scheme, int64(apiRule.FrontPort)), err)
		}
	}

	return nil
}

// sameRuleKeys reports whether the configured listeners and the API's listeners
// describe the same set of scheme/front-port pairs.
func sameRuleKeys(rules []lbRuleModel, apiRules []govpsie.LBRuleDetail) bool {
	if len(rules) != len(apiRules) {
		return false
	}

	keys := map[string]struct{}{}
	for _, rule := range rules {
		keys[ruleKey(rule.Scheme.ValueString(), rule.FrontPort.ValueInt64())] = struct{}{}
	}

	for _, apiRule := range apiRules {
		if _, ok := keys[ruleKey(apiRule.Scheme, int64(apiRule.FrontPort))]; !ok {
			return false
		}
	}

	return true
}

// rulesFromAPI rebuilds the listener blocks from what the API reports, used
// when the managed listener set has drifted.
func rulesFromAPI(apiRules []govpsie.LBRuleDetail) []lbRuleModel {
	rules := make([]lbRuleModel, 0, len(apiRules))

	for _, apiRule := range apiRules {
		rule := lbRuleModel{
			Scheme:    types.StringValue(apiRule.Scheme),
			FrontPort: types.Int64Value(int64(apiRule.FrontPort)),
			BackPort:  types.Int64Value(int64(apiRule.BackPort)),
			ProxyMode: types.BoolValue(false),
			RuleID:    types.StringValue(apiRule.RuleID),
			Backends:  backendsFromAPI(apiRule.Backends),
		}

		for _, apiDomain := range apiRule.Domains {
			subdomain := ""
			if apiDomain.Subdomain != nil {
				subdomain = *apiDomain.Subdomain
			}

			rule.Domains = append(rule.Domains, lbDomainModel{
				DomainName:      types.StringValue(apiDomain.DomainName),
				Subdomain:       types.StringValue(subdomain),
				BackPort:        types.Int64Value(int64(apiDomain.BackPort)),
				Algorithm:       types.StringValue(apiDomain.Algorithm),
				RedirectHTTP:    types.Int64Value(int64(apiDomain.RedirectHTTP)),
				CookieCheck:     types.BoolValue(apiDomain.CookieCheck != 0),
				CookieName:      types.StringValue(apiDomain.CookieName),
				CheckInterval:   types.Int64Value(int64(apiDomain.CheckInterval)),
				FastInterval:    types.Int64Value(int64(apiDomain.FastInterval)),
				Rise:            types.Int64Value(int64(apiDomain.Rise)),
				Fall:            types.Int64Value(int64(apiDomain.Fall)),
				HealthCheckPath: types.StringValue(apiDomain.HealthCheckPath),
				BackendScheme:   types.StringValue(apiDomain.BackendScheme),
				PassThrough:     types.BoolValue(false),
				DomainID:        types.StringValue(apiDomain.DomainID),
				Backends:        backendsFromAPI(apiDomain.Backends),
			})
		}

		rules = append(rules, rule)
	}

	return rules
}

func backendsFromAPI(apiBackends []govpsie.LBBackendsDetail) []lbBackendModel {
	backends := make([]lbBackendModel, 0, len(apiBackends))
	for _, apiBackend := range apiBackends {
		backends = append(backends, lbBackendModel{
			IP:           types.StringValue(apiBackend.IP),
			VMIdentifier: types.StringValue(apiBackend.VMIdentifier),
			Type:         types.StringValue("vm"),
			Identifier:   types.StringValue(apiBackend.Identifier),
		})
	}
	return backends
}
