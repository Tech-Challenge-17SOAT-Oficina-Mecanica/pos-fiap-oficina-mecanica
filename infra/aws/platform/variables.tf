variable "cluster_name" {
  description = "Mesmo cluster usado na geracao do kubeconfig e do token EKS."
  type        = string
  default     = "oficina-academy"
}

variable "kubeconfig_path" {
  type    = string
  default = "~/.kube/config"
}

variable "db_host" {
  description = "Endereco DNS retornado por terraform output -raw db_host em aws/cluster."
  type        = string
  validation {
    condition     = endswith(var.db_host, ".rds.amazonaws.com")
    error_message = "Informe o endpoint DNS do RDS em us-east-1."
  }
}

variable "db_password" {
  type      = string
  sensitive = true
  validation {
    condition     = can(regex("^[A-Za-z0-9_-]{16,}$", var.db_password))
    error_message = "Use ao menos 16 caracteres alfanumericos, _ ou -."
  }
}

variable "jwt_secret" {
  type      = string
  sensitive = true
  validation {
    condition     = length(var.jwt_secret) >= 32
    error_message = "Use pelo menos 32 caracteres."
  }
}
