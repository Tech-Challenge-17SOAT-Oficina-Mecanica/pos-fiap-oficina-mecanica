# Namespace compartilhado com k8s/local/namespace.yaml: Terraform e seu unico gestor.
resource "kubernetes_namespace_v1" "oficina" {
  metadata { name = "oficina" }
}

resource "kubernetes_secret_v1" "api" {
  metadata {
    name      = "oficina-api-secret"
    namespace = kubernetes_namespace_v1.oficina.metadata[0].name
  }
  data = {
    DB_USER     = "oficina"
    DB_PASSWORD = var.db_password
    JWT_SECRET  = var.jwt_secret
  }
}

resource "kubernetes_manifest" "postgres" {
  for_each   = toset(["pv", "pvc", "service", "statefulset"])
  manifest   = yamldecode(file("${path.module}/../../../k8s/local/database/${each.key}.yaml"))
  depends_on = [kubernetes_secret_v1.api]
}
