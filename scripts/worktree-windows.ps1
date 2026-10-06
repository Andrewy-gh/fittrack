param([Parameter(Mandatory=$true)][string]$OutputPath)
$ErrorActionPreference = 'Stop'
# Compile once per source hash. Long-running commands do not retain PowerShell.
Add-Type -Path (Join-Path $PSScriptRoot 'worktree-windows.cs') -OutputAssembly $OutputPath -OutputType ConsoleApplication
