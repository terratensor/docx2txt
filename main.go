package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/terratensor/docx2txt/internal/docc"
)

// Функция для извлечения текста из DOCX и сохранения в TXT или GZ
func convertDocxToTxt(inputPath, outputPath string, wg *sync.WaitGroup, counter *int, totalFiles int, mutex *sync.Mutex, compress bool, compressionLevel int, errorDir string, logger *log.Logger) {
	defer wg.Done()

	// Улучшенное регулярное выражение для поиска base64-кодированных данных
	reBase64 := regexp.MustCompile(`(?:[A-Za-z0-9+/]{40,}={0,2}|iVBORw0KGgo[^"]+)`)

	// Открываем DOCX файл
	r, err := docc.NewReader(inputPath, reBase64)
	if err != nil {
		logErrorAndMoveFile(inputPath, errorDir, fmt.Sprintf("Ошибка при открытии файла: %v", err), logger)
		return
	}
	defer r.Close()

	var doc string
	// Парсим текст из docx файла, если получим ошибку, то надо запустить brokenXML парсер
	// Извлекаем текст из документа
	for {

		text, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			logErrorAndMoveFile(inputPath, errorDir, fmt.Sprintf("Ошибка при чтении из файла: %v", err), logger)
			return
		}

		doc += text + "\n"
	}

	// Сохраняем текст в файл
	if compress {
		// Упаковываем в GZ
		outputFile, err := os.Create(outputPath + ".gz")
		if err != nil {
			logErrorAndMoveFile(inputPath, errorDir, fmt.Sprintf("Ошибка при создании файла: %v", err), logger)
			return
		}
		defer outputFile.Close()

		// Настраиваем уровень сжатия
		gzWriter, err := gzip.NewWriterLevel(outputFile, compressionLevel)
		if err != nil {
			logErrorAndMoveFile(inputPath, errorDir, fmt.Sprintf("Ошибка при создании GZIP writer: %v", err), logger)
			return
		}
		defer gzWriter.Close()

		// Записываем текст в GZIP
		_, err = gzWriter.Write([]byte(doc))
		if err != nil {
			logErrorAndMoveFile(inputPath, errorDir, fmt.Sprintf("Ошибка при записи в GZIP файл: %v", err), logger)
			return
		}
	} else {
		// Сохраняем в TXT
		err = os.WriteFile(outputPath, []byte(doc), 0644)
		if err != nil {
			logErrorAndMoveFile(inputPath, errorDir, fmt.Sprintf("Ошибка при записи файла: %v", err), logger)
			return
		}
	}

	// Обновляем счетчик и выводим прогресс
	mutex.Lock()
	*counter++
	fmt.Printf("Обработан файл %d из %d: %s -> %s\n", *counter, totalFiles, inputPath, outputPath)
	mutex.Unlock()

}

// Функция для логирования ошибок и копирования проблемных файлов
func logErrorAndMoveFile(filePath, errorDir, errorMsg string, logger *log.Logger) {
	// Логируем ошибку
	logger.Println(errorMsg)

	// Копируем файл в папку с ошибками
	fileName := filepath.Base(filePath)
	destPath := filepath.Join(errorDir, fileName)

	// Открываем исходный файл
	srcFile, err := os.Open(filePath)
	if err != nil {
		logger.Printf("Ошибка при открытии файла %s для копирования: %v\n", filePath, err)
		return
	}
	defer srcFile.Close()

	// Создаем новый файл в папке с ошибками
	destFile, err := os.Create(destPath)
	if err != nil {
		logger.Printf("Ошибка при создании файла %s: %v\n", destPath, err)
		return
	}
	defer destFile.Close()

	// Копируем содержимое файла
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		logger.Printf("Ошибка при копировании файла %s: %v\n", filePath, err)
		return
	}

	logger.Printf("Файл %s успешно скопирован в папку с ошибками: %s\n", filePath, destPath)
}

func main() {
	// Определение флагов командной строки
	inputDir := flag.String("input", "", "Путь к папке с DOCX файлами")
	outputDir := flag.String("output", "", "Путь к папке для сохранения TXT/GZ файлов")
	maxGoroutines := flag.Int("goroutines", runtime.NumCPU(), "Максимальное количество одновременно работающих горутин (по умолчанию равно количеству ядер CPU)")
	compress := flag.Bool("compress", false, "Упаковывать ли выходные файлы в GZ (по умолчанию false)")
	compressionLevel := flag.Int("level", gzip.DefaultCompression, "Уровень сжатия GZIP (от 0 до 9, по умолчанию 6)")
	errorDir := flag.String("errors", "errors", "Путь к папке для сохранения файлов с ошибками")

	// Парсинг аргументов командной строки
	flag.Parse()

	// Проверка обязательных параметров
	if *inputDir == "" || *outputDir == "" {
		fmt.Println("Необходимо указать параметры -input и -output")
		flag.PrintDefaults()
		return
	}

	// Проверка уровня сжатия
	if *compressionLevel < gzip.NoCompression || *compressionLevel > gzip.BestCompression {
		fmt.Println("Уровень сжатия должен быть от 0 (без сжатия) до 9 (максимальное сжатие)")
		return
	}

	// Создаем папку для выходных файлов, если она не существует
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		os.Mkdir(*outputDir, 0755)
	}

	// Создаем папку для ошибок, если она не существует
	if _, err := os.Stat(*errorDir); os.IsNotExist(err) {
		os.Mkdir(*errorDir, 0755)
	}

	// Создаем лог-файл
	logFile, err := os.OpenFile("conversion_errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Ошибка при создании лог-файла: %v", err)
	}
	defer logFile.Close()

	// Настраиваем логгер
	logger := log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Получаем список всех DOCX файлов в папке
	files, err := filepath.Glob(filepath.Join(*inputDir, "*.docx"))
	if err != nil {
		logger.Fatalf("Ошибка при получении списка файлов: %v", err)
	}

	// Общее количество файлов
	totalFiles := len(files)
	if totalFiles == 0 {
		fmt.Println("В указанной папке нет DOCX файлов.")
		return
	}

	// Счетчик обработанных файлов
	var counter int
	var mutex sync.Mutex

	// Создаем WaitGroup для ожидания завершения всех горутин
	var wg sync.WaitGroup

	// Ограничиваем количество одновременно работающих горутин
	guard := make(chan struct{}, *maxGoroutines)

	for _, file := range files {
		wg.Add(1)
		guard <- struct{}{} // Блокируем, если уже запущено maxGoroutines горутин

		go func(inputPath string) {
			defer func() { <-guard }()
			// Формируем путь для сохранения файла
			outputFileName := filepath.Base(inputPath[:len(inputPath)-5] + ".txt")
			outputPath := filepath.Join(*outputDir, outputFileName)
			convertDocxToTxt(inputPath, outputPath, &wg, &counter, totalFiles, &mutex, *compress, *compressionLevel, *errorDir, logger)
		}(file)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()
	fmt.Printf("Обработка завершена. Всего обработано файлов: %d из %d.\n", counter, totalFiles)
}
