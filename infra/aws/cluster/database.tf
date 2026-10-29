resource "aws_db_subnet_group" "oficina" {
  name       = "${var.cluster_name}-postgres"
  subnet_ids = var.subnet_ids
}

resource "aws_security_group" "database" {
  name_prefix = "${var.cluster_name}-postgres-"
  description = "PostgreSQL acessivel somente a partir do grupo das maquinas EKS"
  vpc_id      = var.vpc_id
}

# Os nos gerenciados usam o security group do cluster quando nao ha launch template.
resource "aws_vpc_security_group_ingress_rule" "database_from_eks" {
  security_group_id            = aws_security_group.database.id
  referenced_security_group_id = aws_eks_cluster.oficina.vpc_config[0].cluster_security_group_id
  ip_protocol                  = "tcp"
  from_port                    = 5432
  to_port                      = 5432
}

resource "aws_db_parameter_group" "oficina" {
  name_prefix = "${var.cluster_name}-"
  family      = "postgres16"
  parameter {
    name  = "rds.force_ssl"
    value = "1"
  }
}

# shortcut: Single-AZ e usuario administrador compartilhado pela API neste laboratorio;
# em producao, usar usuario de aplicacao com privilegios limitados e avaliar Multi-AZ.
resource "aws_db_instance" "oficina" {
  identifier                 = "${var.cluster_name}-postgres"
  engine                     = "postgres"
  engine_version             = "16"
  engine_lifecycle_support   = "open-source-rds-extended-support-disabled"
  instance_class             = "db.t3.micro"
  allocated_storage          = 20
  storage_type               = "gp2"
  storage_encrypted          = true
  db_name                    = "oficina"
  username                   = "oficina"
  password                   = var.db_password
  port                       = 5432
  db_subnet_group_name       = aws_db_subnet_group.oficina.name
  vpc_security_group_ids     = [aws_security_group.database.id]
  parameter_group_name       = aws_db_parameter_group.oficina.name
  publicly_accessible        = false
  multi_az                   = false
  monitoring_interval        = 0
  backup_retention_period    = 1
  auto_minor_version_upgrade = true
  deletion_protection        = var.db_deletion_protection
  skip_final_snapshot        = false
  final_snapshot_identifier  = "${var.cluster_name}-postgres-final"
  copy_tags_to_snapshot      = true
}
