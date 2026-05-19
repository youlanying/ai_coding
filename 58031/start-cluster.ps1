Write-Host "Starting Raft KV Cluster..." -ForegroundColor Green

$ErrorActionPreference = "Stop"

Write-Host "Building..." -ForegroundColor Yellow
go build -o raft-kv.exe ./cmd

if (-not (Test-Path "./data")) {
    New-Item -ItemType Directory -Path "./data" | Out-Null
}

Write-Host "Starting Node 1 (Leader candidate) on port 8001..." -ForegroundColor Cyan
$node1 = Start-Process -FilePath "./raft-kv.exe" -ArgumentList "--id node1 --http 127.0.0.1:8001 --peers node2=127.0.0.1:8002,node3=127.0.0.1:8003 --data ./data" -PassThru -NoNewWindow

Write-Host "Starting Node 2 (Follower) on port 8002..." -ForegroundColor Cyan
$node2 = Start-Process -FilePath "./raft-kv.exe" -ArgumentList "--id node2 --http 127.0.0.1:8002 --peers node1=127.0.0.1:8001,node3=127.0.0.1:8003 --data ./data" -PassThru -NoNewWindow

Write-Host "Starting Node 3 (Follower) on port 8003..." -ForegroundColor Cyan
$node3 = Start-Process -FilePath "./raft-kv.exe" -ArgumentList "--id node3 --http 127.0.0.1:8003 --peers node1=127.0.0.1:8001,node2=127.0.0.1:8002 --data ./data" -PassThru -NoNewWindow

Write-Host ""
Write-Host "Cluster started!" -ForegroundColor Green
Write-Host "Node1: http://127.0.0.1:8001"
Write-Host "Node2: http://127.0.0.1:8002"
Write-Host "Node3: http://127.0.0.1:8003"
Write-Host ""
Write-Host "Press Ctrl+C to stop all nodes..." -ForegroundColor Yellow

try {
    while ($true) {
        Start-Sleep -Seconds 1
        if (-not $node1.IsRunning -or -not $node2.IsRunning -or -not $node3.IsRunning) {
            Write-Host "One or more nodes stopped, exiting..." -ForegroundColor Red
            break
        }
    }
} finally {
    Write-Host "Stopping all nodes..." -ForegroundColor Yellow
    if ($node1.IsRunning) { Stop-Process -Id $node1.Id -Force }
    if ($node2.IsRunning) { Stop-Process -Id $node2.Id -Force }
    if ($node3.IsRunning) { Stop-Process -Id $node3.Id -Force }
    Write-Host "All nodes stopped." -ForegroundColor Green
}
