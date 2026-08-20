package provider

import (
	"context"
	"encoding/json"
	"time"

	"github.com/epilayer-public/epilayer-go"
	datasourcetimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LoadBalancerPortModel struct {
	Port       types.Int64    `tfsdk:"port"`
	TargetPort types.Int64    `tfsdk:"target_port"`
	Targets    []types.String `tfsdk:"targets"`
}

type LoadBalancerHealthCheckModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int64  `tfsdk:"port"`
	Path     types.String `tfsdk:"path"`
}

type LoadBalancerResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Region       types.String `tfsdk:"region"`
	Mode         types.String `tfsdk:"mode"`
	Network      types.String `tfsdk:"network"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	ExternalIp   types.String `tfsdk:"external_ip"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`

	Ports       []LoadBalancerPortModel       `tfsdk:"ports"`
	HealthCheck *LoadBalancerHealthCheckModel `tfsdk:"health_check"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *LoadBalancerResourceModel) PopulateFromClientResponse(ctx context.Context, loadbalancer *epilayer.Loadbalancer, rawBody []byte) (diag diag.Diagnostics) {
	data.Id = types.StringValue(loadbalancer.Id)
	data.Name = types.StringValue(loadbalancer.Name)
	data.Description = types.StringValue(loadbalancer.Description)
	data.Region = types.StringValue(string(loadbalancer.Region))
	data.Mode = types.StringValue(string(loadbalancer.Mode))
	data.Network = types.StringValue(loadbalancer.Network)

	if loadbalancer.FloatingIpId != nil {
		data.FloatingIpId = types.StringValue(*loadbalancer.FloatingIpId)
	} else {
		data.FloatingIpId = types.StringNull()
	}

	if loadbalancer.ExternalIp != nil {
		data.ExternalIp = types.StringValue(*loadbalancer.ExternalIp)
	} else {
		data.ExternalIp = types.StringNull()
	}

	data.Status = types.StringValue(string(loadbalancer.Status))
	data.CreatedAt = types.StringValue(loadbalancer.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(loadbalancer.UpdatedAt.Format(time.RFC3339))

	data.Ports = nil
	for _, port := range loadbalancer.Ports {
		targets := make([]types.String, len(port.Targets))
		for i, t := range port.Targets {
			targets[i] = types.StringValue(t)
		}

		data.Ports = append(data.Ports, LoadBalancerPortModel{
			Port:       types.Int64Value(int64(port.Port)),
			TargetPort: types.Int64Value(int64(port.TargetPort)),
			Targets:    targets,
		})
	}

	if len(rawBody) > 0 {
		var rawResp struct {
			Loadbalancer struct {
				HealthCheck *struct {
					Protocol string  `json:"protocol"`
					Port     int     `json:"port"`
					Path     *string `json:"path"`
				} `json:"health_check"`
			} `json:"loadbalancer"`
		}
		if err := json.Unmarshal(rawBody, &rawResp); err == nil && rawResp.Loadbalancer.HealthCheck != nil {
			hc := LoadBalancerHealthCheckModel{
				Protocol: types.StringValue(rawResp.Loadbalancer.HealthCheck.Protocol),
				Port:     types.Int64Value(int64(rawResp.Loadbalancer.HealthCheck.Port)),
			}
			if rawResp.Loadbalancer.HealthCheck.Path != nil {
				hc.Path = types.StringValue(*rawResp.Loadbalancer.HealthCheck.Path)
			} else {
				hc.Path = types.StringNull()
			}
			data.HealthCheck = &hc
		} else {
			data.HealthCheck = nil
		}
	} else if loadbalancer.HealthCheck != nil {
		hc := LoadBalancerHealthCheckModel{
			Protocol: types.StringValue(string(loadbalancer.HealthCheck.Protocol)),
			Port:     types.Int64Value(int64(loadbalancer.HealthCheck.Port)),
		}
		if loadbalancer.HealthCheck.Path != nil {
			hc.Path = types.StringValue(*loadbalancer.HealthCheck.Path)
		} else {
			hc.Path = types.StringNull()
		}
		data.HealthCheck = &hc
	} else {
		data.HealthCheck = nil
	}

	return
}

type LoadBalancerDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Region       types.String `tfsdk:"region"`
	Mode         types.String `tfsdk:"mode"`
	Network      types.String `tfsdk:"network"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	ExternalIp   types.String `tfsdk:"external_ip"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`

	Ports       []LoadBalancerPortModel       `tfsdk:"ports"`
	HealthCheck *LoadBalancerHealthCheckModel `tfsdk:"health_check"`

	Timeouts datasourcetimeouts.Value `tfsdk:"timeouts"`
}

func (data *LoadBalancerDataSourceModel) PopulateFromClientResponse(ctx context.Context, loadbalancer *epilayer.Loadbalancer) (diag diag.Diagnostics) {
	data.Id = types.StringValue(loadbalancer.Id)
	data.Name = types.StringValue(loadbalancer.Name)
	data.Description = types.StringValue(loadbalancer.Description)
	data.Region = types.StringValue(string(loadbalancer.Region))
	data.Mode = types.StringValue(string(loadbalancer.Mode))
	data.Network = types.StringValue(loadbalancer.Network)

	if loadbalancer.FloatingIpId != nil {
		data.FloatingIpId = types.StringValue(*loadbalancer.FloatingIpId)
	} else {
		data.FloatingIpId = types.StringNull()
	}

	if loadbalancer.ExternalIp != nil {
		data.ExternalIp = types.StringValue(*loadbalancer.ExternalIp)
	} else {
		data.ExternalIp = types.StringNull()
	}

	data.Status = types.StringValue(string(loadbalancer.Status))
	data.CreatedAt = types.StringValue(loadbalancer.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(loadbalancer.UpdatedAt.Format(time.RFC3339))

	data.Ports = nil
	for _, port := range loadbalancer.Ports {
		targets := make([]types.String, len(port.Targets))
		for i, t := range port.Targets {
			targets[i] = types.StringValue(t)
		}

		data.Ports = append(data.Ports, LoadBalancerPortModel{
			Port:       types.Int64Value(int64(port.Port)),
			TargetPort: types.Int64Value(int64(port.TargetPort)),
			Targets:    targets,
		})
	}

	if loadbalancer.HealthCheck != nil {
		hc := LoadBalancerHealthCheckModel{
			Protocol: types.StringValue(string(loadbalancer.HealthCheck.Protocol)),
			Port:     types.Int64Value(int64(loadbalancer.HealthCheck.Port)),
		}
		if loadbalancer.HealthCheck.Path != nil {
			hc.Path = types.StringValue(*loadbalancer.HealthCheck.Path)
		} else {
			hc.Path = types.StringNull()
		}
		data.HealthCheck = &hc
	}

	return
}
