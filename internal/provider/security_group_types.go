package provider

import (
	"context"
	"time"

	"github.com/epilayer-public/epilayer-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityGroupRuleModel struct {
	// Direction The direction of the rule.
	Direction types.String `tfsdk:"direction"`

	// PortRangeMax The maximum port number of the rule.
	PortRangeMax types.Int64 `tfsdk:"port_range_max"`

	// PortRangeMin The minimum port number of the rule.
	PortRangeMin types.Int64 `tfsdk:"port_range_min"`

	// Protocol The protocol of the rule.
	Protocol types.String `tfsdk:"protocol"`
}

type SecurityGroupResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Description The human-readable description for the security group.
	Description types.String `tfsdk:"description"`

	// Id The unique ID of the security group.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the security group.
	Name types.String `tfsdk:"name"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// Rules The security group rules.
	Rules []SecurityGroupRuleModel `tfsdk:"rules"`

	// Status The security group status.
	Status types.String `tfsdk:"status"`

	// Internal

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *SecurityGroupResourceModel) PopulateFromClientResponse(ctx context.Context, securityGroup *epilayer.SecurityGroup) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(securityGroup.CreatedAt.Format(time.RFC3339))
	data.Description = types.StringValue(securityGroup.Description)
	data.Id = types.StringValue(securityGroup.Id)
	data.Name = types.StringValue(securityGroup.Name)
	data.Region = types.StringValue(string(securityGroup.Region))

	priorRules := data.Rules
	data.Rules = nil
	for i, rule := range securityGroup.Rules {
		// The rules from the plan or the prior state, nil on import
		var prior *SecurityGroupRuleModel
		if i < len(priorRules) {
			prior = &priorRules[i]
		}

		// The API fills an empty port range with the whole range of the protocol: ICMP types 0 to 255
		// for icmp, ports 0 to 65535 for everything else
		highest := 65535
		if rule.Protocol == "icmp" {
			highest = 255
		}
		var priorMin, priorMax *types.Int64
		if prior != nil {
			priorMin, priorMax = &prior.PortRangeMin, &prior.PortRangeMax
		}

		// The API stores the protocol numbers 1, 6, 17 and 58 as names. Keep the number from the
		// configuration when it means the protocol the API returned, so the state matches the
		// configuration.
		protocol := types.StringValue(string(rule.Protocol))
		if prior != nil && securityGroupRuleProtocolName(prior.Protocol.ValueString()) == string(rule.Protocol) {
			protocol = prior.Protocol
		}

		data.Rules = append(data.Rules, SecurityGroupRuleModel{
			Direction:    types.StringValue(string(rule.Direction)),
			PortRangeMax: portFromClientResponse(rule.PortRangeMax, highest, priorMax),
			PortRangeMin: portFromClientResponse(rule.PortRangeMin, 0, priorMin),
			Protocol:     protocol,
		})
	}

	data.Status = types.StringValue(string(securityGroup.Status))

	return
}

// portFromClientResponse keeps a port empty when the API returns the value it fills in for an empty
// port (fill) and the configuration left it empty, so the state matches the configuration. Without a
// prior value (on import), the filled-in value counts as empty.
func portFromClientResponse(port *int, fill int, prior *types.Int64) types.Int64 {
	if port == nil {
		return types.Int64Null()
	}
	if *port == fill && (prior == nil || prior.IsNull()) {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*port))
}
