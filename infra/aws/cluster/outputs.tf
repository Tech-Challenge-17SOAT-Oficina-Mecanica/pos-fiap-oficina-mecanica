output "cluster_name" { value = aws_eks_cluster.oficina.name }
output "db_host" { value = aws_db_instance.oficina.address }
output "configure_kubectl" {
  value = "aws eks update-kubeconfig --region us-east-1 --name ${aws_eks_cluster.oficina.name} --alias oficina-aws"
}
