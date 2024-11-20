package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Groups []Group `yaml:"groups"`
}

type Rules struct {
	Include []string `yaml:"include"`
	Ignore  []string `yaml:"ignore"`
}

type Group struct {
	Name    string        `yaml:"name"`
	Paths   Rules         `yaml:"paths"`
	Files   FileRules     `yaml:"files"`
	Texts   TextRules     `yaml:"texts"`
	Entropy EntropyConfig `yaml:"entropy"`
}

type FileRules struct {
	Types    Rules `yaml:"types"`
	Names    Rules `yaml:"names"`
	Patterns Rules `yaml:"patterns"`
}

type TextRules struct {
	Keywords Rules `yaml:"keywords"`
	Patterns Rules `yaml:"patterns"`
}

type EntropyConfig struct {
	Enable    bool    `yaml:"enable"`
	Threshold float64 `yaml:"threshold"`
}

func LoadConfig(filePath string) (*Config, error) {
	// please normalize using this formatting.
	// I hate treating error handling as part of the MAIN code.
	//Yes its important but as someone who reads a code I want to quickly understand the main functinoality.

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func getAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println(err)
	}
	return absPath
}

func isIgnored(path string, ignorePaths []string) bool {

	for _, ignorePath := range ignorePaths {
		if getAbs(path) == getAbs(ignorePath) {
			return true
		}
	}
	return false
}

func checkPathRecursively(path string, config Group) {
	//get the needed config
	ignorePaths := config.Paths.Ignore

	//fmt.Printf("you entered %s , %v\n", path, ignorePaths)
	if isIgnored(path, ignorePaths) {
		//fmt.Println("Ignoring path:", path)
		return
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error checking path:", err)
		return
	}

	if fileInfo.IsDir() {
		err := filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if isIgnored(subPath, ignorePaths) {
				//fmt.Println("Ignoring directory:", subPath)
				return filepath.SkipDir
			}

			if !info.IsDir() {
				if checkFile(subPath, config) {
					//fmt.Println("lets scan file", subPath)
					scanFile(subPath, config)
				}
			}

			return nil
		})

		if err != nil {

			fmt.Println("Error walking the path:", err)
		}
	} else {
		if checkFile(path, config) {
			//fmt.Println("lets scan file", path)
			scanFile(path, config)
		}
	}
}
func scanFile(path string, config Group) {
	//fmt.Println("processing", path)
	report := ""
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", path, err)
		return
	}
	defer file.Close()

	lineNumber := 1
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			lineNumber++
			continue
		}
		keywordFound := false
		if len(config.Texts.Keywords.Include) == 0 || (len(config.Texts.Keywords.Include) == 1 && config.Texts.Keywords.Include[0] == "*") {
			fmt.Println("anything could be accepted")
			keywordFound = false
		} else {
			for _, keyword := range config.Texts.Keywords.Include {
				// fmt.Println("checking ", keyword)
				if strings.Contains(line, keyword) {
					//fmt.Println("secret found")
					keywordFound = true
					break
				}
			}
		}

		if keywordFound {
			for _, ignoreKeyword := range config.Texts.Keywords.Ignore {
				if strings.Contains(line, ignoreKeyword) {
					fmt.Println("ignore found. ignored")
					keywordFound = false
					break
				}
			}
		}

		// if !keywordFound {
		// 	lineNumber++
		// 	continue
		// }

		// Pattern matching
		patternMatch := false
		if len(config.Texts.Patterns.Include) == 0 {
			// fmt.Println("pattern found because no pattern")
			patternMatch = true
		}
		for _, pattern := range config.Texts.Patterns.Include {
			re := regexp.MustCompile(pattern)
			if re.MatchString(line) {
				patternMatch = true
				break
			}
		}
		for _, ignorePattern := range config.Texts.Patterns.Ignore {
			re := regexp.MustCompile(ignorePattern)
			if re.MatchString(line) {
				// fmt.Println("found ignore pattern, ignored")
				patternMatch = false
				break
			}
		}

		entropyOk := true
		if config.Entropy.Enable {
			entropyOk = false
			words := strings.Fields(line)
			for _, word := range words {
				entropy := calculateEntropy(word)
				if entropy < config.Entropy.Threshold {
					entropyOk = false
					break
				}
			}
		}

		if keywordFound && patternMatch && entropyOk {
			report += fmt.Sprintf("Found potential secret in file %s at line %d: %s\n", path, lineNumber, line)
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", path, err)
	}

	if report != "" {
		fmt.Println(report)
	} else {
		fmt.Println("No secrets found.")
	}

}

func calculateEntropy(input string) float64 {
	freq := make(map[rune]float64)
	for _, r := range input {
		freq[r]++
	}
	entropy := 0.0
	length := float64(len(input))
	for _, count := range freq {
		p := count / length
		entropy -= p * (math.Log2(p))
	}
	return entropy
}

func main() {
	configFile := flag.String("c", "config.yaml", "The configuration file, if not specified uses secretInvasionConfig environment variable or config.yaml in current directory")
	flag.Parse()

	_, err := os.Stat(*configFile)
	if os.IsNotExist(err) {
		fmt.Println("Specified config file doesn't exist, using environment variable secretInvasionConfig")

		envFile := os.Getenv("secretInvasionConfig")
		if envFile == "" {
			fmt.Println("Environment variable secretInvasionConfig is not set, exiting")
			os.Exit(1)
		}

		*configFile = envFile

		_, err := os.Stat(*configFile)
		if os.IsNotExist(err) {
			fmt.Println("Could not find the file specified by secretInvasionConfig. Exiting")
			os.Exit(1)
		}
	}

	config, err := LoadConfig(*configFile)
	if err != nil {
		panic(err)
	}

	for _, group := range config.Groups {
		fmt.Println("Group:", group.Name)
		for _, path := range group.Paths.Include {
			checkPathRecursively(path, group)
		}
	}
}

func checkFileExtension(path string, include []string, ignore []string) bool {
	//fmt.Println("checking file extension", path)
	ext := filepath.Ext(path)
	for _, i := range ignore {
		if ext == i {
			return false
		}
	}
	if len(include) == 0 {
		return true
	}
	for _, i := range include {
		// if i == "*" {
		// 	return true
		// }
		if ext == i {
			return true
		}
	}
	return false
}

func checkFileName(path string, include []string, ignore []string) bool {
	fileName := filepath.Base(path)
	// fmt.Println("checking file name", fileName, include)
	for _, i := range ignore {
		if fileName == i {
			return false
		}
	}
	if len(include) == 0 {
		return true
	}
	for _, i := range include {
		// if i == "*" {
		// 	return true
		// }
		if fileName == i {
			// fmt.Println("file name matched")
			return true
		}
	}
	return false
}

func checkFilePattern(path string, include []string, ignore []string) bool {
	fileName := filepath.Base(path)

	for _, pattern := range ignore {
		matched, err := regexp.MatchString(pattern, fileName)
		if err == nil && matched {
			return false
		}
	}
	for _, pattern := range include {
		matched, err := regexp.MatchString(pattern, fileName)
		if err == nil && matched {
			return true
		}
	}
	return len(include) == 0

}

func checkFile(path string, config Group) bool {
	if !checkFileExtension(path, config.Files.Types.Include, config.Files.Types.Ignore) {
		return false
	}
	if !checkFileName(path, config.Files.Names.Include, config.Files.Names.Ignore) {
		return false
	}
	if !checkFilePattern(path, config.Files.Patterns.Include, config.Files.Patterns.Ignore) {
		return false
	}
	return true
}
