<#
.SYNOPSIS
Testa a automacao local de provisionar-infra.ps1 sem criar infraestrutura real.
.DESCRIPTION
Substitui docker, terraform e kubectl por simulacoes dentro deste processo.
Verifica preparacao inicial, preservacao/recuperacao de credenciais, protecao do
caminho dos dados, bloqueio de planos destrutivos/troca de senha e tratamento de
falhas. Nao exige Docker iniciado nem acessa AWS, Kubernetes ou bancos reais.
Cria arquivos ficticios em pastas temporarias e restaura as variaveis de ambiente
alteradas pelo teste. As pastas temporarias ficam disponiveis para inspecao.
.EXAMPLE
# Na raiz do projeto, pelo Git Bash ou PowerShell:
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/test-provisionar-infra.ps1
.NOTES
Resultado esperado: oito mensagens PASS. Uma falha interrompe o script com erro.
Esses testes verificam a automacao; nao comprovam que o cluster real funciona.
#>
$ErrorActionPreference = 'Stop'
$testRoot = Join-Path ([IO.Path]::GetTempPath()) ('oficina-infra-test-' + [guid]::NewGuid().ToString('N'))
$source = Join-Path $PSScriptRoot 'provisionar-infra.ps1'
$savedEnvironment = @{}
foreach ($name in @('LOCALAPPDATA', 'TF_VAR_db_password', 'TF_VAR_jwt_secret')) {
    $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}
