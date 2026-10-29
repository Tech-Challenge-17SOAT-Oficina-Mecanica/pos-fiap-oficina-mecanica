output "bucket_name" {
  description = "Bucket exclusivo da conta autenticada; usar no backend dos estados AWS."
  value       = local.bucket_name
}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}
