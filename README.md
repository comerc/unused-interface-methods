# Unused Interface Methods

Линтер для поиска неиспользуемых методов в интерфейсах Go.

## Возможности

- ✅ Находит неиспользуемые методы в интерфейсах
- ✅ Анализирует код рекурсивно по всем вложенным модулям
- ✅ Поддерживает различные паттерны использования методов
- ✅ Настраиваемые правила игнорирования файлов и директорий

## Установка

```bash
go install github.com/comerc/unused-interface-methods@latest
```

## Использование

```bash
# Анализ текущей директории
./unused-interface-methods

# Анализ конкретной директории
./unused-interface-methods ./path

# С подробным выводом
./unused-interface-methods -v ./path

# Справка
./unused-interface-methods -h
```

## Конфигурация

```yaml
# unused-interface-methods.yml
ignore:
  - "**/*_test.go"
  - "test/**"
  - "**/mock/**"
```

Файл ищется автоматически в текущей директории (или `.config/`) с опциональной точкой в префиксе файла.

## Пример вывода

```
UNUSED: EventHandler.OnError(handler unknown) error (test/data/interfaces.go:38)
UNUSED: EventHandler.Subscribe(filter unknown, cb unknown) error (test/data/interfaces.go:39)
UNUSED: ChannelProcessor.ReceiveData(ch chan string) error (test/data/interfaces.go:45)
UNUSED: DataProcessor.ProcessSlice(data []string) error (test/data/interfaces.go:52)

⚠️  7 generic interfaces skipped (26 methods not analyzed)
```

## Ограничения

Генерик-интерфейсы обнаруживаются, но не анализируются из-за сложностей системы типов Go. Детальное техническое объяснение см. в [GENERICS_PROBLEM.md](./doc/GENERICS_PROBLEM.md).

## Лицензия

MIT

---

А, теперь понял! Нам нужно:
Скопировать файлы
Найти все интерфейсы
Для каждого интерфейса удалять по одному методу и проверять через staticcheck
Если staticcheck не выдает ошибок - метод не используется