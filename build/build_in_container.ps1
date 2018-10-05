param (
  [string]$platform,
  [string]$arch
)

$ErrorActionPreference = "Stop";

$binary = "portainer.exe"

Set-Item env:GOPATH "$PWD\api"

New-Item -Name dist -Path "$PWD" -ItemType Directory | Out-Null
New-Item -Name portainer -Path "$PWD\api\src\github.com\" -ItemType Directory | Out-Null

Copy-Item -Path "$PWD\api" -Destination "$PWD\api\src\github.com\portainer" -Recurse -Force -ErrorAction:SilentlyContinue
Rename-Item -Path "$PWD\api\src\github.com\portainer\api" -NewName "portainer" -ErrorAction:SilentlyContinue

Set-Location -Path "$PWD\api\cmd\portainer"

C:\go\bin\go.exe get -t -d -v ./...
C:\go\bin\go.exe build -v

Move-Item -Path "$PWD\api\cmd\portainer\$($binary)" -Destination "$PWD\dist"