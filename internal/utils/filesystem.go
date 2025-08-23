package utils

import (
    "fmt"
    "os"
    "path/filepath"
)

func FileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func DirectoryExists(dirname string) bool {
    info, err := os.Stat(dirname)
    if os.IsNotExist(err) {
        return false
    }
    return info.IsDir()
}

func EnsureDirectory(path string) error {
    if DirectoryExists(path) {
        return nil
    }
    
    if err := os.MkdirAll(path, 0755); err != nil {
        return fmt.Errorf("failed to create directory %s: %w", path, err)
    }
    
    return nil
}

func SafeWriteFile(filename string, content []byte, perm os.FileMode) error {
    dir := filepath.Dir(filename)
    if err := EnsureDirectory(dir); err != nil {
        return err
    }

    if FileExists(filename) {
        backup := filename + ".backup"
        if err := os.Rename(filename, backup); err != nil {
            return fmt.Errorf("failed to create backup %s: %w", backup, err)
        }
        fmt.Printf("📋 Created backup: %s\n", backup)
    }

    return os.WriteFile(filename, content, perm)
}

func GetAbsolutePath(path string) (string, error) {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
    }
    return absPath, nil
}