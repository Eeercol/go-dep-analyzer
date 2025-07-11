package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// DepInfo хранит информацию по зависимости
type DepInfo struct {
	Path    string // путь модуля
	Current string // текущая версия
	Latest  string // доступная новая версия (если есть)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		_, err := fmt.Fprintf(os.Stderr, "Использование: %s <git_repo_url>\n", os.Args[0])
		if err != nil {
			return
		}
		os.Exit(1)
	}
	repoURL := args[0]

	// Клонируем репозиторий во временную папку
	tmpDir, err := cloneRepo(repoURL)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Ошибка клонирования: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {

		}
	}(tmpDir) // удаляем папку после выполнения

	// Анализируем модуль: имя, версия Go, прямые зависимости
	moduleName, goVersion, deps, err := analyzeModule(tmpDir)
	if err != nil {
		_, err2 := fmt.Fprintf(os.Stderr, "Ошибка анализа модуля: %v\n", err)
		if err2 != nil {
			return
		}
		os.Exit(1)
	}

	// Для каждой прямой зависимости узнаем, есть ли более новая версия
	for i := range deps {
		if latest, err := getLatestVersion(tmpDir, deps[i].Path); err == nil {
			deps[i].Latest = latest
		}
	}

	// Выводим результаты на консоль
	fmt.Printf("Модуль: %s\n", moduleName)
	fmt.Printf("Версия Go: %s\n", goVersion)
	fmt.Println("Зависимости:")

	if len(deps) == 0 {
		fmt.Println("  (прямых зависимостей не найдено)")
	}

	for _, d := range deps {
		if d.Latest != "" && d.Latest != d.Current {
			fmt.Printf("- %s: текущая %s → доступна %s\n", d.Path, d.Current, d.Latest)
		} else {
			fmt.Printf("- %s: версия %s (обновления нет)\n", d.Path, d.Current)
		}
	}
}

// cloneRepo клонирует git-репозиторий во временную папку и возвращает путь
func cloneRepo(url string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "gomod-*")
	if err != nil {
		return "", fmt.Errorf("не удалось создать временную директорию: %v", err)
	}
	cmd := exec.Command("git", "clone", url, tmpDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone не удался: %v", err)
	}
	return tmpDir, nil
}

// analyzeModule читает go.mod, возвращает Имя модуля, версию Go и список прямых зависимостей
func analyzeModule(dir string) (string, string, []DepInfo, error) {
	filePath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", nil, fmt.Errorf("не найден go.mod: %v", err)
	}

	mf, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", "", nil, fmt.Errorf("ошибка парсинга go.mod: %v", err)
	}

	modPath := ""
	if mf.Module != nil {
		modPath = mf.Module.Mod.Path
	}

	goVer := ""
	if mf.Go != nil {
		goVer = mf.Go.Version
	}

	var deps []DepInfo
	for _, req := range mf.Require {
		// только прямые зависимости
		if req.Indirect {
			continue
		}
		deps = append(deps, DepInfo{
			Path:    req.Mod.Path,
			Current: req.Mod.Version,
		})
	}

	return modPath, goVer, deps, nil
}

// getLatestVersion использует go list -m -u -json, чтобы получить последнюю версию модуля
func getLatestVersion(dir, modulePath string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "-u", "-json", modulePath)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// временная структура для разбора JSON
	var info struct {
		Update *struct {
			Version string `json:"Version"`
		} `json:"Update"`
	}

	if err := json.Unmarshal(out, &info); err != nil {
		return "", err
	}
	if info.Update != nil {
		return info.Update.Version, nil
	}
	return "", nil
}
