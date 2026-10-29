// O Terraform precisa saber conversar com o Kind. Para isso, utiliza um provider, um plugin que implementa essa integração.
terraform { // bloco que define configurações do Terraform
  required_version = ">= 1.6, < 2.0" // versão mínima do Terraform que será usada para executar essa configuração
  required_providers {
    kind = {
      source  = "tehcyx/kind" // qual provider será utilizado.
      version = "0.11.0" // versão do provider que será utilizada.
    }
  }
}
