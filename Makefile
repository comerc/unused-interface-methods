.PHONY: build test clean help cover

# Переменные
BINARY_NAME=unused-interface-methods
COVERAGE_DIR=.coverage
COVERAGE_OUT=$(COVERAGE_DIR)/.out
COVERAGE_TMP=$(COVERAGE_DIR)/.tmp
COVERAGE_HTML=$(COVERAGE_DIR)/.html
COVERAGE_TXT=$(COVERAGE_DIR)/.txt

# Сборка основной утилиты
build:
	@echo "🔨 Сборка линтера..."
	@go build -o $(BINARY_NAME) .

# Тестирование
test:
	@echo "🧪 Запуск тестов..."
	@go test -v ./...

# Проверка неиспользуемых методов интерфейсов
check: build
	@echo "🔍 Проверка неиспользуемых методов интерфейсов..."
	@./$(BINARY_NAME) test/data/

# Установка зависимостей
deps:
	@echo "📦 Установка зависимостей..."
	@go mod tidy
	@go mod download

# Очистка
clean:
	@echo "🧹 Очистка..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(COVERAGE_DIR)

# Отчет о покрытии кода
cover:
	@echo "📊 Генерация отчета о покрытии кода..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_TMP) ./...
	@cat $(COVERAGE_TMP) | grep -v "/test/data/" > $(COVERAGE_OUT)
	@rm $(COVERAGE_TMP)
	@go tool cover -func=$(COVERAGE_OUT) | tee $(COVERAGE_TXT)
	@go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "✨ Отчеты о покрытии сгенерированы в директории $(COVERAGE_DIR):"
	@echo "   - .txt  - текстовый отчет"
	@echo "   - .html - HTML отчет"
	@echo "   - .out  - исходные данные"

# Справка
help:
	@echo "📋 Доступные команды:"
	@echo "  build          - Собрать основную утилиту"
	@echo "  test           - Запустить тесты"
	@echo "  check          - Быстрая проверка тестовых данных"
	@echo "  deps           - Установить зависимости"
	@echo "  clean          - Очистить собранные файлы"
	@echo "  cover          - Сгенерировать отчет о покрытии кода"
	@echo "  help           - Показать эту справку"

.PHONY: golangci-lint
golangci-lint:
	@echo "🔍 Запуск golangci-lint..."
	@golangci-lint run 