$global:OficinaTestCalls = New-Object 'System.Collections.Generic.List[string]'
$global:OficinaTestActions = @('create')
$global:OficinaTestFailApply = $false
$global:OficinaTestPasswordChange = $false
$global:OficinaTestKubeFailure = $false
function global:docker { $global:LASTEXITCODE = 0; 'linux' }
function global:kubectl {
    $global:LASTEXITCODE = 0
    $global:OficinaTestCalls.Add('kubectl')
    if ($args -contains 'rollout' -and $global:OficinaTestKubeFailure) { $global:LASTEXITCODE = 1 }
}
function global:terraform {
    $global:LASTEXITCODE = 0
    $global:OficinaTestCalls.Add(($args -join ' '))
    if ($args -contains 'show') {
        $change = @{
            address = 'test.resource'; type = 'kind_cluster'
            change = @{ actions = $global:OficinaTestActions }
        }
        if ($global:OficinaTestPasswordChange -and $args[0] -match 'platform$') {
            $change.type = 'kubernetes_secret_v1'
            $change.change = @{ actions = @('update'); before = @{ data = @{ DB_PASSWORD = 'anterior' } }; after = @{ data = @{ DB_PASSWORD = 'outra' } } }
        }
        @{ resource_changes = @($change) } | ConvertTo-Json -Depth 8 -Compress
    }
    if ($args -contains 'apply' -and $global:OficinaTestFailApply) { $global:LASTEXITCODE = 1 }
}
function New-Fixture([string]$Name) {
    $root = Join-Path $testRoot $Name
    foreach ($directory in @('scripts', 'infra/local/cluster', 'infra/local/platform')) {
        New-Item -ItemType Directory -Path (Join-Path $root $directory) -Force | Out-Null
    }
    Copy-Item -LiteralPath $source -Destination (Join-Path $root 'scripts/provisionar-infra.ps1')
    $global:OficinaTestCalls.Clear()
    $global:OficinaTestActions = @('create')
    $global:OficinaTestFailApply = $false
    $global:OficinaTestPasswordChange = $false
    $global:OficinaTestKubeFailure = $false
    return $root
}
function Assert([bool]$Condition, [string]$Message) {
    if (-not $Condition) { throw "TESTE FALHOU: $Message" }
}
function Expect-Failure([scriptblock]$Action, [string]$Pattern) {
    $message = $null
    try { & $Action 6>$null } catch { $message = $_.Exception.Message }
    Assert ($null -ne $message -and $message -match $Pattern) "Falha esperada: $Pattern"
}
try {
    $env:LOCALAPPDATA = Join-Path $testRoot 'Local App Data'
    Remove-Item Env:TF_VAR_db_password -ErrorAction SilentlyContinue
    Remove-Item Env:TF_VAR_jwt_secret -ErrorAction SilentlyContinue
    $fixture = New-Fixture 'fresh'
    & (Join-Path $fixture 'scripts/provisionar-infra.ps1') 6>$null
    $data = Join-Path $env:LOCALAPPDATA 'oficina-mecanica/postgres'
    Assert (Test-Path -LiteralPath $data -PathType Container) 'Criar o diretorio automaticamente.'
    $credentialsPath = Join-Path $fixture 'infra/local/.local/platform.tfvars.json'
    $credentials = Get-Content -LiteralPath $credentialsPath -Raw | ConvertFrom-Json
    Assert ($credentials.db_password -match '^[0-9a-f]{64}$') 'Gerar senha aleatoria compativel com DB_*.'
    Assert ($credentials.jwt_secret.Length -eq 64) 'Gerar chave JWT.'
    $hash = (Get-FileHash -LiteralPath $credentialsPath).Hash
    $env:TF_VAR_db_password = 'nao-trocar-esta-senha'
    $env:TF_VAR_jwt_secret = 'nao-trocar-chave-existente-1234567890'
    & (Join-Path $fixture 'scripts/provisionar-infra.ps1') 6>$null
    Assert ((Get-FileHash -LiteralPath $credentialsPath).Hash -eq $hash) 'Reexecucao preserva credenciais salvas mesmo com variaveis diferentes no terminal.'
    Remove-Item Env:TF_VAR_db_password, Env:TF_VAR_jwt_secret
    Write-Host 'PASS: preparacao inicial e reexecucao preservando credenciais.'

    $fixture = New-Fixture 'replacement'
    $global:OficinaTestActions = @('delete', 'create')
    Expect-Failure { & (Join-Path $fixture 'scripts/provisionar-infra.ps1') -DataPath (Join-Path $fixture 'data') } 'remove ou substitui'
    Assert (@($global:OficinaTestCalls | Where-Object { $_ -match ' apply ' }).Count -eq 0) 'Nao aplicar plano que substitui cluster.'
    Write-Host 'PASS: plano destrutivo bloqueado antes do apply.'

    $fixture = New-Fixture 'existing-data'
    $data = Join-Path $fixture 'data'
    New-Item -ItemType Directory -Path $data | Out-Null
    Set-Content -LiteralPath (Join-Path $data 'PG_VERSION') -Value '16'
    Expect-Failure { & (Join-Path $fixture 'scripts/provisionar-infra.ps1') -DataPath $data } 'Ja existem dados'
    Assert (-not (Test-Path (Join-Path $fixture 'infra/local/.local/platform.tfvars.json'))) 'Nao inventar credenciais para dados existentes.'
    Write-Host 'PASS: dados existentes sem credenciais impedem geracao de nova senha.'

    $fixture = New-Fixture 'failure'
    $global:OficinaTestFailApply = $true
    Expect-Failure { & (Join-Path $fixture 'scripts/provisionar-infra.ps1') -DataPath (Join-Path $fixture 'data') } 'terraform falhou'
    Assert (@($global:OficinaTestCalls | Where-Object { $_ -match 'platform.* init' }).Count -eq 0) 'Nao iniciar platform se cluster falhar.'
    Write-Host 'PASS: falha no cluster interrompe as etapas seguintes.'

    $fixture = New-Fixture 'preserve-path'
    $oldPath = Join-Path $fixture 'original'
    @{ resources = @(@{ type = 'kind_cluster'; name = 'oficina'; instances = @(@{ attributes = @{
        kind_config = @(@{ node = @(@{ extra_mounts = @(@{ host_path = $oldPath }) }) })
    } }) }) } | ConvertTo-Json -Depth 15 | Set-Content -LiteralPath (Join-Path $fixture 'infra/local/cluster/terraform.tfstate')
    Expect-Failure { & (Join-Path $fixture 'scripts/provisionar-infra.ps1') -DataPath (Join-Path $fixture 'outro') } 'outro diretorio'
    Assert ($global:OficinaTestCalls.Count -eq 0) 'Nao executar Terraform para mudar caminho de cluster existente.'
    Write-Host 'PASS: caminho persistente existente protegido.'

    $fixture = New-Fixture 'recover-credentials'
    $data = Join-Path $fixture 'data'
    New-Item -ItemType Directory -Path $data | Out-Null
    Set-Content -LiteralPath (Join-Path $data 'PG_VERSION') -Value '16'
    @{ resources = @(@{ type = 'kubernetes_secret_v1'; name = 'api'; instances = @(@{ attributes = @{
        data = @{ DB_PASSWORD = 'senha-anterior-recuperada'; JWT_SECRET = 'chave-anterior-recuperada-1234567890' }
    } }) }) } | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $fixture 'infra/local/platform/terraform.tfstate')
    & (Join-Path $fixture 'scripts/provisionar-infra.ps1') -DataPath $data 6>$null
    $restored = Get-Content (Join-Path $fixture 'infra/local/.local/platform.tfvars.json') -Raw | ConvertFrom-Json
    Assert ($restored.db_password -ceq 'senha-anterior-recuperada') 'Reutilizar a senha do estado existente.'
    Write-Host 'PASS: credenciais recuperadas do estado sem rotacao.'

    $fixture = New-Fixture 'password-change'
    $global:OficinaTestPasswordChange = $true
    Expect-Failure { & (Join-Path $fixture 'scripts/provisionar-infra.ps1') -DataPath (Join-Path $fixture 'data') } 'troca a senha'
    Assert (@($global:OficinaTestCalls | Where-Object { $_ -match 'platform.* apply ' }).Count -eq 0) 'Nao aplicar troca de senha de PostgreSQL existente.'
    Write-Host 'PASS: troca de senha bloqueada antes do apply do banco.'

    $fixture = New-Fixture 'readiness-failure'
    $global:OficinaTestKubeFailure = $true
    Expect-Failure { & (Join-Path $fixture 'scripts/provisionar-infra.ps1') -DataPath (Join-Path $fixture 'data') } 'kubectl falhou'
    Assert (Test-Path (Join-Path $fixture 'infra/local/.local/platform.tfvars.json')) 'Preservar credenciais se o pod nao ficar pronto.'
    Write-Host 'PASS: erro de prontidao nao e reportado como provisionamento concluido.'
} finally {
    foreach ($name in $savedEnvironment.Keys) { [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process') }
    foreach ($name in @('docker', 'terraform', 'kubectl')) { Remove-Item -LiteralPath "Function:\global:$name" -ErrorAction SilentlyContinue }
    # Cada fixture fica em um diretorio temporario exclusivo; nao apagar recursivamente no teste.
    Write-Host "Fixtures isoladas para inspecao: $testRoot"
}
