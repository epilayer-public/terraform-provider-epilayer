# Fetch an existing Kubernetes cluster by ID
data "epilayer_kubernetes_cluster" "example" {
  id = "my-k8s-cluster"
}

# Use the kubeconfig to configure the Kubernetes provider
provider "kubernetes" {
  host                   = yamldecode(data.epilayer_kubernetes_cluster.example.kubeconfig).clusters[0].cluster.server
  cluster_ca_certificate = base64decode(yamldecode(data.epilayer_kubernetes_cluster.example.kubeconfig).clusters[0].cluster.certificate-authority-data)
  client_certificate     = base64decode(yamldecode(data.epilayer_kubernetes_cluster.example.kubeconfig).users[0].user.client-certificate-data)
  client_key             = base64decode(yamldecode(data.epilayer_kubernetes_cluster.example.kubeconfig).users[0].user.client-key-data)
}

# Output cluster information
output "cluster_status" {
  value = data.epilayer_kubernetes_cluster.example.status
}

output "kubeconfig" {
  value     = data.epilayer_kubernetes_cluster.example.kubeconfig
  sensitive = true
}
