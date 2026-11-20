# Bootstrap a fresh local Supabase PostgreSQL 17.12 cluster for this linked checkout.
# Run in PowerShell 7 from this checkout. Existing clusters and env files are never adopted.
param ([string]$Runtime = "$HOME\go\bin\worktree-runtime.exe")
$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true
$root = (git rev-parse --show-toplevel).Trim()
Push-Location -LiteralPath $root
try {
    $image = 'supabase/postgres@sha256:8630950b489ecc70563300180287a98c192a4a48f5ae00b4febdaeea1abe85de'
    $goose = Join-Path $root '.worktree\bin\goose.exe'
    if (!(Test-Path -LiteralPath $Runtime) -or !(Test-Path -LiteralPath $goose) -or
        (& $goose -version) -ne 'goose version: v3.24.3') {
        throw 'Install worktree-runtime and local Goose v3.24.3 before bootstrap.'
    }
    foreach ($path in @('server/.env', 'client/.env', 'server/.env.local', 'client/.env.local', '.worktree/postgres-target', '.worktree/pg17-profile.json')) {
        if (Test-Path -LiteralPath $path) { throw "Refusing to overwrite $path" }
    }
    # Ignore Docker contexts and remote daemon settings for this local-only workflow.
    $docker = @('--host', 'npipe:////./pipe/docker_engine')
    $dockerVars = @{}
    foreach ($entry in @(Get-ChildItem Env:DOCKER_*)) {
        $dockerVars[$entry.Name] = $entry.Value
        Remove-Item -LiteralPath "Env:$($entry.Name)"
    }
    & $Runtime init
    $allocation = (& $Runtime status --json | ConvertFrom-Json)
    if ($allocation.id -notmatch '^[a-z0-9_]+_([a-f0-9]{12})$') { throw 'Unexpected worktree identity' }
    $container = 'fittrack-local-pg17-' + $Matches[1]
    $volume = "${container}-data"
    if ((& docker @docker ps -a --format '{{.Names}}') -contains $container -or
        (& docker @docker volume ls --format '{{.Name}}') -contains $volume) {
        throw 'Refusing to adopt an existing container or volume.'
    }
    if (Get-NetTCPConnection -State Listen -LocalPort 26432 -ErrorAction SilentlyContinue) {
        throw 'The dedicated loopback port 26432 is occupied.'
    }
    # Check image exists before allocating state. Operator pulls the documented digest if needed.
    & docker @docker image inspect $image --format '{{.Id}}' | Out-Null
    & docker @docker volume create --label "fittrack.worktree=$($allocation.id)" $volume | Out-Null
    $admin = [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(24))
    $owner = [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(24))
    $app = [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(24))
    $priorPassword = $env:POSTGRES_PASSWORD
    try {
        $env:POSTGRES_PASSWORD = $admin
        & docker @docker run -d --pull=never --name $container --label fittrack.local-postgres=17.12 `
            --label "fittrack.worktree=$($allocation.id)" -p 127.0.0.1:26432:5432 `
            --mount "type=volume,source=$volume,target=/var/lib/postgresql/data" -e POSTGRES_PASSWORD $image | Out-Null
    } finally {
        $env:POSTGRES_PASSWORD = $priorPassword
    }
    $ready = $false
    for ($i = 0; $i -lt 90; $i++) {
        $PSNativeCommandUseErrorActionPreference = $false
        & docker @docker exec $container pg_isready -h 127.0.0.1 -U supabase_admin -d postgres *> $null
        $PSNativeCommandUseErrorActionPreference = $true
        if ($LASTEXITCODE -eq 0) { $ready = $true; break }
        Start-Sleep -Seconds 1
    }
    if (!$ready) { throw 'PostgreSQL did not become ready; container and data retained for inspection.' }
    if ((& docker @docker exec $container psql -U supabase_admin -d postgres -X -Atqc 'SHOW server_version_num') -ne '170012') {
        throw 'The dedicated container did not report PostgreSQL 17.12.'
    }
    "CREATE ROLE fittrack_local_owner LOGIN CREATEDB CREATEROLE PASSWORD '$owner';" |
        & docker @docker exec -i $container psql -U supabase_admin -d postgres -X -v ON_ERROR_STOP=1 | Out-Null
    @("FITTRACK_LOCAL_DATABASE_URL=postgresql://fittrack_local_owner:${owner}@127.0.0.1:26432/postgres?sslmode=disable",
      "FITTRACK_LOCAL_RUNTIME_DATABASE_URL=postgresql://fittrack_app:${app}@127.0.0.1:26432/postgres?sslmode=disable") |
        Set-Content -LiteralPath server/.env.local
    @{version=1;environment=@{FITTRACK_LOCAL_POSTGRES_CONTAINER=$container;FITTRACK_GOOSE=$goose;E2E_FRONTEND_MODE='development'}} |
        ConvertTo-Json | Set-Content -LiteralPath .worktree/pg17-profile.json
    Set-Content -LiteralPath .worktree/postgres-target -Value supabase-17.12
    Write-Output "Created $container and its dedicated volume. Continue with the documented prepare step."
} finally {
    if ($dockerVars) {
        foreach ($name in $dockerVars.Keys) { Set-Item -LiteralPath "Env:$name" -Value $dockerVars[$name] }
    }
    Pop-Location
}
