# Plan-only: nao cria cluster e nao exige Docker iniciado.
run "local_cluster_and_database_contract" {
  command = plan
  variables {
    data_path = "/tmp/oficina-postgres"
  }
  assert {
    condition     = kind_cluster.oficina.name == "oficina" && kind_cluster.oficina.wait_for_ready
    error_message = "O cluster deve usar nome/contexto esperado e aguardar prontidao."
  }
  assert {
    condition = (
      yamldecode(file("${path.module}/../../../k8s/local/database/pvc.yaml")).spec.volumeName ==
      yamldecode(file("${path.module}/../../../k8s/local/database/pv.yaml")).metadata.name
    )
    error_message = "O PVC deve selecionar o PV persistente do projeto."
  }
  assert {
    condition = (
      yamldecode(file("${path.module}/../../../k8s/local/database/statefulset.yaml")).spec.template.spec.volumes[0].persistentVolumeClaim.claimName ==
      yamldecode(file("${path.module}/../../../k8s/local/database/pvc.yaml")).metadata.name
    )
    error_message = "O banco deve montar o PVC esperado."
  }
  assert {
    condition = (
      yamldecode(file("${path.module}/../../../k8s/local/database/service.yaml")).spec.selector ==
      yamldecode(file("${path.module}/../../../k8s/local/database/statefulset.yaml")).spec.template.metadata.labels
    )
    error_message = "O Service precisa selecionar os pods do PostgreSQL."
  }
  assert {
    condition = (
      yamldecode(file("${path.module}/../../../k8s/base/config/configmap.yaml")).data.DB_HOST ==
      yamldecode(file("${path.module}/../../../k8s/local/database/service.yaml")).metadata.name
    )
    error_message = "O host do banco da API deve corresponder ao Service provisionado."
  }
}

run "reject_relative_data_path" {
  command = plan
  variables {
    data_path = "relative/data"
  }
  expect_failures = [var.data_path]
}
