$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$protoDir = Join-Path $root "proto"
$outDir = Join-Path $root "pb"

if (!(Test-Path $protoDir)) {
    throw "proto directory not found: $protoDir"
}

if (!(Test-Path $outDir)) {
    New-Item -ItemType Directory -Path $outDir | Out-Null
}

Get-ChildItem -Path $outDir -Filter "*.pb.go" -File -ErrorAction SilentlyContinue | Remove-Item -Force

$protoFiles = Get-ChildItem -Path $protoDir -Filter "*.proto" -File
if ($protoFiles.Count -eq 0) {
    throw "no .proto files found in $protoDir"
}

$protoFilePaths = @()
foreach ($file in $protoFiles) {
    $protoFilePaths += $file.FullName
}

protoc `
    --proto_path="$protoDir" `
    --go_out="$outDir" `
    --go_opt=paths=source_relative `
    --go-grpc_out="$outDir" `
    --go-grpc_opt=paths=source_relative `
    $protoFilePaths

Write-Host "proto generated into $outDir"
