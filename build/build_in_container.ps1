param (
  [string]$platform,
  [string]$arch
)

$binary = "portainer-$($platform)-$($arch)"

New-Item -Name dist -ItemType Directory | Out-Null

docker run -e CGO_ENABLED=0 -v "$PWD\api:C:\gopath" golang:1.10-windowsservercore-1803 go get -t -d -v ./...; go build -v

Move-Item -Path "$PWD\api\cmd\portainer\$($binary)" -Destination dist/
