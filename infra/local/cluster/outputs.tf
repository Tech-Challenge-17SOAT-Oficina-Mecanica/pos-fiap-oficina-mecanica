// mostra uma informação útil ao final
// pede ao Terraform que mostre o caminho do kubeconfig depois do provisionamento.
output "kubeconfig_path" { // nome do output
  value = abspath("${path.module}/kubeconfig") // caminho atual do arquivo local, inclusive apos mover a pasta
}
