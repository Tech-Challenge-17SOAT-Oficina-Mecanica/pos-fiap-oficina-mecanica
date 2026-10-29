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

resource "kubernetes_config_map_v1" "api" {
  metadata {
    name      = "oficina-api-config"
    namespace = kubernetes_namespace_v1.oficina.metadata[0].name
  }
  data = merge(yamldecode(file("${path.module}/../../../k8s/base/config/configmap.yaml")).data, {
    DB_HOST        = var.db_host
    DB_SSLMODE     = "verify-full"
    DB_SSLROOTCERT = "/etc/ssl/certs/aws-rds-global-bundle.pem"
  })
}
