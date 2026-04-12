# Скрипт для копирования фруктовых спрайтов в проект match3

$sourceBase = "D:\Projects\sprites\03_Food\food-ocal\32x32\fruit"
$destBase = "D:\Projects\playgo\match3\assets\fruits"

# Создаем папку назначения
if (Test-Path $destBase) {
    Remove-Item $destBase -Recurse -Force
}
New-Item -ItemType Directory -Path $destBase -Force | Out-Null

# Выбираем 6 фруктов для игры
$fruits = @(
    "apple.png",
    "banana.png",
    "strawberry.png",
    "orange.png",
    "kiwi.png",
    "grapes.png"
)

# Копируем каждый фрукт
for ($i = 0; $i -lt $fruits.Count; $i++) {
    $fruit = $fruits[$i]
    $source = Join-Path $sourceBase $fruit
    $dest = Join-Path $destBase "$i.png"
    
    if (Test-Path $source) {
        Copy-Item $source $dest
        Write-Host "Copied: $fruit -> $i.png"
    } else {
        Write-Host "Warning: $fruit not found"
    }
}

Write-Host "`nFruit sprites copied successfully!"
