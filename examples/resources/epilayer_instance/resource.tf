resource "epilayer_instance" "example" {
  name = "example"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}

# Example with private network
resource "epilayer_private_network" "app_network" {
  name    = "app-network"
  cidr_v4 = "10.0.0.0/24"
}

resource "epilayer_instance" "with_private_network" {
  name = "app-server"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  private_network_ids = [epilayer_private_network.app_network.id]

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}

# Example with Kubernetes cluster
resource "epilayer_kubernetes_cluster" "example" {
  name = "my-cluster"
}

resource "epilayer_instance" "with_k8s_cluster" {
  name = "k8s-node"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  k8s_cluster_id = epilayer_kubernetes_cluster.example.id

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}
