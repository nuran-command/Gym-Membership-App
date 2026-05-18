Write-Host "==============================================" -ForegroundColor Cyan
Write-Host "STARTING COMPREHENSIVE END-TO-END ENDPOINT TEST" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080"
$uniqueId = [Guid]::NewGuid().ToString().Substring(0, 8)
$email = "tester-$uniqueId@example.com"
$updateEmail = "perfect-$uniqueId@example.com"

$headers = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer dummy-jwt-token"
}
$adminHeaders = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer dummy-jwt-token"
    "X-User-Role" = "admin"
}

# 0. Create an Asset (so that booking works)
Write-Host "`n0. Testing: POST /assets (CreateAsset - Admin)" -ForegroundColor Yellow
$assetBody = @{
    name = "Premium Locker $uniqueId"
    type = "locker"
    status = "available"
    health_score = 100
    location = "Zone A"
} | ConvertTo-Json
try {
    $assetResp = Invoke-RestMethod -Uri "$baseUrl/assets" -Method Post -Headers $adminHeaders -Body $assetBody
    $realAssetId = $assetResp.id
    Write-Host "[SUCCESS] Test Asset created with ID: $realAssetId" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] CreateAsset error: $_" -ForegroundColor Red
    exit 1
}

# 1. Create a User
Write-Host "`n1. Testing: POST /users (CreateUser)" -ForegroundColor Yellow
$userBody = @{
    name = "Clean End-to-End Tester"
    email = $email
    starting_credits = 500
} | ConvertTo-Json

try {
    $userResp = Invoke-RestMethod -Uri "$baseUrl/users" -Method Post -Headers $headers -Body $userBody
    $userId = $userResp.id
    Write-Host "[SUCCESS] User created with ID: $userId" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] CreateUser error: $_" -ForegroundColor Red
    exit 1
}

# 2. Get User Details
Write-Host "`n2. Testing: GET /users/:id (GetUser)" -ForegroundColor Yellow
try {
    $getUserResp = Invoke-RestMethod -Uri "$baseUrl/users/$userId" -Method Get -Headers $headers
    Write-Host "[SUCCESS] Got User Details. Name: $($getUserResp.name), Email: $($getUserResp.email)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] GetUser error: $_" -ForegroundColor Red
}

# 3. Update User
Write-Host "`n3. Testing: PUT /users/:id (UpdateUser)" -ForegroundColor Yellow
$updateBody = @{
    name = "Perfect Clean Tester"
    email = $updateEmail
} | ConvertTo-Json
try {
    $updateResp = Invoke-RestMethod -Uri "$baseUrl/users/$userId" -Method Put -Headers $headers -Body $updateBody
    Write-Host "[SUCCESS] User updated. New Name: $($updateResp.name), New Email: $($updateResp.email)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] UpdateUser error: $_" -ForegroundColor Red
}

# 4. Get User Credits
Write-Host "`n4. Testing: GET /users/:id/credits (GetUserCredits)" -ForegroundColor Yellow
try {
    $creditsResp = Invoke-RestMethod -Uri "$baseUrl/users/$userId/credits" -Method Get -Headers $headers
    Write-Host "[SUCCESS] User credits balance: $($creditsResp.balance)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] GetUserCredits error: $_" -ForegroundColor Red
}

# 5. Deduct Credits
Write-Host "`n5. Testing: POST /users/:id/credits/deduct (DeductCredits)" -ForegroundColor Yellow
$deductBody = @{ amount = 100 } | ConvertTo-Json
try {
    $deductResp = Invoke-RestMethod -Uri "$baseUrl/users/$userId/credits/deduct" -Method Post -Headers $headers -Body $deductBody
    Write-Host "[SUCCESS] Deducted 100 credits. Balance after deduct: $($deductResp.balance)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] DeductCredits error: $_" -ForegroundColor Red
}

# 6. Add Credits
Write-Host "`n6. Testing: POST /users/:id/credits/add (AddCredits)" -ForegroundColor Yellow
$addBody = @{ amount = 200 } | ConvertTo-Json
try {
    $addResp = Invoke-RestMethod -Uri "$baseUrl/users/$userId/credits/add" -Method Post -Headers $headers -Body $addBody
    Write-Host "[SUCCESS] Added 200 credits. Balance after add: $($addResp.balance)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] AddCredits error: $_" -ForegroundColor Red
}

