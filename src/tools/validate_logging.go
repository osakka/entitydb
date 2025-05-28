package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LogIssue represents a logging standard violation
type LogIssue struct {
	File    string
	Line    int
	Issue   string
	Content string
}

func main() {
	fmt.Println("EntityDB Logging Standards Validator")
	fmt.Println("====================================")
	
	issues := []LogIssue{}
	
	// Walk through src directory
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip non-Go files and test files
		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "_test.go") {
			return nil
		}
		
		// Skip vendor, tools, and test directories
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "/tools/") || strings.Contains(path, "/tests/") {
			return nil
		}
		
		// Check the file
		fileIssues := checkFile(path)
		issues = append(issues, fileIssues...)
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
	
	// Report findings
	if len(issues) == 0 {
		fmt.Println("\n✅ No logging issues found!")
	} else {
		fmt.Printf("\n❌ Found %d logging issues:\n\n", len(issues))
		
		// Group by file
		fileMap := make(map[string][]LogIssue)
		for _, issue := range issues {
			fileMap[issue.File] = append(fileMap[issue.File], issue)
		}
		
		for file, fileIssues := range fileMap {
			fmt.Printf("📄 %s (%d issues)\n", file, len(fileIssues))
			for _, issue := range fileIssues {
				fmt.Printf("   Line %d: %s\n", issue.Line, issue.Issue)
				if issue.Content != "" {
					fmt.Printf("   > %s\n", strings.TrimSpace(issue.Content))
				}
			}
			fmt.Println()
		}
	}
}

func checkFile(path string) []LogIssue {
	issues := []LogIssue{}
	
	file, err := os.Open(path)
	if err != nil {
		return issues
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	// Patterns to check
	wrongLoggerMethods := regexp.MustCompile(`logger\.(DEBUG|INFO|WARN|ERROR|TRACE)\(`)
	directLogPackage := regexp.MustCompile(`\blog\.(Print|Printf|Println)\(`)
	redundantInfo := regexp.MustCompile(`logger\.\w+\("[^"]*(?:Function:|File:|Line:)[^"]*"`)
	vagueMessages := regexp.MustCompile(`logger\.\w+\("(?:Starting|Stopping|Processing|Done|Error|Failed|Success)\.?"[,)]`)
	fmtPrintInCore := regexp.MustCompile(`fmt\.(Print|Printf|Println)\(`)
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		
		// Check for wrong logger method names
		if wrongLoggerMethods.MatchString(line) {
			issues = append(issues, LogIssue{
				File:    path,
				Line:    lineNum,
				Issue:   "Incorrect logger method name (should be Debug, not DEBUG)",
				Content: line,
			})
		}
		
		// Check for direct log package usage (except in logger package)
		if !strings.Contains(path, "logger/") && directLogPackage.MatchString(line) {
			issues = append(issues, LogIssue{
				File:    path,
				Line:    lineNum,
				Issue:   "Using log package directly instead of logger package",
				Content: line,
			})
		}
		
		// Check for redundant information
		if redundantInfo.MatchString(line) {
			issues = append(issues, LogIssue{
				File:    path,
				Line:    lineNum,
				Issue:   "Log message contains redundant file/function/line info",
				Content: line,
			})
		}
		
		// Check for vague messages
		if vagueMessages.MatchString(line) {
			issues = append(issues, LogIssue{
				File:    path,
				Line:    lineNum,
				Issue:   "Vague log message - be more specific",
				Content: line,
			})
		}
		
		// Check for fmt.Print in core code (not tools, tests, or cmd utilities)
		if !strings.Contains(path, "/tools/") && 
		   !strings.Contains(path, "/tests/") &&
		   !strings.Contains(path, "/cmd/") &&
		   !strings.Contains(path, "main.go") && 
		   fmtPrintInCore.MatchString(line) {
			issues = append(issues, LogIssue{
				File:    path,
				Line:    lineNum,
				Issue:   "Using fmt.Print in core code - use logger instead",
				Content: line,
			})
		}
	}
	
	return issues
}