resource "kind_cluster" "oficina" { // Significa: “Terraform, gerencie um cluster Kind.”
/*
  resource: declara algo que o Terraform vai gerenciar.
  kind_cluster: o tipo do recurso.
  oficina: o identificador desse recurso no código Terraform.
*/
  name            = "oficina" // nome real do cluster
  node_image      = "kindest/node:v1.32.2" // magem usada para executar seu nó
  wait_for_ready  = true // Pede ao Terraform que aguarde o cluster ficar pronto antes de considerar a criação concluída.
  kubeconfig_path = abspath("${path.module}/kubeconfig")
  # O provider marca esse caminho como ForceNew; mover a pasta nao deve recriar o cluster.
  lifecycle {
    ignore_changes = [kubeconfig_path]
  }
  /*
    kubeconfig_path Define onde salvar o kubeconfig, o arquivo com endereço e credenciais para acessar o cluster.
    - ${path.module} → diretório dessa configuração: infra/local/cluster
    - /kubeconfig → nome do arquivo que será gerado, sem extensão.
    - abspath(...): transforma o caminho em absoluto.
  */
  kind_config { // bloco que descreve a configuração do Kind
    kind        = "Cluster" // tipo de recurso que o Kind vai criar
    api_version = "kind.x-k8s.io/v1alpha4" // versão da API do Kind que será usada para criar o cluster
    node { // bloco que descreve um nó do cluster
      role = "control-plane" // função do nó no cluster. Pode ser control-plane ou worker.
      extra_mounts { // bloco que descreve um volume que será montado no nó
        host_path      = var.data_path // caminho no host (máquina que executa o Kind) que será montado no nó
                        // var.data_path é a variável definida em infra/local/cluster/variables.tf
        container_path = "/var/local/oficina-postgres" // caminho no nó (container) onde o volume será montado
      }
    }
  }
}

/*
  Etapas:
  1. O Terraform cria um cluster Kind chamado oficina.
  2. O Kind cria um container Docker chamado oficina-control-plane, que é o nó do cluster Kubernetes.
  3. O Kind monta o volume do host /home/usuario/oficina-data no container, no caminho /var/local/oficina-postgres.
*/