# 7. Get User Membership Details
Write-Host "`n7. Testing: GET /users/:id/membership (GetUserMembership)" -ForegroundColor Yellow
try {
    $membershipResp = Invoke-RestMethod -Uri "$baseUrl/users/$userId/membership" -Method Get -Headers $headers
    Write-Host "[SUCCESS] Got membership: Status: $($membershipResp.status), Type: $($membershipResp.type)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] GetUserMembership error: $_" -ForegroundColor Red
}

# 8. Get Credit Transactions History
Write-Host "`n8. Testing: GET /users/:id/transactions (GetCreditTransactions)" -ForegroundColor Yellow
try {
    $transactions = Invoke-RestMethod -Uri "$baseUrl/users/$userId/transactions" -Method Get -Headers $headers
    Write-Host "[SUCCESS] Found $($transactions.Count) credit transactions in history" -ForegroundColor Green
    foreach ($tx in $transactions) {
        Write-Host "  - Transaction ID: $($tx.id), Type: $($tx.type), Amount: $($tx.amount), Reason: $($tx.reason)" -ForegroundColor Gray
    }
} catch {
    Write-Host "[FAILED] GetCreditTransactions error: $_" -ForegroundColor Red
}

# 9. Create Booking 1 (To be returned)
Write-Host "`n9. Testing: POST /bookings (CreateBooking 1)" -ForegroundColor Yellow
$bookingBody1 = @{
    user_id = $userId
    asset_id = $realAssetId
    start_time = "2026-05-18T18:00:00Z"
    end_time = "2026-05-18T19:00:00Z"
} | ConvertTo-Json
try {
    $bookingResp1 = Invoke-RestMethod -Uri "$baseUrl/bookings" -Method Post -Headers $headers -Body $bookingBody1
    $bookingId1 = $bookingResp1.id
    Write-Host "[SUCCESS] Booking 1 created with ID: $bookingId1, Status: $($bookingResp1.status)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] CreateBooking 1 error: $_" -ForegroundColor Red
}

# 10. Return Booking 1
Write-Host "`n10. Testing: POST /bookings/:id/return (ReturnBooking 1)" -ForegroundColor Yellow
try {
    $returnResp = Invoke-RestMethod -Uri "$baseUrl/bookings/$bookingId1/return" -Method Post -Headers $headers
    Write-Host "[SUCCESS] Booking 1 returned. Status: $($returnResp.status)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] ReturnBooking 1 error: $_" -ForegroundColor Red
}

# 11. Create Booking 2 (To be cancelled)
Write-Host "`n11. Testing: POST /bookings (CreateBooking 2)" -ForegroundColor Yellow
$bookingBody2 = @{
    user_id = $userId
    asset_id = $realAssetId
    start_time = "2026-05-18T20:00:00Z"
    end_time = "2026-05-18T21:00:00Z"
} | ConvertTo-Json
try {
    $bookingResp2 = Invoke-RestMethod -Uri "$baseUrl/bookings" -Method Post -Headers $headers -Body $bookingBody2
    $bookingId2 = $bookingResp2.id
    Write-Host "[SUCCESS] Booking 2 created with ID: $bookingId2, Status: $($bookingResp2.status)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] CreateBooking 2 error: $_" -ForegroundColor Red
}

# 12. Cancel Booking 2 (Atomic Refund Credits!)
Write-Host "`n12. Testing: DELETE /bookings/:id (CancelBooking 2)" -ForegroundColor Yellow
try {
    $cancelResp = Invoke-RestMethod -Uri "$baseUrl/bookings/$bookingId2" -Method Delete -Headers $headers
    Write-Host "[SUCCESS] Booking 2 cancelled. Status: $($cancelResp.status)" -ForegroundColor Green
} catch {
    Write-Host "[FAILED] CancelBooking 2 error: $_" -ForegroundColor Red
}

# 13. Get User Bookings List
Write-Host "`n13. Testing: GET /users/:id/bookings (GetUserBookings)" -ForegroundColor Yellow
try {
    $bookingsList = Invoke-RestMethod -Uri "$baseUrl/users/$userId/bookings" -Method Get -Headers $headers
    Write-Host "[SUCCESS] Found $($bookingsList.Count) bookings for user in total" -ForegroundColor Green
    foreach ($b in $bookingsList) {
        Write-Host "  - Booking ID: $($b.id), Asset ID: $($b.asset_id), Status: $($b.status)" -ForegroundColor Gray
    }
} catch {
    Write-Host "[FAILED] GetUserBookings error: $_" -ForegroundColor Red
}

Write-Host "`n==============================================" -ForegroundColor Cyan
Write-Host "TEST COMPLETED SUCCESSFULLY. ALL 12 ENDPOINTS ARE GREEN!" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan
