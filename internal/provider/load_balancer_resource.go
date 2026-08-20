package provider

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/epilayer-public/epilayer-go"
	"github.com/epilayer-public/terraform-provider-epilayer/internal/defaultplanmodifier"
	"github.com/epilayer-public/terraform-provider-epilayer/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                     = &LoadBalancerResource{}
	_ resource.ResourceWithConfigure        = &LoadBalancerResource{}
	_ resource.ResourceWithImportState      = &LoadBalancerResource{}
)

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

type LoadBalancerResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *LoadBalancerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (r *LoadBalancerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Load Balancer resource",

		Attributes: map[string]schema.Attribute{
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the load balancer.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the load balancer.",
				Required:            true,
			}),
			"description": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable description for the load balancer.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					defaultplanmodifier.String(""),
				},
			}),
			"region": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The region identifier. Currently, only `NORD-NO-KRS-1` is supported.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("NORD-NO-KRS-1"),
				},
			}),
			"mode": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The load balancer mode. Supported modes: `tcp`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("tcp"),
				},
			}),
			"network": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The private network to use for the load balancer.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}),
			"floating_ip_id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The ID of an existing floating IP to use as the external IP of the load balancer.",
				Optional:            true,
			}),
			"external_ip": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The external IP of the load balancer.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The load balancer status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when the load balancer was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"updated_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when the load balancer was last updated.",
				Computed:            true,
			}),
			"ports": schema.ListNestedAttribute{
				MarkdownDescription: "List of port mappings.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
							MarkdownDescription: "The port exposed on the load balancer.",
							Required:            true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						}),
						"target_port": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
							MarkdownDescription: "The port to forward traffic to on the targets.",
							Required:            true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						}),
						"targets": schema.ListAttribute{
							MarkdownDescription: "List of target IP addresses.",
							ElementType:         types.StringType,
							Required:            true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"health_check": schema.SingleNestedAttribute{
				MarkdownDescription: "Health check configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"protocol": resourceenhancer.Attribute(ctx, schema.StringAttribute{
						MarkdownDescription: "The protocol to use for the health check. Supported protocols: `http`, `https`, `tcp`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("http", "https", "tcp"),
						},
					}),
					"port": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
						MarkdownDescription: "The port to use for the health check.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.Between(1, 65535),
						},
					}),
					"path": resourceenhancer.Attribute(ctx, schema.StringAttribute{
						MarkdownDescription: "The path to use for http/https health checks.",
						Optional:            true,
					}),
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *LoadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Create)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	bodyMap := map[string]interface{}{
		"name":    data.Name.ValueString(),
		"region":  data.Region.ValueString(),
		"network": data.Network.ValueString(),
	}

	if !data.Description.IsNull() {
		bodyMap["description"] = data.Description.ValueString()
	}
	if !data.FloatingIpId.IsNull() && !data.FloatingIpId.IsUnknown() {
		bodyMap["floating_ip_id"] = data.FloatingIpId.ValueString()
	}

	var ports []map[string]interface{}
	for _, port := range data.Ports {
		targets := make([]string, len(port.Targets))
		for i, t := range port.Targets {
			targets[i] = t.ValueString()
		}
		ports = append(ports, map[string]interface{}{
			"port":        int(port.Port.ValueInt64()),
			"target_port": int(port.TargetPort.ValueInt64()),
			"targets":     targets,
		})
	}
	bodyMap["ports"] = ports

	if data.HealthCheck != nil {
		hc := map[string]interface{}{
			"port":     int(data.HealthCheck.Port.ValueInt64()),
			"protocol": data.HealthCheck.Protocol.ValueString(),
		}
		if !data.HealthCheck.Path.IsNull() {
			hc["path"] = data.HealthCheck.Path.ValueString()
		}
		bodyMap["health_check"] = hc
	}

	b, err := json.Marshal(bodyMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create load_balancer", err))
		return
	}

	response, err := r.client.CreateLoadbalancerWithBodyWithResponse(ctx, "application/json", bytes.NewReader(b))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create load_balancer", err))
		return
	}

	lbResponse := response.JSON201
	if lbResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create load_balancer", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &lbResponse.Loadbalancer, response.Body)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a load balancer resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbId := lbResponse.Loadbalancer.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling load_balancer", err))
			return
		}

		tflog.Trace(ctx, "polling a load balancer resource")

		pollResponse, err := r.client.GetLoadbalancerWithResponse(ctx, lbId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling load_balancer", err))
			return
		}

		pollLbResponse := pollResponse.JSON200
		if pollLbResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling load_balancer", ErrorResponse{
				Body:         pollResponse.Body,
				HTTPResponse: pollResponse.HTTPResponse,
				Error:        pollResponse.JSONDefault,
			}))
			return
		}

		status := pollLbResponse.Loadbalancer.Status
		if status == epilayer.LoadbalancerStatusActive || status == epilayer.LoadbalancerStatusError {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &pollLbResponse.Loadbalancer, pollResponse.Body)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == epilayer.LoadbalancerStatusError {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling load_balancer", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *LoadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Read)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	lbId := data.Id.ValueString()

	response, err := r.client.GetLoadbalancerWithResponse(ctx, lbId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read load_balancer", err))
		return
	}

	lbResponse := response.JSON200
	if lbResponse == nil {
		if response.StatusCode() == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read load_balancer", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &lbResponse.Loadbalancer, response.Body)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a load balancer resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Update)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	body := epilayer.UpdateLoadbalancerJSONRequestBody{}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		body.Name = pointer(data.Name.ValueString())
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		body.Description = pointer(data.Description.ValueString())
	}
	if !data.FloatingIpId.IsNull() && !data.FloatingIpId.IsUnknown() {
		body.FloatingIpId = pointer(data.FloatingIpId.ValueString())
	}

	var ports []epilayer.LoadbalancerPort
	for _, port := range data.Ports {
		targets := make([]string, len(port.Targets))
		for i, t := range port.Targets {
			targets[i] = t.ValueString()
		}
		ports = append(ports, epilayer.LoadbalancerPort{
			Port:       int(port.Port.ValueInt64()),
			TargetPort: int(port.TargetPort.ValueInt64()),
			Targets:    targets,
		})
	}
	body.Ports = &ports



	lbId := data.Id.ValueString()

	response, err := r.client.UpdateLoadbalancerWithResponse(ctx, lbId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update load_balancer", err))
		return
	}

	lbResponse := response.JSON200
	if lbResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update load_balancer", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &lbResponse.Loadbalancer, response.Body)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a load balancer resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling load_balancer", err))
			return
		}

		tflog.Trace(ctx, "polling a load balancer resource")

		pollResponse, err := r.client.GetLoadbalancerWithResponse(ctx, lbId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling load_balancer", err))
			return
		}

		pollLbResponse := pollResponse.JSON200
		if pollLbResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling load_balancer", ErrorResponse{
				Body:         pollResponse.Body,
				HTTPResponse: pollResponse.HTTPResponse,
				Error:        pollResponse.JSONDefault,
			}))
			return
		}

		status := pollLbResponse.Loadbalancer.Status
		if status == epilayer.LoadbalancerStatusActive || status == epilayer.LoadbalancerStatusError {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &pollLbResponse.Loadbalancer, pollResponse.Body)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == epilayer.LoadbalancerStatusError {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling load_balancer", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *LoadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Delete)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	lbId := data.Id.ValueString()

	response, err := r.client.DeleteLoadbalancerWithResponse(ctx, lbId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete load_balancer", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete load_balancer", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling load_balancer", err))
			return
		}

		tflog.Trace(ctx, "polling a load balancer resource")

		pollResponse, err := r.client.GetLoadbalancerWithResponse(ctx, lbId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling load_balancer", err))
			return
		}

		if pollResponse.StatusCode() == 404 {
			return
		}
	}
}

func (r *LoadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
