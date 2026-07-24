# Example: control public IP assignment

# Disable public IP on create
resource "epilayer_instance" "no_public_ip" {
  name   = "no-public-ip"
  region = "NORD-NO-KRS-1"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  # ensure no public IPv4 is assigned
  public_ip = "none"

  ssh_key_ids = ["my-ssh-key-id"]
}

# Attach an existing floating IP created in this configuration
resource "epilayer_floating_ip" "floating_ip" {
  name    = "terraform-floating-ip"
  region  = "NORD-NO-KRS-1"
  version = "ipv4"
}

resource "epilayer_instance" "with_floating_ip" {
  name   = "with-floating-ip"
  region = "NORD-NO-KRS-1"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  # attach an existing floating IP by id
  floating_ip_id = epilayer_floating_ip.floating_ip.id

  ssh_key_ids = ["my-ssh-key-id"]
}
