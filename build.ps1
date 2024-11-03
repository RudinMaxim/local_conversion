# Принимаем параметр для выбора команды
param (
    [Parameter(Mandatory=$true)]
    [ValidateSet("build", "deploy", "clean", "init")]
    [string]$Command
)

# Переменные
$AppName = "local_conversion.exe"
$BinaryPath = "bin/$AppName"
$SourceDir = "./input"
$TargetDir = "./output"

# Функция для сборки приложения
function Build {
    Write-Output "Assembling the application..."
    go build -o $BinaryPath cmd/main.go
    if (Test-Path $BinaryPath) {
        Write-Output "Assembly complete: $BinaryPath"
    } else {
        Write-Output "Build error: file $BinaryPath was not created."
        exit 1
    }
}

# Функция для деплоя приложения в ветку main
function Deploy {
    Write-Output "Deploy the application to main..."

    # Сохраняем изменения в стэш
    git stash push -m "Temp changes before deploying"

    # Переключаемся на main и выполняем деплой, затем возвращаемся в develop
    git checkout main
    Build  # Пересборка перед деплоем
    if (Test-Path $BinaryPath) {
        if (!(Test-Path -Path bin)) { New-Item -ItemType Directory -Path bin }
        Move-Item -Path $BinaryPath -Destination "bin/" -Force
        git add bin/$AppName
        git commit -m "Deploy binary to main branch"
        git push origin main
    } else {
        Write-Output "Deployment failed: Binary file missing."
    }
    git checkout develop

    # Восстанавливаем изменения из стэша
    git stash pop
    Write-Output "Deployment completed and changes restored to develop"
}

# Функция для очистки папок input, output и бинарника
function Clean {
    Write-Output "Cleaning folders and binary file..."
    if (Test-Path $BinaryPath) { Remove-Item -Recurse -Force $BinaryPath }
    if (Test-Path "$SourceDir/*") { Remove-Item -Recurse -Force "$SourceDir/*" }
    if (Test-Path "$TargetDir/*") { Remove-Item -Recurse -Force "$TargetDir/*" }
    Write-Output "Cleaning completed"
}

# Функция для создания структуры проекта
function Init {
    Write-Output "Creating a project structure..."
    if (!(Test-Path -Path $SourceDir)) { New-Item -ItemType Directory -Path $SourceDir }
    if (!(Test-Path -Path $TargetDir)) { New-Item -ItemType Directory -Path $TargetDir }
    if (!(Test-Path -Path "bin")) { New-Item -ItemType Directory -Path "bin" }
    Write-Output "The project structure has been created"
}

# Основная логика для выполнения команд
switch ($Command) {
    "build" { Build }
    "deploy" { Deploy }
    "clean" { Clean }
    "init" { Init }
    default { Write-Output "Неизвестная команда: $Command" }
}
