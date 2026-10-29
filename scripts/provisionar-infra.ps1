<#
.SYNOPSIS
Prepara a infraestrutura LOCAL da oficina: Kubernetes Kind e PostgreSQL persistente.
.DESCRIPTION
Requer Docker Desktop em modo Linux, Terraform e kubectl instalados.
Prepara a pasta dos dados e reutiliza credenciais existentes; quando nao existem,
gera e salva credenciais em infra/local/.local, ignorada pelo Git.
Executa init, validate, plan e apply de infra/local/cluster e infra/local/platform,
bloqueando planos que removem/substituem recursos ou trocam a senha do banco.
Ao final, aguarda o PostgreSQL ficar pronto. Este script ALTERA o ambiente local.
Nao provisiona AWS, nao executa migrations e nao instala API, Metrics Server ou HPA.
.PARAMETER DataPath
Caminho absoluto opcional para os dados do PostgreSQL. Sem ele, reutiliza a
configuracao/estado existente ou usa LOCALAPPDATA/oficina-mecanica/postgres.
Nao muda automaticamente o caminho de dados de um cluster ja existente.
.EXAMPLE
# Na raiz do projeto, pelo Git Bash ou PowerShell:
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/provisionar-infra.ps1
.EXAMPLE
# Para escolher o armazenamento na primeira preparacao:
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/provisionar-infra.ps1 -DataPath C:/oficina-data/postgres
.NOTES
Resultado esperado: PostgreSQL pronto e PVC Bound. Guarde as credenciais locais;
reexecutar o script deve reaproveita-las. Nao use este script para fazer deploy AWS.
#>
[CmdletBinding()]
param([string]$DataPath)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$clusterDir = Join-Path $repoRoot 'infra/local/cluster'
$platformDir = Join-Path $repoRoot 'infra/local/platform'
$localDir = Join-Path $repoRoot 'infra/local/.local'
$utf8 = New-Object System.Text.UTF8Encoding($false)

function Read-TerraformState([string]$Directory) {
    $stateFile = Join-Path $Directory 'terraform.tfstate'
    if (Test-Path -LiteralPath $stateFile) {
        return (Get-Content -LiteralPath $stateFile -Raw | ConvertFrom-Json)
    }
}

function Invoke-Checked([string]$Command, [string[]]$Arguments) {
    & $Command @Arguments | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "$Command falhou. Provisionamento interrompido." }
}

function Apply-Terraform([string]$Directory, [string[]]$Variables) {
    Invoke-Checked terraform (@("-chdir=$Directory", 'init', '-input=false') )
    Invoke-Checked terraform (@("-chdir=$Directory", 'validate'))
    $planFile = Join-Path $localDir ((Split-Path $Directory -Leaf) + '.tfplan')
    $planArguments = @("-chdir=$Directory", 'plan', '-input=false', '-no-color', "-out=$planFile") + $Variables
    $planText = & terraform @planArguments
    if ($LASTEXITCODE -ne 0) {
        $planText | Out-Host
        throw 'terraform falhou ao preparar o plano. Nenhum apply desta etapa foi executado.'
    }
    # O plano JSON pode conter segredos: analisar em memoria, nunca imprimir.
    $json = & terraform "-chdir=$Directory" show -json $planFile
    if ($LASTEXITCODE -ne 0) { throw 'Nao foi possivel conferir o plano Terraform.' }
    $plan = ($json -join "`n") | ConvertFrom-Json
    foreach ($change in $plan.resource_changes) {
        if ($change.change.actions -contains 'delete') {
            throw "O plano remove ou substitui $($change.address). Revise manualmente; nenhum apply desta etapa foi executado."
        }
        if ($change.type -eq 'kubernetes_secret_v1' -and $null -ne $change.change.before) {
            if ($change.change.before.data.DB_PASSWORD -cne $change.change.after.data.DB_PASSWORD) {
                throw 'O plano troca a senha do banco existente. Rotacao exige alterar a senha no PostgreSQL; revise manualmente.'
            }
        }
    }
    $creates = @($plan.resource_changes | Where-Object { $_.change.actions -contains 'create' }).Count
    $updates = @($plan.resource_changes | Where-Object { $_.change.actions -contains 'update' }).Count
    Write-Host "Plano conferido: $creates criar, $updates atualizar, 0 remover."
    Invoke-Checked terraform (@("-chdir=$Directory", 'apply', '-input=false', $planFile))
}

