# build.ps1

# Accepts a parameter to select a command
param (
    [Parameter(Mandatory=$true)]
    [ValidateSet("build", "deploy", "clean", "init")]
    [string]$Command
)

# Variables
$AppName = "local_conversion.exe" # Adjust as needed for Linux/macOS
$BinaryPath = "bin/$AppName"
$SourceDir = "./input"
$TargetDir = "./output"

# Build function
function Build {
    Write-Output "Building application..."
    go build -o $BinaryPath main.go
    if (Test-Path $BinaryPath) {
        Write-Output "Build completed: $BinaryPath"
    } else {
        Write-Output "Build error: file $BinaryPath was not created."
        exit 1
    }
}

# Deploy function to the main branch
# Deploy function to the main branch
function Deploy {
    Write-Output "Deploying the application to main..."

    # Stash any unsaved changes in develop
    git checkout develop
    git stash push -m "Temp changes before deploying"

    # Switch to main branch
    git checkout main

    # Ensure the necessary directories exist in main
    if (!(Test-Path -Path "bin")) { New-Item -ItemType Directory -Path "bin" }
    if (!(Test-Path -Path "input")) { New-Item -ItemType Directory -Path "input" }
    if (!(Test-Path -Path "output")) { New-Item -ItemType Directory -Path "output" }

    # Build the application in main
    Build

    # Clean up unnecessary files in main (excluding bin, input, output, .gitignore, and README.md)
    Write-Output "Cleaning up unnecessary files in main branch..."
    Get-ChildItem -Recurse | Where-Object {
        $_.FullName -notmatch "bin" -and
        $_.FullName -notmatch "input" -and
        $_.FullName -notmatch "output" -and
        $_.Name -notmatch "\.gitignore" -and
        $_.Name -notmatch "README.md"
    } | Remove-Item -Recurse -Force

    # Add the binary and necessary directories to main
    Copy-Item -Path $BinaryPath -Destination "bin/" -Force
    git add bin/$AppName input output .gitignore README.md
    git commit -m "Deploy binary and essential files to main branch"
    git push origin main

    # Switch back to develop branch and apply stashed changes
    git checkout develop
    git stash pop
    Write-Output "Deployment complete. Develop branch unchanged."
}


# Clean function for input, output, and binary file
function Clean {
    Write-Output "Cleaning folders and binary file..."
    if (Test-Path $BinaryPath) { Remove-Item -Recurse -Force $BinaryPath }
    if (Test-Path "$SourceDir/*") { Remove-Item -Recurse -Force "$SourceDir/*" }
    if (Test-Path "$TargetDir/*") { Remove-Item -Recurse -Force "$TargetDir/*" }
    Write-Output "Clean completed."
}

# Initialize the project structure
function Init {
    Write-Output "Initializing project structure..."
    if (!(Test-Path -Path $SourceDir)) { New-Item -ItemType Directory -Path $SourceDir }
    if (!(Test-Path -Path $TargetDir)) { New-Item -ItemType Directory -Path $TargetDir }
    if (!(Test-Path -Path "bin")) { New-Item -ItemType Directory -Path "bin" }
    Write-Output "Project structure initialized."
}

# Main command logic
switch ($Command) {
    "build" { Build }
    "deploy" { Deploy }
    "clean" { Clean }
    "init" { Init }
    default { Write-Output "Unknown command: $Command" }
}
