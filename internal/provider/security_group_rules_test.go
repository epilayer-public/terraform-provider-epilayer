package provider

import (
	"context"
	"testing"

	"github.com/epilayer-public/epilayer-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSecurityGroupRuleProtocol(t *testing.T) {
	valid := []string{"all", "icmp", "tcp", "udp", "0", "1", "6", "9", "17", "47", "50", "51", "58", "99", "112", "199", "255"}
	invalid := []string{"256", "300", "-1", "007", "01", "vrrp", "TCP", "", " tcp"}

	for _, protocol := range valid {
		if !securityGroupRuleProtocol.MatchString(protocol) {
			t.Errorf("expected protocol %q to be valid", protocol)
		}
	}
	for _, protocol := range invalid {
		if securityGroupRuleProtocol.MatchString(protocol) {
			t.Errorf("expected protocol %q to be invalid", protocol)
		}
	}
}

func TestPopulateFromClientResponseKeepsProtocolNumbers(t *testing.T) {
	tests := []struct {
		name          string
		apiProtocol   string
		priorProtocol *string
		want          string
	}{
		{"6 for tcp", "tcp", pointer("6"), "6"},
		{"17 for udp", "udp", pointer("17"), "17"},
		{"1 for icmp", "icmp", pointer("1"), "1"},
		{"58 for icmp", "icmp", pointer("58"), "58"},
		{"name stays a name", "tcp", pointer("tcp"), "tcp"},
		{"other protocol number", "112", pointer("112"), "112"},
		{"changed outside Terraform", "tcp", pointer("17"), "tcp"},
		{"import", "tcp", nil, "tcp"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := SecurityGroupResourceModel{}
			if test.priorProtocol != nil {
				data.Rules = []SecurityGroupRuleModel{{Protocol: types.StringValue(*test.priorProtocol)}}
			}
			securityGroup := epilayer.SecurityGroup{Rules: []epilayer.SecurityGroupRule{{
				Direction:    epilayer.SecurityGroupRuleDirectionIngress,
				Protocol:     epilayer.SecurityGroupRuleProtocol(test.apiProtocol),
				PortRangeMin: pointer(0),
				PortRangeMax: pointer(65535),
			}}}

			diags := data.PopulateFromClientResponse(context.Background(), &securityGroup)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if got := data.Rules[0].Protocol.ValueString(); got != test.want {
				t.Errorf("got protocol %q, want %q", got, test.want)
			}
		})
	}
}

func TestPopulateFromClientResponseKeepsEmptyPortsEmpty(t *testing.T) {
	empty := types.Int64Null()
	port := func(value int64) types.Int64 { return types.Int64Value(value) }

	tests := []struct {
		name             string
		protocol         string
		apiMin, apiMax   int
		prior            *SecurityGroupRuleModel
		wantMin, wantMax types.Int64
	}{
		{"all without ports", "all", 0, 65535, &SecurityGroupRuleModel{PortRangeMin: empty, PortRangeMax: empty}, empty, empty},
		{"protocol number without ports", "112", 0, 65535, &SecurityGroupRuleModel{PortRangeMin: empty, PortRangeMax: empty}, empty, empty},
		{"icmp without types", "icmp", 0, 255, &SecurityGroupRuleModel{PortRangeMin: empty, PortRangeMax: empty}, empty, empty},
		{"icmp type 8", "icmp", 8, 8, &SecurityGroupRuleModel{PortRangeMin: port(8), PortRangeMax: port(8)}, port(8), port(8)},
		{"tcp up to the highest port", "tcp", 1024, 65535, &SecurityGroupRuleModel{PortRangeMin: port(1024), PortRangeMax: port(65535)}, port(1024), port(65535)},
		{"tcp with only the highest port given", "tcp", 0, 1000, &SecurityGroupRuleModel{PortRangeMin: empty, PortRangeMax: port(1000)}, empty, port(1000)},
		{"import of all", "all", 0, 65535, nil, empty, empty},
		{"import of tcp 22", "tcp", 22, 22, nil, port(22), port(22)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := SecurityGroupResourceModel{}
			if test.prior != nil {
				data.Rules = []SecurityGroupRuleModel{*test.prior}
			}
			securityGroup := epilayer.SecurityGroup{Rules: []epilayer.SecurityGroupRule{{
				Direction:    epilayer.SecurityGroupRuleDirectionIngress,
				Protocol:     epilayer.SecurityGroupRuleProtocol(test.protocol),
				PortRangeMin: pointer(test.apiMin),
				PortRangeMax: pointer(test.apiMax),
			}}}

			diags := data.PopulateFromClientResponse(context.Background(), &securityGroup)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			rule := data.Rules[0]
			if !rule.PortRangeMin.Equal(test.wantMin) || !rule.PortRangeMax.Equal(test.wantMax) {
				t.Errorf("got ports %v-%v, want %v-%v", rule.PortRangeMin, rule.PortRangeMax, test.wantMin, test.wantMax)
			}
		})
	}
}
