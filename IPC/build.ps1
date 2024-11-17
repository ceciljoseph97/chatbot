# build.ps1
# Build script for PeriChat project
# This script builds both Chatbot and WebDeploy executables for each target platform
# and places them, along with their necessary configuration files, into a single bin/{os}/{arch}/ directory.

param (
    [string]$OutputDir = "bin"
)

# Define build targets
$targets = @(
    @{OS='windows'; Arch='amd64'; Extension='.exe'; GoARM=$null},
    @{OS='linux'; Arch='amd64'; Extension=''; GoARM=$null},
    @{OS='linux'; Arch='arm'; Extension=''; GoARM='7'}
)

# Function to build a Go executable
function Build-GoExecutable {
    param (
        [string]$Dir,
        [string]$OutputPath,
        [string]$SourceFile
    )
    Push-Location $Dir
    # Build the executable
    Write-Host "Building $SourceFile in $Dir..."
    go build -o $OutputPath $SourceFile
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed for $SourceFile in $Dir"
        Pop-Location
        exit 1
    }
    if (Test-Path $OutputPath) {
        Write-Host "Successfully built $SourceFile -> $OutputPath"
    } else {
        Write-Error "Build reported success but $OutputPath does not exist."
        Pop-Location
        exit 1
    }
    Pop-Location
}

# Function to copy directory recursively
function Copy-Directory {
    param (
        [string]$SourceDir,
        [string]$DestinationDir
    )
    if (Test-Path $SourceDir) {
        # Remove the destination directory if it exists to ensure a clean copy
        if (Test-Path $DestinationDir) {
            Remove-Item -Path $DestinationDir -Recurse -Force
            Write-Host "Removed existing directory $DestinationDir"
        }
        Copy-Item -Path $SourceDir -Destination $DestinationDir -Recurse -Force
        Write-Host "Copied $SourceDir to $DestinationDir"
    } else {
        Write-Warning "Source directory $SourceDir does not exist. Skipping."
    }
}

# Function to clean service-specific bin directories
function Clean-ServiceBins {
    param (
        [string]$ServiceDir
    )
    $binPath = Join-Path $ServiceDir "bin"
    if (Test-Path $binPath) {
        Write-Host "Removing existing bin directory in $ServiceDir..."
        Remove-Item -Path $binPath -Recurse -Force
        Write-Host "Removed $binPath"
    } else {
        Write-Host "No bin directory found in $ServiceDir. Skipping removal."
    }
}

# Function to verify binaries in central bin directory
function Verify-Binaries {
    param (
        [string]$BinDir
    )
    Write-Host "Verifying binaries in $BinDir..."
    Get-ChildItem -Path $BinDir -File | ForEach-Object {
        Write-Host " - $($_.Name)"
    }
}

# Set OutputDir to absolute path
$OutputDir = Join-Path $PSScriptRoot $OutputDir

# Ensure output directory exists
if (-Not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
    Write-Host "Created output directory $OutputDir"
} else {
    Write-Host "Output directory $OutputDir already exists."
}

# Define service directories
$chatbotDir = Join-Path $PSScriptRoot "Chatbot"
$webDir = Join-Path $PSScriptRoot "web"

# Clean service-specific bin directories to prevent duplication
Clean-ServiceBins -ServiceDir $chatbotDir
Clean-ServiceBins -ServiceDir $webDir

# Iterate over each target
foreach ($target in $targets) {
    $os = $target.OS
    $arch = $target.Arch
    $ext = $target.Extension
    $goarm = $target.GoARM

    Write-Host "------------------------------"
    Write-Host "Building for OS: $os, Arch: $arch, GOARM: $goarm"
    Write-Host "------------------------------"

    # Set environment variables for cross-compilation
    $env:GOOS = $os
    $env:GOARCH = $arch
    if ($goarm) {
        $env:GOARM = $goarm
    } else {
        Remove-Item Env:\GOARM -ErrorAction SilentlyContinue
    }
    $env:CGO_ENABLED = "0"

    # Define output subdirectory for this target
    $outputSubDir = Join-Path $OutputDir "$os\$arch"
    if (-Not (Test-Path $outputSubDir)) {
        New-Item -ItemType Directory -Path $outputSubDir -Force | Out-Null
        Write-Host "Created output subdirectory $outputSubDir"
    } else {
        Write-Host "Output subdirectory $outputSubDir already exists."
    }

    # === Build Chatbot Executable ===
    $chatbotSourceFile = "chatbot.go"
    $chatbotBuildName = "chatbot$ext"
    $chatbotOutputPath = Join-Path $outputSubDir $chatbotBuildName

    try {
        Build-GoExecutable -Dir $chatbotDir -OutputPath $chatbotOutputPath -SourceFile $chatbotSourceFile
    }
    catch {
        Write-Error "Failed to build Chatbot for OS: $os, Arch: $arch"
        exit 1
    }

    # === Build WebDeploy Executable ===
    $webSourceFile = "webDeploy.go"
    $webBuildName = "webDeploy$ext"
    $webOutputPath = Join-Path $outputSubDir $webBuildName

    try {
        Build-GoExecutable -Dir $webDir -OutputPath $webOutputPath -SourceFile $webSourceFile
    }
    catch {
        Write-Error "Failed to build WebDeploy for OS: $os, Arch: $arch"
        exit 1
    }

    # === Copy Configuration and Associated Files ===

    # List of files to copy for both Chatbot and WebDeploy
    $commonFiles = @(
        # Chatbot configuration files
        @{Source = Join-Path $chatbotDir "config_local_gen.yaml"; Destination = Join-Path $outputSubDir "config_local_gen.yaml"},
        @{Source = Join-Path $chatbotDir "PMFuncOverView.gob"; Destination = Join-Path $outputSubDir "PMFuncOverView.gob"}
    )

    foreach ($file in $commonFiles) {
        $src = $file.Source
        $dest = $file.Destination
        if (Test-Path $src) {
            Copy-Item -Path $src -Destination $dest -Force
            Write-Host "Copied $src to $dest"
        } else {
            Write-Warning "Configuration file $src does not exist. Skipping."
        }
    }

    # Copy 'etc' directory from Chatbot
    $chatbotEtcSrc = Join-Path $chatbotDir "etc"
    $chatbotEtcDest = Join-Path $outputSubDir "etc"
    Copy-Directory -SourceDir $chatbotEtcSrc -DestinationDir $chatbotEtcDest

    # Copy 'static' directory from WebDeploy
    $webStaticSrc = Join-Path $webDir "static"
    $webStaticDest = Join-Path $outputSubDir "static"
    Copy-Directory -SourceDir $webStaticSrc -DestinationDir $webStaticDest

    # === Additional WebDeploy Configuration Files ===
    # If WebDeploy has its own configuration files, add them here.
    # For example, if there's a 'web_config.yaml', uncomment and modify the following lines:
    #
    # $webConfigFiles = @("web_config.yaml")
    # foreach ($file in $webConfigFiles) {
    #     $src = Join-Path $webDir $file
    #     $dest = Join-Path $outputSubDir $file
    #     if (Test-Path $src) {
    #         Copy-Item -Path $src -Destination $dest -Force
    #         Write-Host "Copied $src to $dest"
    #     } else {
    #         Write-Warning "WebDeploy config file $src does not exist. Skipping."
    #     }
    # }

    # === Verify Binaries ===
    Verify-Binaries -BinDir $outputSubDir

    Write-Host "Build for OS: $os, Arch: $arch completed successfully."
}

Write-Host "All builds completed successfully."
