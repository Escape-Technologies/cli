$ErrorActionPreference = 'Stop'

$tempDir = Join-Path $env:TEMP "escape-cli-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
Set-Location $tempDir

try {
    $releaseUrl = "https://github.com/Escape-Technologies/cli/releases/latest"
    $response = Invoke-WebRequest -Uri $releaseUrl -MaximumRedirection 0 -ErrorAction SilentlyContinue -WarningAction SilentlyContinue
    $version = $response.Headers.Location -replace '.*v(.*?)$', '$1'
    
    Write-Host "Installing escape-cli v$version for Windows..."
    
    $installDir = "$env:ProgramFiles\Escape-Technologies\CLI"
    
    $zipName = "cli_${version}_windows_amd64.zip"
    $zipUrl = "https://github.com/Escape-Technologies/cli/releases/download/v${version}/${zipName}"
    
    Write-Host "Downloading from: $zipUrl"
    Invoke-WebRequest -Uri $zipUrl -OutFile "$tempDir\$zipName"
    
    Write-Host "Extracting..."
    Expand-Archive -Path "$tempDir\$zipName" -DestinationPath $tempDir -Force
    
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    
    Copy-Item -Path "$tempDir\escape-cli.exe" -Destination "$installDir\escape-cli.exe" -Force
    
    # Add to PATH if not already there
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($currentPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "Machine")
        Write-Host "Added Escape CLI to the system PATH"
    }
    
    Write-Host @"
Done! You can now use the escape-cli command.

Please restart your terminal for PATH changes to take effect.

Then, go to https://app.escape.tech/user/profile/ copy your API key and set it as an environment variable:

[Environment]::SetEnvironmentVariable("ESCAPE_API_KEY", "your-api-key", "User")

Or run this in your terminal:

`$env:ESCAPE_API_KEY = "your-api-key"

Then, you should be able to run:

escape-cli applications list
"@
}
finally {
    # Clean up
    Set-Location $env:USERPROFILE
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
} 