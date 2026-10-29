mock_provider "aws" {
  mock_data "aws_caller_identity" {
    defaults = { account_id = "123456789012" }
  }
}

run "private_account_state" {
  command = plan
  assert {
    condition     = aws_s3_bucket.terraform.bucket == "oficina-tfstate-123456789012-us-east-1"
    error_message = "O bucket deve identificar a conta autenticada."
  }
  assert {
    condition = (
      aws_s3_bucket_public_access_block.terraform.block_public_acls &&
      aws_s3_bucket_public_access_block.terraform.block_public_policy &&
      aws_s3_bucket_public_access_block.terraform.ignore_public_acls &&
      aws_s3_bucket_public_access_block.terraform.restrict_public_buckets
    )
    error_message = "O bucket de estados nao pode permitir acesso publico."
  }
  assert {
    condition     = aws_s3_bucket_versioning.terraform.versioning_configuration[0].status == "Enabled" && !aws_s3_bucket.terraform.force_destroy
    error_message = "Preservar versoes e nao permitir apagar objetos automaticamente."
  }
  assert {
    condition     = one(aws_s3_bucket_server_side_encryption_configuration.terraform.rule).apply_server_side_encryption_by_default[0].sse_algorithm == "AES256"
    error_message = "Os estados precisam de criptografia em repouso."
  }
}