foreach ($command in @('docker', 'terraform', 'kubectl')) {
    if (-not (Get-Command $command -ErrorAction SilentlyContinue)) { throw "Instale $command antes de continuar." }
}
$dockerType = & docker info --format '{{.OSType}}'
if ($LASTEXITCODE -ne 0) { throw 'Docker indisponivel. Inicie o Docker Desktop e tente novamente.' }
if (($dockerType -join '').Trim() -ne 'linux') { throw 'Use o Docker Desktop em modo Linux containers.' }
$clusterState = Read-TerraformState $clusterDir
$clusterVars = Join-Path $clusterDir 'terraform.tfvars'
if (-not $DataPath) {
    if (Test-Path -LiteralPath $clusterVars) {
        $config = Get-Content -LiteralPath $clusterVars -Raw
        $match = [regex]::Match($config, '(?m)^\s*data_path\s*=\s*("(?:[^"\\]|\\.)*")')
        if (-not $match.Success) { throw 'Nao foi possivel ler data_path. Forneca -DataPath ou use um caminho literal em terraform.tfvars.' }
        $DataPath = $match.Groups[1].Value | ConvertFrom-Json
    } elseif ($clusterState) {
        $resource = @($clusterState.resources | Where-Object { $_.type -eq 'kind_cluster' -and $_.name -eq 'oficina' })
        if ($resource.Count -ne 1) { throw 'Estado existente sem o cluster esperado; revise antes de continuar.' }
        $DataPath = $resource[0].instances[0].attributes.kind_config[0].node[0].extra_mounts[0].host_path
    } else {
        if (-not $env:LOCALAPPDATA) { throw 'Informe -DataPath com um diretorio absoluto para os dados.' }
        $DataPath = Join-Path $env:LOCALAPPDATA 'oficina-mecanica/postgres'
    }
}
if (-not [IO.Path]::IsPathRooted($DataPath)) { throw 'DataPath deve ser um caminho absoluto.' }
$DataPath = [IO.Path]::GetFullPath($DataPath).Replace('\', '/')
if ($clusterState) {
    $resource = @($clusterState.resources | Where-Object { $_.type -eq 'kind_cluster' -and $_.name -eq 'oficina' })
    if ($resource.Count -eq 1) {
        $previousPath = $resource[0].instances[0].attributes.kind_config[0].node[0].extra_mounts[0].host_path
        if ([IO.Path]::GetFullPath($previousPath) -ine [IO.Path]::GetFullPath($DataPath)) {
            throw 'O cluster existente usa outro diretorio de dados. A migracao nao e automatica; nenhum recurso foi alterado.'
        }
    }
}
if (-not (Test-Path -LiteralPath $DataPath)) { New-Item -ItemType Directory -Path $DataPath -Force | Out-Null }
if (-not (Test-Path -LiteralPath $DataPath -PathType Container)) { throw 'O caminho de dados nao e um diretorio.' }
if (-not (Test-Path -LiteralPath $localDir)) { New-Item -ItemType Directory -Path $localDir -Force | Out-Null }
if (-not (Test-Path -LiteralPath $clusterVars)) {
    [IO.File]::WriteAllText($clusterVars, "data_path = $($DataPath | ConvertTo-Json)`n", $utf8)
}

$credentialFile = Join-Path $localDir 'platform.tfvars.json'
foreach ($candidate in @('terraform.tfvars', 'terraform.tfvars.json')) {
    $customVars = Join-Path $platformDir $candidate
    if (Test-Path -LiteralPath $customVars) { $credentialFile = $customVars; break }
}
if (-not (Test-Path -LiteralPath $credentialFile)) {
    $platformState = Read-TerraformState $platformDir
    $secret = @($platformState.resources | Where-Object { $_.type -eq 'kubernetes_secret_v1' -and $_.name -eq 'api' })
    if ($secret.Count -eq 1) {
        $credentials = $secret[0].instances[0].attributes.data
    } else {
        $hasData = $null -ne (Get-ChildItem -LiteralPath $DataPath -Force | Select-Object -First 1)
        if ($hasData -and (-not $env:TF_VAR_db_password -or -not $env:TF_VAR_jwt_secret)) {
            throw 'Ja existem dados, mas nao ha credenciais locais salvas. Forneca as credenciais anteriores por TF_VAR_db_password e TF_VAR_jwt_secret.'
        }
        $random = [Security.Cryptography.RandomNumberGenerator]::Create()
        try {
            $bytes = New-Object byte[] 32
            $random.GetBytes($bytes)
            $password = -join ($bytes | ForEach-Object { $_.ToString('x2') })
            $random.GetBytes($bytes)
            $jwt = -join ($bytes | ForEach-Object { $_.ToString('x2') })
        } finally { $random.Dispose() }
        $credentials = @{
            db_password = $(if ($env:TF_VAR_db_password) { $env:TF_VAR_db_password } else { $password })
            jwt_secret = $(if ($env:TF_VAR_jwt_secret) { $env:TF_VAR_jwt_secret } else { $jwt })
        }
    }
    $dbPassword = if ($credentials.DB_PASSWORD) { $credentials.DB_PASSWORD } else { $credentials.db_password }
    $jwtSecret = if ($credentials.JWT_SECRET) { $credentials.JWT_SECRET } else { $credentials.jwt_secret }
    if (-not $dbPassword -or -not $jwtSecret) { throw 'Credenciais incompletas. Nenhuma senha nova foi aplicada.' }
    [IO.File]::WriteAllText($credentialFile, (@{ db_password = $dbPassword; jwt_secret = $jwtSecret } | ConvertTo-Json), $utf8)
}

Write-Host "Armazenamento persistente: $DataPath"
Write-Host '1/2 - Provisionando ou conferindo o cluster existente...'
Apply-Terraform $clusterDir @("-var=data_path=$DataPath")
Write-Host '2/2 - Provisionando os recursos do banco...'
Apply-Terraform $platformDir @("-var-file=$credentialFile")
$kubeconfig = Join-Path $clusterDir 'kubeconfig'
try {
    Invoke-Checked kubectl @('--kubeconfig', $kubeconfig, '--context', 'kind-oficina', '-n', 'oficina', 'rollout', 'status', 'statefulset/postgres', '--timeout=300s')
} catch {
    Write-Host 'O Terraform concluiu, mas PostgreSQL ainda nao ficou pronto. Dados e credenciais foram preservados.'
    Write-Host 'Consulte: kubectl --kubeconfig infra/local/cluster/kubeconfig --context kind-oficina -n oficina get events'
    throw
}
Invoke-Checked kubectl @('--kubeconfig', $kubeconfig, '--context', 'kind-oficina', '-n', 'oficina', 'get', 'pods,pvc')
Write-Host 'Infraestrutura preparada. As tabelas (migrations), API e HPA sao as proximas etapas.'
