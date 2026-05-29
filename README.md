# Crypto Bot Parser

[English](#english) | [Русский](#русский)

---

## English

**Crypto Bot Parser** is a Go-based utility designed to automatically monitor (intercept) P2C transactions within a CryptoBot bot, with the ability to confirm or decline operations.

### Features
* **Automated Interception**: Real-time monitoring of incoming P2C transactions.
* **Telegram Integration**: Prompt command delivery for confirming or canceling operations.
* **Built with Go**: High performance, reliability, and efficient concurrency management.

### Project Structure
* `cmd/parser/` — Application entry point (main script).
* `internal/` — Internal parser logic hidden from external imports.
* `pkg/` — Auxiliary reusable packages and libraries.

### Requirements
* Go (version 1.18 or higher)

### Installation & Usage

1. Clone the repository:
   ```bash
   git clone https://github.com
   cd crypto_bot_parser
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

3. Run the project:
   ```bash
   go run cmd/parser/main.go
   ```

### Disclaimer

This software has been developed **solely for educational, informational, and demonstration purposes**. 

The developer **is completely absolved of any and all liability** for any direct or indirect consequences resulting from the use, modification, or execution of this software. 

Cryptocurrency operations and automated platform interactions may be restricted or prohibited by the laws of your country, or they may violate the Terms of Service of third-party platforms. The end user assumes full and sole responsibility for compliance with all local laws, tax regulations, and third-party rules when using this tool.

### License

This code is distributed under a custom non-commercial license:
1. You are free to use, study, and modify this code.
2. **Mandatory Condition**: Any further distribution (copying, publishing, forking) of this code or its modified versions must be done **exclusively on a free-of-charge basis**. Selling this code or charging a fee for access to it is strictly prohibited.
3. **Commercial Services**: Custom development, tailored modifications, and server deployment are commercial services provided exclusively by the author and must be negotiated separately. Contact: [Telegram @hopstee](https://t.me.hopstee).

> **Use the code at your own risk. Good luck :)**

---

## Русский

**Crypto Bot Parser** — это утилита на языке Go, предназначенная для автоматического отслеживания (перехвата) P2C транзакций в боте CryptoBot с возможностью последующего подтверждения или отклонения операций.

### Особенности
* **Автоматический перехват**: Мониторинг входящих P2C-транзакций в реальном времени.
* **Интеграция с Telegram**: Быстрая отправка команд на подтверждение или отмену операций.
* **Написано на Go**: Высокая производительность, надежность и эффективная работа с многопоточностью.

### Структура проекта
* `cmd/parser/` — Точка входа в приложение (основной скрипт запуска).
* `internal/` — Внутренняя логика парсера, скрытая от внешнего импорта.
* `pkg/` — Вспомогательные переиспользуемые пакеты и библиотеки.

### Требования
* Установленный Go (версии 1.18 или выше)

### Установка и запуск

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com
   cd crypto_bot_parser
   ```

2. Установите зависимости:
   ```bash
   go mod download
   ```

3. Запустите проект:
   ```bash
   go run cmd/parser/main.go
   ```

### Отказ от ответственности

Данный программный код был разработан **исключительно в ознакомительных, учебных и демонстрационных целях**. 

Разработчик **полностью освобождается от любой ответственности** за любые прямые или косвенные последствия, возникшие в результате использования, модификации или запуска данного ПО. 

Криптовалютные операции и автоматизация взаимодействия с сервисами могут быть ограничены или запрещены законодательством вашей страны, а также могут нарушать правила использования (Terms of Service) сторонних платформ. Вся ответственность за соблюдение местного законодательства, налоговых норм и правил третьих лиц при использовании данного инструмента целиком и полностью ложится на конечного пользователя.

### Лицензия (License)

Этот код распространяется на условиях кастомной некоммерческой лицензии:
1. Вы имеете право свободно использовать, изучать и модифицировать данный код.
2. **Обязательное условие**: Любое дальнейшее распространение (копирование, публикация, форкинг) данного кода или его модифицированных версий должно осуществляться **исключительно на бесплатной основе**. Продажа данного кода или взимание платы за доступ к нему строго запрещены.
3. **Коммерческие услуги**: Кастомная доработка кода под ваши нужды, индивидуальные модификации и развертывание (деплой) на сервере являются коммерческими услугами автора и обсуждаются отдельно. Для связи: [Telegram @hopstee](https://t.me.hopstee).

> **Используй код на свой страх и риск. Удачи :)**
