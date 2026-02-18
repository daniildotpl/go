package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL = "http://srv.msk01.gigacorp.local/_stats"
	// Пороговые значения
	loadAvgThreshold    = 30.0
	memoryUsagePercent  = 80.0
	diskUsagePercent    = 90.0
	networkUsagePercent = 90.0
	// Количество ошибок для вывода сообщения о недоступности
	errorThreshold = 3
)

type ServerStats struct {
	LoadAvg      float64
	MemoryTotal  uint64
	MemoryUsed   uint64
	DiskTotal    uint64
	DiskUsed     uint64
	NetworkTotal uint64
	NetworkUsed  uint64
}

func main() {
	errorCount := 0
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			fmt.Printf("Error fetching stats: %v\n", err)
			if errorCount >= errorThreshold {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0 // Сбрасываем счетчик после сообщения
			}
			continue
		}

		// Сбрасываем счетчик ошибок при успешном получении данных
		errorCount = 0

		// Проверяем пороговые значения
		checkThresholds(stats)
	}
}

func fetchStats() (*ServerStats, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Парсим CSV данные
	values := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(values) != 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(values))
	}

	stats := &ServerStats{}

	// Парсим каждое значение
	if stats.LoadAvg, err = strconv.ParseFloat(strings.TrimSpace(values[0]), 64); err != nil {
		return nil, fmt.Errorf("failed to parse load average: %v", err)
	}

	if stats.MemoryTotal, err = strconv.ParseUint(strings.TrimSpace(values[1]), 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse memory total: %v", err)
	}

	if stats.MemoryUsed, err = strconv.ParseUint(strings.TrimSpace(values[2]), 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse memory used: %v", err)
	}

	if stats.DiskTotal, err = strconv.ParseUint(strings.TrimSpace(values[3]), 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse disk total: %v", err)
	}

	if stats.DiskUsed, err = strconv.ParseUint(strings.TrimSpace(values[4]), 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse disk used: %v", err)
	}

	if stats.NetworkTotal, err = strconv.ParseUint(strings.TrimSpace(values[5]), 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse network total: %v", err)
	}

	if stats.NetworkUsed, err = strconv.ParseUint(strings.TrimSpace(values[6]), 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse network used: %v", err)
	}

	return stats, nil
}

func checkThresholds(stats *ServerStats) {
	// Проверка Load Average
	if stats.LoadAvg > loadAvgThreshold {
		fmt.Printf("Load Average is too high: %.2f\n", stats.LoadAvg)
	}

	// Проверка использования памяти
	if stats.MemoryTotal > 0 {
		memoryUsedPercent := (float64(stats.MemoryUsed) / float64(stats.MemoryTotal)) * 100
		if memoryUsedPercent > memoryUsagePercent {
			fmt.Printf("Memory usage too high: %.0f%%\n", memoryUsedPercent)
		}
	}

	// Проверка использования диска
	if stats.DiskTotal > 0 {
		diskUsedPercent := (float64(stats.DiskUsed) / float64(stats.DiskTotal)) * 100
		if diskUsedPercent > diskUsagePercent {
			freeDiskBytes := stats.DiskTotal - stats.DiskUsed
			freeDiskMB := freeDiskBytes / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskMB)
		}
	}

	// Проверка использования сети
	if stats.NetworkTotal > 0 {
		networkUsedPercent := (float64(stats.NetworkUsed) / float64(stats.NetworkTotal)) * 100
		if networkUsedPercent > networkUsagePercent {
			freeNetworkBytes := stats.NetworkTotal - stats.NetworkUsed
			// Переводим в мегабиты в секунду (1 байт = 8 бит)
			freeNetworkMbits := (float64(freeNetworkBytes) * 8) / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeNetworkMbits)
		}
	}
}
