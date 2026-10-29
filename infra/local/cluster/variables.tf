variable "data_path" { // valor fornecido em terraform.tfvars
  description = "Diretorio absoluto existente no host, compartilhado com Docker, para os dados do PostgreSQL."
  type        = string // tipo da variável
  validation { // validação do valor da variável
    condition     = can(regex("^(/|[A-Za-z]:[/\\\\])", var.data_path))
    error_message = "Informe um caminho absoluto; no Windows use barras / (C:/...)."
  }
}

// Declarar a variável não define seu valor.
// É como criar um campo em um formulário que ainda precisa ser preenchido.

/*
  variables.tf              terraform.tfvars
  Declara data_path         Define seu valor
            \                /
            ↓              ↓
                  main.tf
        Usa var.data_path na montagem
*/
