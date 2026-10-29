output "bucket_name" {
  description = "Bucket exclusivo da conta autenticada; usar no backend dos estados AWS."
  value       = aws_s3_bucket.terraform.id
}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}
