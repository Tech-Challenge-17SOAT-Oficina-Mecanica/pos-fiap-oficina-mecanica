data "aws_iam_role" "cluster" { name = var.cluster_role_name }
data "aws_iam_role" "node" { name = var.node_role_name }
data "aws_subnet" "selected" {
  for_each = var.subnet_ids
  id       = each.value
}

resource "aws_eks_cluster" "oficina" {
  name     = var.cluster_name
  role_arn = data.aws_iam_role.cluster.arn
  version  = var.kubernetes_version
  access_config {
    authentication_mode                         = "API_AND_CONFIG_MAP"
    bootstrap_cluster_creator_admin_permissions = true
  }
  upgrade_policy { support_type = "STANDARD" }
  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = true
    public_access_cidrs     = var.admin_cidrs
  }
  lifecycle {
    precondition {
      condition = (
        alltrue([for subnet in data.aws_subnet.selected : subnet.vpc_id == var.vpc_id && subnet.map_public_ip_on_launch]) &&
        length(toset([for subnet in data.aws_subnet.selected : subnet.availability_zone])) >= 2
      )
      error_message = "Use sub-redes da VPC informada, com IPv4 publico automatico e em duas AZs diferentes. Verifique tambem as rotas para Internet Gateway."
    }
  }
}

# shortcut: um no para demonstracao academica; nao oferece alta disponibilidade.
resource "aws_eks_node_group" "oficina" {
  cluster_name    = aws_eks_cluster.oficina.name
  node_group_name = "oficina"
  node_role_arn   = data.aws_iam_role.node.arn
  subnet_ids      = var.subnet_ids
  instance_types  = ["t3.medium"]
  capacity_type   = "ON_DEMAND"
  disk_size       = 20
  scaling_config {
    desired_size = 1
    min_size     = 1
    max_size     = 1
  }
}

resource "aws_ecr_repository" "api" {
  name         = var.ecr_repository
  force_delete = false
}

resource "aws_eks_addon" "metrics" {
  cluster_name = aws_eks_cluster.oficina.name
  addon_name   = "metrics-server"
  depends_on   = [aws_eks_node_group.oficina]
}
