package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// AudioAnalysisResult — структура для парсинга ответа от Python-скрипта
type AudioAnalysisResult struct {
	Success       bool      `json:"success"`
	Duration      float64   `json:"duration"`
	SampleRate    int       `json:"sample_rate"`
	WaveformData  []float64 `json:"waveform_data"`
	FrequencyData []float64 `json:"frequency_data"`
	Error         string    `json:"error"`
}

// AnalyzeAudioFile — вызывает Python-скрипт для анализа аудио
func AnalyzeAudioFile(filePath string) (*AudioAnalysisResult, error) {
	// 1. Определяем путь к скрипту (относительно корня проекта)
	//    __FILE__ не существует в Go, поэтому используем относительный путь
	scriptPath := filepath.Join("python", "audio_analyzer.py")

	// 2. Формируем команду: py python/audio_analyzer.py <file_path>
	//    Используем "py" для Windows (Python Launcher), "python3" для Linux/Mac
	cmdName := "py"
	if runtime.GOOS != "windows" {
		cmdName = "python3"
	}

	cmd := exec.Command(cmdName, scriptPath, filePath)

	// 3. Запускаем и собираем вывод
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run python script: %w, stderr: %s", err, stderr.String())
	}

	// 4. Парсим JSON-ответ от Python
	var result AudioAnalysisResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse python output: %w, output: %s", err, stdout.String())
	}

	// 5. Проверяем, не вернул ли скрипт ошибку
	if !result.Success {
		return nil, fmt.Errorf("python script error: %s", result.Error)
	}

	return &result, nil
}
