# CLI‑инструмент для анализа прямых зависимостей Go‑модулей и проверки наличия обновлений.

## Соберите бинарь:

```bash
go build -o go-dep-analyzer main.go
```

## Запустите на публичном репозитории:
```bash
./go-dep-analyzer https://github.com/user/example-go-module.git
```
## Пример вывода:
```bash
Модуль: github.com/user/example-go-module
Версия Go: 1.21
Зависимости:
- github.com/gin-gonic/gin: текущая v1.9.0 → доступна v1.10.1
- github.com/stretchr/testify: версия v1.7.0 (обновления нет)
```
