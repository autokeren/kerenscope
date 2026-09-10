$ErrorActionPreference = "Stop"

$Repo = "autokeren/kerenscope"
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
  "AMD64" { "amd64" }
  "ARM64" { "arm64" }
  default { "amd64" }
}
$Base = "https://github.com/$Repo/releases/latest/download"
$Dir = "$env:LOCALAPPDATA\Programs\kerenscope"

Write-Host "-> Downloading kerenscope for windows-$Arch..."
New-Item -ItemType Directory -Force -Path $Dir | Out-Null
Invoke-WebRequest -Uri "$Base/kerenscope-windows-$Arch.zip" -OutFile "$env:TEMP\kerenscope.zip"

Write-Host "-> Installing to $Dir..."
Expand-Archive -Path "$env:TEMP\kerenscope.zip" -DestinationPath $Dir -Force
Remove-Item "$Dir\keren.exe" -ErrorAction SilentlyContinue
Rename-Item "$Dir\kerenscope-windows-$Arch.exe" "$Dir\keren.exe"

Remove-Item "$env:TEMP\kerenscope.zip" -ErrorAction SilentlyContinue

if ($env:Path -notlike "*$Dir*") {
  [Environment]::SetEnvironmentVariable("Path", "$([Environment]::GetEnvironmentVariable('Path','User'));$Dir", "User")
  Write-Host "   Added $Dir to your PATH (restart your terminal to pick it up)."
}

Write-Host ""
& "$Dir\keren.exe" --version
Write-Host "OK  Installed: keren - try: keren company BBCA"
