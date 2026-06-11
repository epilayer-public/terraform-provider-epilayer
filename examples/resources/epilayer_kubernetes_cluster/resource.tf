# Create a private network for the Kubernetes cluster
resource "epilayer_private_network" "cluster_network" {
  name    = "k8s-network"
  region  = "NORD-NO-KRS-1"
  cidr_v4 = "10.1.0.0/24"
}

# Create a Kubernetes cluster with a private network
resource "epilayer_kubernetes_cluster" "example" {
  name    = "my-k8s-cluster"
  network = epilayer_private_network.cluster_network.id
}

# Access cluster credentials via the data source
data "epilayer_kubernetes_cluster" "example" {
  id = epilayer_kubernetes_cluster.example.id
}

# Output the kubeconfig (mark as sensitive)
output "kubeconfig" {
  value     = data.epilayer_kubernetes_cluster.example.kubeconfig
  sensitive = true
}
