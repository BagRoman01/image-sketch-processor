package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BagRoman01/image-sketch-processor/internal/logging"
)

type ImageProcessor struct {
	Config PrimitiveConfig
}

type PrimitiveConfig struct {
	NumShapes   int    // -n: количество фигур
	Mode        int    // -m: тип фигур (1=треугольники, 2=прямоугольники, 3=эллипсы, 4=круги, 5=rotatedrect, 6=beziers, 7=rotatedellipse, 8=polygon)
	Alpha       int    // -a: прозрачность (0-255, 0=auto)
	Repeat      int    // -rep: доп. попытки для сложных фигур
	Resize      int    // -r: ресайз перед обработкой
	OutputSize  int    // -s: размер выходного изображения
	Background  string // -bg: фоновый цвет (hex или "avg", "white", "black")
	Workers     int    // -j: количество потоков (0=все ядра)
	Verbose     bool   // -v: подробный вывод
	VeryVerbose bool   // -vv: очень подробный вывод
}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		Config: PrimitiveConfig{
			NumShapes:   150,
			Mode:        1,   // треугольники
			Alpha:       125, // полупрозрачные
			Repeat:      0,
			Resize:      256,   // быстрая обработка
			OutputSize:  1024,  // хорошее разрешение
			Background:  "avg", // средний цвет фона
			Workers:     0,     // все ядра
			Verbose:     false,
			VeryVerbose: false,
		},
	}
}

// CreatePencilSketch - обёртка над primitive CLI с полной конфигурацией
func (p *ImageProcessor) CreatePencilSketch(
	ctx context.Context,
	fileData []byte,
) ([]byte, error) {
	logger := logging.LoggerFromContext(ctx)

	// 1. Создаём временные файлы
	tmpDir, err := os.MkdirTemp("", "primitive-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Определяем расширение входного файла (пробуем угадать формат)
	ext := ".png" // по умолчанию
	if len(fileData) > 8 {
		// Простая проверка сигнатур
		if fileData[0] == 0xFF && fileData[1] == 0xD8 {
			ext = ".jpg"
		} else if fileData[0] == 0x89 && fileData[1] == 0x50 {
			ext = ".png"
		}
	}

	inputPath := filepath.Join(tmpDir, "input"+ext)
	outputPath := filepath.Join(tmpDir, "output.png")

	// 2. Сохраняем входные данные
	if err := os.WriteFile(inputPath, fileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// 3. Формируем аргументы для primitive
	args := []string{
		"-i", inputPath,
		"-o", outputPath,
		"-n", fmt.Sprint(p.Config.NumShapes),
		"-m", fmt.Sprint(p.Config.Mode),
		"-a", fmt.Sprint(p.Config.Alpha),
		"-r", fmt.Sprint(p.Config.Resize),
		"-s", fmt.Sprint(p.Config.OutputSize),
		"-bg", p.Config.Background,
	}

	// Добавляем опциональные параметры
	if p.Config.Repeat > 0 {
		args = append(args, "-rep", fmt.Sprint(p.Config.Repeat))
	}

	if p.Config.Workers > 0 {
		args = append(args, "-j", fmt.Sprint(p.Config.Workers))
	}

	if p.Config.Verbose {
		args = append(args, "-v")
	}

	if p.Config.VeryVerbose {
		args = append(args, "-vv")
	}

	// 4. Запускаем primitive
	cmd := exec.Command("primitive", args...)

	if p.Config.Verbose {
		logger.Info("running primitive", "args", args)
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("primitive failed: %w", err)
	}

	// 5. Читаем результат
	result, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	logger.Info("primitive sketch created",
		"shapes", p.Config.NumShapes,
		"mode", p.Config.Mode,
		"size", len(result))

	return result, nil
}

// SetStyle - быстрая настройка художественных стилей
func (p *ImageProcessor) SetStyle(style string) {
	switch style {
	case "lowpoly":
		p.Config.Mode = 1
		p.Config.NumShapes = 150
		p.Config.Alpha = 255
		p.Config.Background = "white"
	case "sketch":
		p.Config.Mode = 1
		p.Config.NumShapes = 50
		p.Config.Alpha = 255
		p.Config.Background = "white"
	case "impressionism":
		p.Config.Mode = 3
		p.Config.NumShapes = 200
		p.Config.Alpha = 80
		p.Config.Background = "white"
	case "pointillism":
		p.Config.Mode = 4
		p.Config.NumShapes = 300
		p.Config.Alpha = 80
		p.Config.Background = "white"
	case "abstract":
		p.Config.Mode = 2
		p.Config.NumShapes = 80
		p.Config.Alpha = 0
		p.Config.Background = "black"
	case "portrait":
		p.Config.Mode = 1
		p.Config.NumShapes = 150
		p.Config.Alpha = 255
		p.Config.Resize = 256
		p.Config.OutputSize = 1024
		p.Config.Background = "white"
	default:
		// Сброс на стандартные настройки
		p.Config.NumShapes = 100
		p.Config.Mode = 1
		p.Config.Alpha = 128
		p.Config.Repeat = 0
		p.Config.Resize = 256
		p.Config.OutputSize = 1024
		p.Config.Background = "avg"
		p.Config.Workers = 0
		p.Config.Verbose = false
		p.Config.VeryVerbose = false
	}
}

// SetPortraitStyle - быстрая настройка специально для портретов
func (p *ImageProcessor) SetPortraitStyle(detail string) {
	switch detail {
	case "high":
		p.Config.NumShapes = 200
		p.Config.Mode = 1
		p.Config.Alpha = 255
		p.Config.Resize = 256
		p.Config.OutputSize = 2048
		p.Config.Background = "white"
	case "medium":
		p.Config.NumShapes = 150
		p.Config.Mode = 1
		p.Config.Alpha = 255
		p.Config.Resize = 256
		p.Config.OutputSize = 1024
		p.Config.Background = "white"
	case "low":
		p.Config.NumShapes = 80
		p.Config.Mode = 1
		p.Config.Alpha = 255
		p.Config.Resize = 256
		p.Config.OutputSize = 800
		p.Config.Background = "white"
	default:
		p.SetStyle("portrait")
	}
}
