resource "epilayer_instance" "example" {
  name = "example"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}


resource "epilayer_instance_status" "example" {
  instance_id = epilayer_instance.example.id
  status      = "active"
}
