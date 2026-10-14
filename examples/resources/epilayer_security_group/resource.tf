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
