variable "cluster_name" {
  type    = string
  default = "oficina-academy"
}

variable "kubernetes_version" {
  description = "Confirmar disponibilidade e suporte padrao na regiao antes do apply."
  type        = string
  default     = "1.35"
}

variable "vpc_id" {
  type    = string
  default = "vpc-0ae7cb40d8e37d780"
}

variable "subnet_ids" {
  description = "Sub-redes publicas existentes da mesma VPC, em pelo menos duas AZs, com rota para Internet Gateway e IPv4 publico automatico."
  type        = set(string)
  validation {
    condition     = length(var.subnet_ids) >= 2
    error_message = "Informe pelo menos duas sub-redes."
  }
}

variable "cluster_role_name" {
  type    = string
  default = "c227945a5749351l17859999t1w568255-LabEksClusterRole-vlTG73cqhMRX"
}

variable "node_role_name" {
  type    = string
  default = "c227945a5749351l17859999t1w568255900-LabEksNodeRole-rqFbHhHxmCbN"
}

variable "admin_cidrs" {
  description = "IPs autorizados no endpoint Kubernetes; exemplo: IP publico do administrador/32."
  type        = set(string)
  validation {
    condition     = length(var.admin_cidrs) > 0 && alltrue([for cidr in var.admin_cidrs : can(cidrnetmask(cidr)) && cidr != "0.0.0.0/0"])
    error_message = "Informe CIDRs IPv4 restritos; nao use 0.0.0.0/0."
  }
}

variable "db_password" {
  description = "Senha do PostgreSQL; usar o mesmo valor em aws/platform. Nunca versionar."
  type        = string
  sensitive   = true
  validation {
    condition     = can(regex("^[A-Za-z0-9_-]{16,}$", var.db_password))
    error_message = "Use pelo menos 16 caracteres alfanumericos, _ ou -."
  }
}

variable "db_deletion_protection" {
  description = "Desativar explicitamente antes da remocao definitiva do laboratorio."
  type        = bool
  default     = true
}
