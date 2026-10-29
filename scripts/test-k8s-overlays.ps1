<#
.SYNOPSIS
Confere os manifestos Kubernetes dos ambientes local e AWS sem executar deploy.
.DESCRIPTION
Requer kubectl instalado e usa somente kubectl kustomize para montar os arquivos.
Verifica API, Service, HPA e a separacao entre ambientes: imagem local no Kind,
imagem ECR na AWS e ausencia de recursos que pertencem ao Terraform ou de tarefas
opcionais no deploy da API. Na AWS, nao deve haver ConfigMap que substitua o do RDS.
Nao conecta a clusters, nao aplica manifestos e nao cria recursos na AWS.
.EXAMPLE
# Na raiz do projeto, pelo Git Bash ou PowerShell:
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/test-k8s-overlays.ps1
.NOTES
Resultado esperado: quatro mensagens PASS, para API e migrations de cada ambiente.
Valida a composicao dos arquivos; nao comprova a execucao da API nem se a imagem
existe no ECR. O SHA da imagem AWS ainda deve ser preenchido antes do deploy.
#>
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot

function Assert([bool]$Condition, [string]$Message) {
    if (-not $Condition) { throw "TESTE FALHOU: $Message" }
}

foreach ($environment in @('local', 'aws')) {
    $rendered = & kubectl kustomize (Join-Path $repoRoot "k8s/$environment")
    if ($LASTEXITCODE -ne 0) { throw "Falha ao renderizar $environment" }
    $yaml = $rendered -join "`n"
    $kinds = @([regex]::Matches($yaml, '(?m)^kind: ([^\r\n]+)') | ForEach-Object { $_.Groups[1].Value })
    Assert ($kinds -contains 'Deployment' -and $kinds -contains 'Service' -and $kinds -contains 'HorizontalPodAutoscaler') 'API, acesso e HPA precisam estar presentes.'
    Assert (-not ($kinds | Where-Object { $_ -in @('Namespace','Secret','StatefulSet','PersistentVolume','PersistentVolumeClaim','Job') })) 'Nao incluir recursos do Terraform ou diagnosticos no deploy da API.'
    Assert ($yaml -match 'name: cpu' -and $yaml -match 'name: memory') 'Preservar as duas metricas do HPA.'
    Assert ($yaml -notmatch 'kubelet-insecure-tls') 'Nao misturar instalacao do Metrics Server com deploy da API.'
    if ($environment -eq 'local') {
        Assert ($yaml -notmatch 'type: LoadBalancer') 'Local deve manter acesso interno.'
        Assert ($kinds.Count -eq 4 -and $kinds -contains 'ConfigMap') 'Local precisa das configuracoes compartilhadas.'
        Assert ($yaml -match 'image: oficina-api:local') 'Local deve usar a imagem carregada no Kind.'
        Assert ($yaml -match 'DB_HOST: postgres') 'Local deve usar o Service PostgreSQL.'
    } else {
        Assert ($yaml -match 'type: LoadBalancer' -and $yaml -match 'aws-load-balancer-type: nlb') 'AWS deve expor a API com NLB.'
        Assert ($yaml -match 'aws-load-balancer-internal: "false"') 'NLB deve ser publico.'
        Assert ($kinds.Count -eq 3 -and $kinds -notcontains 'ConfigMap') 'AWS deve preservar o ConfigMap do RDS gerenciado pelo Terraform.'
        Assert ($yaml -match 'image: example\.invalid/oficina-mecanica:') 'AWS deve ter um marcador sem conta fixa, substituido pelo deploy.'
        Assert ($yaml -notmatch 'image: oficina-api:local') 'AWS nao pode usar imagem exclusiva do Kind.'
    }
    Write-Host "PASS: overlay $environment renderizado e responsabilidades preservadas."

    $migrationOutput = & kubectl kustomize (Join-Path $repoRoot "k8s/$environment/migrations")
    if ($LASTEXITCODE -ne 0) { throw "Falha ao renderizar migrations de $environment" }
    $migrationYaml = $migrationOutput -join "`n"
    Assert ($migrationYaml -match 'kind: Job' -and $migrationYaml -match 'name: oficina-migrations') 'Preservar Job Flyway.'
    Assert ($migrationYaml -match 'secretKeyRef:' -and $migrationYaml -notmatch 'kind: Secret') 'Migrations devem consumir, nunca recriar credenciais.'
    $urlPosition = $migrationYaml.IndexOf('name: FLYWAY_URL')
    foreach ($dependency in @('DB_HOST', 'DB_PORT', 'DB_NAME')) {
        $dependencyPosition = $migrationYaml.IndexOf("name: $dependency")
        Assert ($dependencyPosition -ge 0 -and $dependencyPosition -lt $urlPosition) "$dependency deve aparecer antes de FLYWAY_URL para expansao pelo Kubernetes."
    }
    if ($environment -eq 'aws') {
        Assert ($migrationYaml -match 'sslmode=verify-full' -and $migrationYaml -match 'sslrootcert=/flyway/certs/aws-rds-global-bundle.pem') 'Flyway na AWS deve verificar certificado e identidade do RDS.'
    } else {
        Assert ($migrationYaml -notmatch 'sslmode=verify-full') 'Nao exigir certificado RDS do banco local.'
    }
    Write-Host "PASS: migrations de $environment renderizadas."
}
