resource "epilayer_security_group" "allow-https" {
  name = "allow-https"
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 443
      port_range_max = 443
    }
  ]
}

# Any other IP protocol is given by its number. Ports don't apply to it, so leave them empty.
resource "epilayer_security_group" "keepalived" {
  name = "keepalived"
  rules = [
    {
      direction = "ingress"
      protocol  = "112" # VRRP
    },
    {
      direction = "egress"
      protocol  = "all"
    }
  ]
}
