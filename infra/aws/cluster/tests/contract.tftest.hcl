mock_provider "aws" {
  mock_data "aws_iam_role" {
    defaults = { arn = "arn:aws:iam::123456789012:role/existing-lab-role" }
  }
}

variables {
  vpc_id            = "vpc-test"
  cluster_role_name = "test-cluster-role"
  node_role_name    = "test-node-role"
  subnet_ids        = ["subnet-a", "subnet-b"]
  admin_cidrs       = ["203.0.113.10/32"]
  db_password       = "fake-test-only-password-123456"
}

override_data {
  target = data.aws_subnet.selected["subnet-a"]
  values = {
    vpc_id                  = "vpc-test"
    availability_zone       = "us-east-1a"
    map_public_ip_on_launch = true
  }
}
override_data {
  target = data.aws_subnet.selected["subnet-b"]
  values = {
    vpc_id                  = "vpc-test"
    availability_zone       = "us-east-1b"
    map_public_ip_on_launch = true
  }
}

run "academic_cluster" {
  command = plan
  assert {
    condition     = aws_eks_node_group.oficina.scaling_config[0].max_size == 1
    error_message = "O laboratorio deve limitar o grupo a uma maquina."
  }
  assert {
    condition     = aws_eks_cluster.oficina.vpc_config[0].endpoint_private_access
    error_message = "Os nos devem acessar o endpoint privado."
  }
  assert {
    condition     = !aws_db_instance.oficina.publicly_accessible && aws_db_instance.oficina.storage_encrypted
    error_message = "O banco deve ser privado e ter armazenamento criptografado."
  }
  assert {
    condition     = aws_vpc_security_group_ingress_rule.database_from_eks.from_port == 5432 && aws_vpc_security_group_ingress_rule.database_from_eks.cidr_ipv4 == null && aws_vpc_security_group_ingress_rule.database_from_eks.cidr_ipv6 == null
    error_message = "O banco so deve aceitar PostgreSQL do security group EKS, sem CIDR aberto."
  }
  assert {
    condition     = aws_db_instance.oficina.deletion_protection && !aws_db_instance.oficina.skip_final_snapshot && aws_db_instance.oficina.storage_type == "gp2"
    error_message = "Preservar protecao contra exclusao, snapshot final e armazenamento permitido pelo Academy."
  }
}

run "reject_public_endpoint_for_everyone" {
  command = plan
  variables { admin_cidrs = ["0.0.0.0/0"] }
  expect_failures = [var.admin_cidrs]
}

run "reject_single_subnet" {
  command = plan
  variables { subnet_ids = ["subnet-a"] }
  expect_failures = [var.subnet_ids]
}
