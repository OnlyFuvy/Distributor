import numpy as np
import librosa
import json
import sys
from pathlib import Path

def analyze_audio(file_path: str, max_points: int = 500) -> dict:
    """
    Анализирует аудиофайл и извлекает данные волны (Фурье).
    
    Args:
        file_path: Путь к аудиофайлу
        max_points: Максимальное количество точек данных (для оптимизации)
    
    Returns:
        Словарь с результатами анализа
    """
    try:
        # 1. Загружаем аудиофайл
        # sr=None сохраняет оригинальную частоту дискретизации
        y, sr = librosa.load(file_path, sr=None)
        
        # 2. Извлекаем огибающую сигнала (амплитуда во времени)
        # Это будет использоваться для визуализации волны
        envelope = librosa.onset.onset_strength(y=y, sr=sr)
        
        # 3. Выполняем быстрое преобразование Фурье (FFT)
        # Получаем частотные компоненты сигнала
        fft = np.fft.fft(y)
        fft_magnitude = np.abs(fft[:len(fft)//2])  # Берём только положительную половину
        
        # 4. Нормализуем данные (приводим к диапазону 0-1)
        if len(envelope) > 0:
            envelope = envelope / np.max(envelope)
        if len(fft_magnitude) > 0:
            fft_magnitude = fft_magnitude / np.max(fft_magnitude)
        
        # 5. Уменьшаем количество точек (даунсемплинг) для оптимизации
        # 500 точек достаточно для красивой визуализации на фронтенде
        if len(envelope) > max_points:
            step = len(envelope) // max_points
            envelope = envelope[::step][:max_points]
        
        if len(fft_magnitude) > max_points:
            step = len(fft_magnitude) // max_points
            fft_magnitude = fft_magnitude[::step][:max_points]
        
        # 6. Получаем длительность трека в секундах
        duration = librosa.get_duration(y=y, sr=sr)
        
        # 7. Формируем результат
        result = {
            "success": True,
            "duration": float(duration),
            "sample_rate": int(sr),
            "waveform_data": envelope.tolist(),  # Огибающая для визуализации
            "frequency_data": fft_magnitude.tolist(),  # Частоты (Фурье)
            "error": None
        }
        
        return result
        
    except Exception as e:
        # Возвращаем ошибку в том же формате
        return {
            "success": False,
            "duration": 0,
            "sample_rate": 0,
            "waveform_data": [],
            "frequency_data": [],
            "error": str(e)
        }


def main():
    """
    Точка входа при запуске скрипта из командной строки.
    Ожидает путь к файлу как аргумент.
    """
    if len(sys.argv) < 2:
        print(json.dumps({
            "success": False,
            "error": "No file path provided. Usage: python audio_analyzer.py <file_path>"
        }))
        sys.exit(1)
    
    file_path = sys.argv[1]
    
    # Проверяем, существует ли файл
    if not Path(file_path).exists():
        print(json.dumps({
            "success": False,
            "error": f"File not found: {file_path}"
        }))
        sys.exit(1)
    
    # Анализируем аудио и выводим результат как JSON
    result = analyze_audio(file_path)
    print(json.dumps(result))


if __name__ == "__main__":
    main()