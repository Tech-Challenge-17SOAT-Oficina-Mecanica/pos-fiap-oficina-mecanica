# A pipeline prepara o bucket e importa seus recursos no backend bootstrap remoto.
# Cada conta recebe seu proprio bucket; nunca reutilizar o estado de outra conta.
data "aws_caller_identity" "current" {}

locals {
  bucket_name = "oficina-tfstate-${data.aws_caller_identity.current.account_id}-us-east-1"
}

# O script cria o bucket. O recurso generico consulta Object Lock, proibido no Academy.
# Desvincula estados anteriores sem apagar o bucket nem seus objetos.
removed {
  from = aws_s3_bucket.terraform
  lifecycle {
    destroy = false
  }
}

resource "aws_s3_bucket_public_access_block" "terraform" {
  bucket                  = local.bucket_name
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "terraform" {
  bucket = local.bucket_name
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform" {
  bucket = local.bucket_name
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_ownership_controls" "terraform" {
  bucket = local.bucket_name
  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}
