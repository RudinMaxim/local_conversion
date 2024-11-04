# local_conversion

powershell -File build.ps1 -Command init — создаёт структуру проекта.
powershell -File build.ps1 -Command build — собирает бинарный файл.
powershell -File build.ps1 -Command deploy — деплоит бинарный файл в ветку main.
powershell -File build.ps1 -Command clean — очищает бинарник и содержимое папок input и output.
