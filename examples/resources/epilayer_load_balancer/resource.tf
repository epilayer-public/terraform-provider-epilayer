data "epilayer_images" "cloud_images" {
  filter = {
    type = "cloud-image"
  }
}

locals {
  # Find the image ID for ubuntu-24.04, or fallback to the first available image if not found
  image_id = try(
    [for img in data.epilayer_images.cloud_images.images : img.id if img.slug == "ubuntu-24.04"][0],
    data.epilayer_images.cloud_images.images[0].id
  )
}

resource "epilayer_instance" "example" {
  name   = "example-instance"
  region = "NORD-NO-KRS-1"

  image = local.image_id
  type  = "vcpu-2_memory-4g"

  private_network_ids = [
    "net-31e8c8f1-bca6-4ff9-af06-0f1d3624d158"
  ]

  ssh_key_ids = [
    "sk-b87aad9c-5ff8-4cd4-bd49-d50e9560194f"
  ]
}

resource "epilayer_load_balancer" "example" {
  name        = "my-load-balancer-1"
  region      = "NORD-NO-KRS-1"
  network     = "net-31e8c8f1-bca6-4ff9-af06-0f1d3624d158"
  mode        = "tcp"
  description = "This is a an example load balancer"

  ports = [
    {
      port        = 80
      target_port = 8080
      targets     = [epilayer_instance.example.private_ip]
    },
    {
      port        = 443
      target_port = 8443
      targets     = [epilayer_instance.example.private_ip]
    }
  ]

  health_check = {
    protocol = "http"
    port     = 8080
    path     = "/"
  }
}
