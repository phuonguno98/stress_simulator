// Package disk provides disk stress simulation capabilities.
//
//nolint:gosec
package disk

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/phuonguno98/stress_simulator/internal/types"
	"github.com/phuonguno98/stress_simulator/internal/utils"
)

const (
	tempDir = "temp"
)

// RunStress starts disk stress simulation with given parameters
func RunStress(avgUtilization float64, paramChan <-chan types.ParameterUpdate, stopChan <-chan os.Signal) {
	// Convert os.Signal channel to struct{} channel
	doneChan := make(chan struct{})
	go func() {
		<-stopChan
		close(doneChan)
	}()

	runStressInternal(avgUtilization, paramChan, doneChan)
}

// runStressInternal does the actual stress work with struct{} channel
func runStressInternal(avgUtilization float64, paramChan <-chan types.ParameterUpdate, stopChan <-chan struct{}) {
	// Create temporary directory for stress operations
	tempDirPath := tempDir

	// Ensure clean state from any previous crashed runs
	_ = os.RemoveAll(tempDirPath)
	err := os.MkdirAll(tempDirPath, 0750)
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(tempDirPath) }()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	fileHandles := make(map[string]*os.File)
	defer closeAllFiles(fileHandles)

	// Current utilization value
	currentUtil := avgUtilization
	// Track temporary overrides
	var tempOverrideEnd time.Time
	var tempMultiplier float64

	// Track random walk state
	var utilVariation float64

	// Ticker for disk cleanup (more aggressive check)
	cleanupTicker := time.NewTicker(10 * time.Second)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case update := <-paramChan:
			// Apply temporary override
			tempMultiplier = update.Value
			tempOverrideEnd = time.Now().Add(update.Duration)

			// Update current value with temporary multiplier
			currentUtil = avgUtilization * tempMultiplier
			if currentUtil > 100 {
				currentUtil = 100
			} else if currentUtil < 0 {
				currentUtil = 0
			}
		case <-ticker.C:
			// Check if temporary override has expired
			if time.Now().After(tempOverrideEnd) && !tempOverrideEnd.IsZero() {
				// Reset to base value
				currentUtil = avgUtilization
				tempOverrideEnd = time.Time{} // Clear the override time
			}

			// Random walk for smooth disk variation
			utilStep := (rand.Float64() - 0.5) * 0.04
			utilVariation += utilStep
			if utilVariation > 0.2 {
				utilVariation = 0.2
			} else if utilVariation < -0.2 {
				utilVariation = -0.2
			}

			// Calculate current disk utilization with smooth variation
			finalUtil := currentUtil + (currentUtil * utilVariation)
			if finalUtil > 100 {
				finalUtil = 100
			} else if finalUtil < 0 {
				finalUtil = 0
			}

			// Perform disk operations based on utilization
			performDiskOperations(finalUtil, fileHandles, tempDirPath)
		case <-cleanupTicker.C:
			// Enforce disk limit (e.g., 1000MB) to prevent full disk
			enforceDiskLimit(tempDirPath, fileHandles, 1000*1024*1024)
		}
	}
}

// performDiskOperations performs various disk operations to simulate utilization
func performDiskOperations(utilizationPercent float64, fileHandles map[string]*os.File, tempDir string) {
	// Determine how many operations to perform based on utilization
	operationCount := max(
		// Scale operations with utilization
		int(utilizationPercent/10.0), 1)

	for range operationCount {
		// Randomly select a disk operation
		opType := rand.Intn(4) // 0: read, 1: write, 2: random I/O, 3: file creation/deletion

		switch opType {
		case 0: // Read operation
			performReadOperation(fileHandles)
		case 1: // Write operation
			performWriteOperation(fileHandles, tempDir)
		case 2: // Random I/O operation
			performRandomIO(fileHandles, tempDir)
		case 3: // File creation/deletion
			performFileOps(tempDir)
		}

		// Small delay to control intensity
		time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)
	}
}

