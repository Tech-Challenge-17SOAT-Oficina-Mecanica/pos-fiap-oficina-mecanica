# A pipeline prepara o bucket e importa seus recursos no backend bootstrap remoto.
# Cada conta recebe seu proprio bucket; nunca reutilizar o estado de outra conta.
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "terraform" {
  bucket        = "oficina-tfstate-${data.aws_caller_identity.current.account_id}-us-east-1"
  force_destroy = false
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_public_access_block" "terraform" {
  bucket                  = aws_s3_bucket.terraform.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "terraform" {
  bucket = aws_s3_bucket.terraform.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform" {
  bucket = aws_s3_bucket.terraform.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_ownership_controls" "terraform" {
  bucket = aws_s3_bucket.terraform.id
  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}
