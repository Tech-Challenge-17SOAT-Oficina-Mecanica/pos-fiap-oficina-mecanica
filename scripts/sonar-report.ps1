param(
    [string]$HostUrl = $(if ($env:SONAR_HOST_URL) { $env:SONAR_HOST_URL } else { "http://localhost:9000" }),
    [string]$ProjectKey = $(if ($env:SONAR_PROJECT_KEY) { $env:SONAR_PROJECT_KEY } else { "oficina-mecanica" }),
    [string]$OutDir = "reports"
)

$ErrorActionPreference = "Stop"

if (-not $env:SONAR_TOKEN) {
    throw "Defina SONAR_TOKEN antes de executar o relatorio."
}

$HostUrl = $HostUrl.TrimEnd("/")
$metricKeys = "coverage,bugs,vulnerabilities,security_hotspots,code_smells,duplicated_lines_density,ncloc"
$headers = @{ Authorization = "Bearer $env:SONAR_TOKEN" }

function Get-SonarJson([string]$Path) {
    $uri = "$HostUrl$Path"
    Invoke-RestMethod -Method Get -Uri $uri -Headers $headers
}

function Get-MeasureValue($Measures, [string]$Metric, [string]$Default = "0") {
    $measure = $Measures | Where-Object { $_.metric -eq $Metric } | Select-Object -First 1
    if ($measure) { return $measure.value }
    return $Default
}

New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$encodedProject = [uri]::EscapeDataString($ProjectKey)
$measures = Get-SonarJson "/api/measures/component?component=$encodedProject&metricKeys=$metricKeys"
$qualityGate = Get-SonarJson "/api/qualitygates/project_status?projectKey=$encodedProject"
$issues = Get-SonarJson "/api/issues/search?componentKeys=$encodedProject&resolved=false&ps=1"

$values = $measures.component.measures
$generatedAt = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$report = [ordered]@{
    Projeto = $ProjectKey
    GeradoEm = $generatedAt
    QualityGate = $qualityGate.projectStatus.status
    Coverage = "$(Get-MeasureValue $values "coverage")%"
    Bugs = Get-MeasureValue $values "bugs"
    Vulnerabilidades = Get-MeasureValue $values "vulnerabilities"
    SecurityHotspots = Get-MeasureValue $values "security_hotspots"
    CodeSmells = Get-MeasureValue $values "code_smells"
    Duplicacao = "$(Get-MeasureValue $values "duplicated_lines_density")%"
    LinhasCodigo = Get-MeasureValue $values "ncloc"
    IssuesAbertas = $issues.total
}

$markdown = @(
    "# Relatorio SonarQube"
    ""
    "| Metrica | Valor |"
    "| --- | ---: |"
)

foreach ($item in $report.GetEnumerator()) {
    $markdown += "| $($item.Key) | $($item.Value) |"
}

$htmlRows = foreach ($item in $report.GetEnumerator()) {
    "<tr><th>$($item.Key)</th><td>$($item.Value)</td></tr>"
}

$html = @"
<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <title>Relatorio SonarQube</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 32px; color: #17202a; }
    table { border-collapse: collapse; min-width: 520px; }
    th, td { border: 1px solid #d5d8dc; padding: 10px 12px; }
    th { background: #f4f6f7; text-align: left; }
    td { text-align: right; }
  </style>
</head>
<body>
  <h1>Relatorio SonarQube</h1>
  <table>
    $($htmlRows -join "`n    ")
  </table>
</body>
</html>
"@

$mdPath = Join-Path $OutDir "sonar-report.md"
$htmlPath = Join-Path $OutDir "sonar-report.html"

$markdown | Set-Content -Encoding UTF8 -LiteralPath $mdPath
$html | Set-Content -Encoding UTF8 -LiteralPath $htmlPath

Write-Host "Relatorios gerados:"
Write-Host $mdPath
Write-Host $htmlPath