// performReadOperation performs read operations on existing files
func performReadOperation(fileHandles map[string]*os.File) {
	// Try to open and read from a file
	if len(fileHandles) > 0 {
		// Pick a random file handle
		keys := make([]string, 0, len(fileHandles))
		for k := range fileHandles {
			keys = append(keys, k)
		}

		if len(keys) > 0 {
			filename := keys[rand.Intn(len(keys))]
			file, exists := fileHandles[filename]
			if exists && file != nil {
				// Read some data from the file
				buf := make([]byte, 1024)
				_, _ = file.ReadAt(buf, int64(rand.Intn(10000))) // Ignoring error as this is for stress testing
			}
		}
	}
}

// performWriteOperation performs write operations
func performWriteOperation(fileHandles map[string]*os.File, tempDir string) {
	filename := filepath.Join(tempDir, "stress_file_"+string(rune(rand.Intn(26)+97))+".dat")

	var file *os.File
	var exists bool

	if file, exists = fileHandles[filename]; !exists {
		// Create new file if it doesn't exist
		var err error
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0644)
		if err != nil {
			return
		}
		fileHandles[filename] = file
	}

	// Write random data to the file using shared payload
	dataSize := 1024 + rand.Intn(4096) // 1KB to 5KB
	data := utils.GetPayload(dataSize)

	_, err := file.Write(data)
	if err != nil {
		// If write fails, remove the file handle
		delete(fileHandles, filename)
		_ = os.Remove(filename)
	}
}

// performRandomIO performs mixed read/write operations
func performRandomIO(fileHandles map[string]*os.File, tempDir string) {
	filename := filepath.Join(tempDir, "random_io_"+string(rune(rand.Intn(26)+97))+".dat")

	var file *os.File
	var exists bool

	if file, exists = fileHandles[filename]; !exists {
		// Create new file if it doesn't exist
		var err error
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
		if err != nil {
			return
		}
		fileHandles[filename] = file
	}

	// Perform random seek and write/read operations
	operation := rand.Intn(2)
	if operation == 0 {
		// Write operation at random position
		pos := int64(rand.Intn(50000))
		_, err := file.Seek(pos, 0)
		if err != nil {
			return // If seek fails, skip this operation
		}

		dataSize := 512 + rand.Intn(1024)
		data := utils.GetPayload(dataSize)

		_, err = file.Write(data)
		if err != nil {
			return // If write fails, skip this operation
		}
	} else {
		// Read operation at random position
		pos := int64(rand.Intn(50000))
		_, err := file.Seek(pos, 0)
		if err != nil {
			return // If seek fails, skip this operation
		}

		buf := make([]byte, 512)
		_, err = file.Read(buf)
		if err != nil {
			return // If read fails, skip this operation
		}
	}
}

// performFileOps performs file creation and deletion operations
func performFileOps(tempDir string) {
	operation := rand.Intn(2)
	if operation == 0 {
		// Create a new file
		filename := filepath.Join(tempDir, "temp_"+string(rune(rand.Intn(10000)))+".tmp")
		file, err := os.Create(filename)
		if err == nil {
			// Write some random data
			data := utils.GetPayload(100 + rand.Intn(500))
			_, _ = file.Write(data)
			_ = file.Close()
		}
	} else {
		// Delete a random file
		files, err := os.ReadDir(tempDir)
		if err == nil && len(files) > 0 {
			randomFile := files[rand.Intn(len(files))]
			if !randomFile.IsDir() {
				_ = os.Remove(filepath.Join(tempDir, randomFile.Name()))
			}
		}
	}
}

// closeAllFiles closes all open file handles
func closeAllFiles(fileHandles map[string]*os.File) {
	for _, file := range fileHandles {
		if file != nil {
			_ = file.Close()
		}
	}
}

// enforceDiskLimit checks the total size of the temp directory and deletes files if it exceeds maxBytes
func enforceDiskLimit(dirPath string, fileHandles map[string]*os.File, maxBytes int64) {
	var totalSize int64
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	// Calculate total size
	for _, f := range files {
		if info, err := f.Info(); err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
	}

	// If over limit, we need to clean up
	if totalSize > maxBytes {
		// Close all handles first so files can be deleted (important on Windows)
		closeAllFiles(fileHandles)
		for k := range fileHandles {
			delete(fileHandles, k)
		}

		// Delete all files in directory
		for _, f := range files {
			if !f.IsDir() {
				_ = os.Remove(filepath.Join(dirPath, f.Name()))
			}
		}
	}
}
