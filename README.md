# DOCX to TXT Converter

[![Go Report Card](https://goreportcard.com/badge/github.com/terratensor/docx2txt)](https://goreportcard.com/report/github.com/terratensor/docx2txt)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**DOCX to TXT Converter** — это утилита на Go для конвертации файлов DOCX в TXT или GZ. Программа поддерживает многопоточную обработку, упаковку в GZ с настраиваемым уровнем сжатия, логирование ошибок и сохранение проблемных файлов в отдельную папку.

---

## Установка

1. Убедитесь, что у вас установлен Go (версия 1.16 или выше).
2. Склонируйте репозиторий:

   ```bash
   git clone https://github.com/terratensor/docx2txt.git
   cd docx2txt
   ```
3. Установите зависимости:

   ```bash
   go mod download
   ```

## Использование

### Основные параметры
- `-input`: Путь к папке с DOCX файлами (обязательный).

- `-output`: Путь к папке для сохранения TXT/GZ файлов (обязательный).

- `-goroutines`: Максимальное количество одновременно работающих горутин (по умолчанию равно количеству ядер CPU).

- `-compress`: Упаковывать ли выходные файлы в GZ (по умолчанию `false`).

- `-level`: Уровень сжатия GZIP (от 0 до 9, по умолчанию 6).

- `-errors`: Путь к папке для сохранения файлов с ошибками (по умолчанию `errors`).

## Примеры запуска
1. **Конвертация в TXT:**

   ```bash
   go run main.go -input ./examples/input -output ./examples/output
   ```
2. **Конвертация в GZ с максимальным сжатием:**

   ```bash
   go run main.go -input ./examples/input -output ./examples/output -compress -level 9
   ```
3. **Указание папки для ошибок:**

   ```bash
   go run main.go -input ./examples/input -output ./examples/output -errors ./examples/errors
   ```
4. **Ограничение количества горутин:**

   ```bash
   go run main.go -input ./examples/input -output ./examples/output -goroutines 4
   ```

## Логирование

Все ошибки записываются в файл `conversion_errors.log` в корне проекта. Пример лога:

```bash
ERROR: 2025/02/15 20:51:10 main.go:91: Ошибка при чтении из файла: XML syntax error on line 2: illegal character code U+000B
ERROR: 2025/02/15 20:51:10 main.go:98: Файл example.docx успешно скопирован в папку с ошибками: errors/example.docx
```

## Пример структуры папок
### Входная папка (input)
```bash
examples/input/
├── example1.docx
├── example2.docx
└── example3.docx
```

### Выходная папка (output)
```bash
examples/output/
├── example1.txt
├── example2.txt.gz
└── example3.txt
```

### Папка с ошибками (errors)
```bash
examples/errors/
└── example_error.docx
```

## Лицензия
Этот проект распространяется под лицензией MIT. Подробнее см. в файле LICENSE.

## Как использовать проект

1. Склонируйте репозиторий:

   ```bash
   git clone https://github.com/terratensor/docx2txt.git
   cd docx2txt
   ```

2. Установите зависимости:

   ```bash
   go mod download
   ```
   
3. Запустите программу:

   ```bash
   go run main.go -input ./examples/input -output ./examples/output
   ```