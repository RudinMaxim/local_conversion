# Принимаем параметр для выбора команды
param (
    [Parameter(Mandatory=$true)]
    [ValidateSet("build", "deploy", "clean", "init")]
    [string]$Command
)

# Переменные
$AppName = "cli-app.exe" # Убедитесь, что на Linux/macOS этот файл не имеет расширения .exe
$BinaryPath = "bin/$AppName"
$SourceDir = "./input"
$TargetDir = "./output"

# Функция для сборки приложения
function Build {
    Write-Output "Сборка приложения..."
    go build -o $BinaryPath cmd/main.go
    if ($?) {
        Write-Output "Сборка завершена: $BinaryPath"
    } else {
        Write-Output "Ошибка сборки"
    }
}

# Функция для деплоя приложения в ветку main
function Deploy {
    git stash push -m "Temp changes before deploying"

    # Переключаемся на main, выполняем деплой, а затем возвращаемся в develop
    git checkout main
    if (!(Test-Path -Path bin)) { New-Item -ItemType Directory -Path bin }
    Move-Item -Path $BinaryPath -Destination "bin/" -Force
    git add bin/$AppName
    git commit -m "Deploy binary to main branch"
    git push origin main
    git checkout develop

    # Восстанавливаем изменения из стэша
    git stash pop
}

# Функция для очистки папок input, output и бинарника
function Clean {
    if (Test-Path $BinaryPath) { Remove-Item -Recurse -Force $BinaryPath }
    if (Test-Path "$SourceDir/*") { Remove-Item -Recurse -Force "$SourceDir/*" }
    if (Test-Path "$TargetDir/*") { Remove-Item -Recurse -Force "$TargetDir/*" }
}

# Функция для создания структуры проекта
function Init {
    Write-Output "Создание структуры проекта..."
    if (!(Test-Path -Path $SourceDir)) { New-Item -ItemType Directory -Path $SourceDir }
    if (!(Test-Path -Path $TargetDir)) { New-Item -ItemType Directory -Path $TargetDir }
    if (!(Test-Path -Path "bin")) { New-Item -ItemType Directory -Path "bin" }
    Write-Output "Структура проекта создана"
}

# Основная логика для выполнения команд
switch ($Command) {
    "build" { Build }
    "deploy" { Deploy }
    "clean" { Clean }
    "init" { Init }
    default { Write-Output "Неизвестная команда: $Command" }
}
