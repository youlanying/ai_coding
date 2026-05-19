Write-Host "Testing Raft KV Cluster..." -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Gray

$baseUrl = "http://127.0.0.1:8001"

Write-Host ""
Write-Host "1. Checking cluster status..." -ForegroundColor Yellow

try {
    $status1 = Invoke-RestMethod -Uri "http://127.0.0.1:8001/status" -Method Get
    Write-Host "Node1: $($status1.state) (term: $($status1.term), leader: $($status1.leader))"
} catch {
    Write-Host "Node1: Not responding" -ForegroundColor Red
}

try {
    $status2 = Invoke-RestMethod -Uri "http://127.0.0.1:8002/status" -Method Get
    Write-Host "Node2: $($status2.state) (term: $($status2.term), leader: $($status2.leader))"
} catch {
    Write-Host "Node2: Not responding" -ForegroundColor Red
}

try {
    $status3 = Invoke-RestMethod -Uri "http://127.0.0.1:8003/status" -Method Get
    Write-Host "Node3: $($status3.state) (term: $($status3.term), leader: $($status3.leader))"
} catch {
    Write-Host "Node3: Not responding" -ForegroundColor Red
}

Write-Host ""
Write-Host "2. Testing SET operation..." -ForegroundColor Yellow

$setBody = @{
    key   = "name"
    value = "Raft KV Store"
} | ConvertTo-Json

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/set" -Method Post -Body $setBody -ContentType "application/json"
    Write-Host "SET 'name' = 'Raft KV Store': $($result.status)" -ForegroundColor Green
} catch {
    Write-Host "SET failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "3. Testing GET operation from all nodes..." -ForegroundColor Yellow

Start-Sleep -Milliseconds 200

try {
    $get1 = Invoke-RestMethod -Uri "http://127.0.0.1:8001/get/name" -Method Get
    Write-Host "Node1 GET 'name': '$($get1.value)' (found: $($get1.found))"
} catch {
    Write-Host "Node1 GET failed" -ForegroundColor Red
}

try {
    $get2 = Invoke-RestMethod -Uri "http://127.0.0.1:8002/get/name" -Method Get
    Write-Host "Node2 GET 'name': '$($get2.value)' (found: $($get2.found))"
} catch {
    Write-Host "Node2 GET failed" -ForegroundColor Red
}

try {
    $get3 = Invoke-RestMethod -Uri "http://127.0.0.1:8003/get/name" -Method Get
    Write-Host "Node3 GET 'name': '$($get3.value)' (found: $($get3.found))"
} catch {
    Write-Host "Node3 GET failed" -ForegroundColor Red
}

Write-Host ""
Write-Host "4. Testing multiple SET operations..." -ForegroundColor Yellow

$testData = @{
    "user"    = "alice"
    "email"   = "alice@example.com"
    "count"   = "42"
    "message" = "Hello Raft!"
}

foreach ($key in $testData.Keys) {
    $body = @{
        key   = $key
        value = $testData[$key]
    } | ConvertTo-Json

    try {
        $result = Invoke-RestMethod -Uri "$baseUrl/set" -Method Post -Body $body -ContentType "application/json"
        Write-Host "SET '$key' = '$($testData[$key])': $($result.status)" -ForegroundColor Green
    } catch {
        Write-Host "SET '$key' failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Start-Sleep -Milliseconds 300

Write-Host ""
Write-Host "5. Listing all keys on Node2..." -ForegroundColor Yellow

try {
    $list = Invoke-RestMethod -Uri "http://127.0.0.1:8002/list" -Method Get
    Write-Host "All data on Node2:" -ForegroundColor Cyan
    $list | ConvertTo-Json -Depth 10
} catch {
    Write-Host "LIST failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "6. Testing DELETE operation..." -ForegroundColor Yellow

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/delete/count" -Method Post
    Write-Host "DELETE 'count': $($result.status)" -ForegroundColor Green
} catch {
    Write-Host "DELETE failed: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Milliseconds 200

Write-Host ""
Write-Host "7. Verifying DELETE on all nodes..." -ForegroundColor Yellow

try {
    $get1 = Invoke-RestMethod -Uri "http://127.0.0.1:8001/get/count" -Method Get
    Write-Host "Node1 GET 'count': found=$($get1.found)"
} catch {
    Write-Host "Node1 GET failed" -ForegroundColor Red
}

try {
    $get2 = Invoke-RestMethod -Uri "http://127.0.0.1:8002/get/count" -Method Get
    Write-Host "Node2 GET 'count': found=$($get2.found)"
} catch {
    Write-Host "Node2 GET failed" -ForegroundColor Red
}

try {
    $get3 = Invoke-RestMethod -Uri "http://127.0.0.1:8003/get/count" -Method Get
    Write-Host "Node3 GET 'count': found=$($get3.found)"
} catch {
    Write-Host "Node3 GET failed" -ForegroundColor Red
}

Write-Host ""
Write-Host "8. Final data on Node3..." -ForegroundColor Yellow

try {
    $list = Invoke-RestMethod -Uri "http://127.0.0.1:8003/list" -Method Get
    Write-Host "All data on Node3:" -ForegroundColor Cyan
    $list | ConvertTo-Json -Depth 10
} catch {
    Write-Host "LIST failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "======================================" -ForegroundColor Gray
Write-Host "Test completed!" -ForegroundColor Green
Write-Host "Check ./data/ directory for persisted data files." -ForegroundColor Yellow
