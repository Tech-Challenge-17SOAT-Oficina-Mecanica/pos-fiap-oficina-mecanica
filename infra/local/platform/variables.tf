variable "db_password" {
  description = "Senha local do PostgreSQL, fornecida por TF_VAR_db_password."
  type        = string
  sensitive   = true
  validation {
    condition     = can(regex("^[A-Za-z0-9_-]{16,}$", var.db_password))
    error_message = "Use ao menos 16 caracteres alfanumericos, _ ou - (compatibilidade com a URL DB_* da API)."
  }
}

variable "jwt_secret" {
  description = "Chave JWT compartilhada pelas replicas, fornecida por TF_VAR_jwt_secret."
  type        = string
  sensitive   = true
  validation {
    condition     = length(var.jwt_secret) >= 32
    error_message = "Use pelo menos 32 caracteres."
  }
}
