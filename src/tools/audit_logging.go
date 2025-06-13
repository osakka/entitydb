//go:build tool
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type LogIssue struct {
	File    string
	Line    int
	Level   string
	Message string
	Issue   string
}

var (
	// Patterns to detect
	logPattern = regexp.MustCompile(`logger\.(Trace|Debug|Info|Warn|Error|Fatal|Panic)f?\s*\(\s*"([^"]*)"`)
	
	// Bad patterns
	badPatterns = []struct {
		pattern *regexp.Regexp
		issue   string
	}{
		{regexp.MustCompile(`(?i)success\s*:`), "Avoid 'SUCCESS:' prefix"},
		{regexp.MustCompile(`\[[A-Z_]+\]`), "Avoid [TAG] format in messages"},
		{regexp.MustCompile(`!!!+`), "Avoid excessive punctuation"},
		{regexp.MustCompile(`^\s*in\s+function\s+`), "Redundant 'in function' - already in log context"},
		{regexp.MustCompile(`^\s*(starting|stopping|loading|saving)\.{3,}`), "Avoid trailing ellipsis"},
		{regexp.MustCompile(`(?i)error\s+occurred`), "Be specific about the error"},
		{regexp.MustCompile(`\s{2,}`), "Avoid multiple consecutive spaces"},
		{regexp.MustCompile(`^\s*[A-Z_]+:`), "Avoid PREFIX: format"},
	}
	
	// Level appropriateness checks
	levelChecks = map[string][]string{
		"Trace": {"entry", "exit", "value=", "state="},
		"Debug": {"attempting", "checking", "processing", "found"},
		"Info":  {"started", "stopped", "created", "deleted", "completed"},
		"Warn":  {"failed to", "retrying", "exceeded", "degraded", "missing"},
		"Error": {"failed", "error", "cannot", "unable", "invalid"},
	}
)

func main() {
	issues := []LogIssue{}
	
	// Walk through all Go files
	err := filepath.Walk("/opt/entitydb/src", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip non-Go files and test files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		// Skip vendor and generated files
		if strings.Contains(path, "/vendor/") || strings.Contains(path, ".pb.go") {
			return nil
		}
		
		// Analyze file
		fileIssues := analyzeFile(path)
		issues = append(issues, fileIssues...)
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
	
	// Report findings
	fmt.Println("EntityDB Logging Audit Report")
	fmt.Println("=============================")
	fmt.Printf("Total issues found: %d\n\n", len(issues))
	
	// Group by issue type
	issueTypes := make(map[string][]LogIssue)
	for _, issue := range issues {
		issueTypes[issue.Issue] = append(issueTypes[issue.Issue], issue)
	}
	
	// Report by issue type
	for issueType, typeIssues := range issueTypes {
		fmt.Printf("\n%s (%d occurrences):\n", issueType, len(typeIssues))
		for i, issue := range typeIssues {
			if i < 5 { // Show first 5 examples
				fmt.Printf("  %s:%d [%s] %s\n", 
					filepath.Base(issue.File), issue.Line, issue.Level, issue.Message)
			}
		}
		if len(typeIssues) > 5 {
			fmt.Printf("  ... and %d more\n", len(typeIssues)-5)
		}
	}
	
	// Summary recommendations
	fmt.Println("\nRecommendations:")
	fmt.Println("1. Update all log messages to follow the new standards")
	fmt.Println("2. Remove redundant context (file/function already logged)")
	fmt.Println("3. Ensure appropriate log levels for the audience")
	fmt.Println("4. Make messages concise and actionable")
	fmt.Println("5. Use consistent terminology throughout")
}

func analyzeFile(path string) []LogIssue {
	issues := []LogIssue{}
	
	file, err := os.Open(path)
	if err != nil {
		return issues
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Find log statements
		matches := logPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				level := match[1]
				message := match[2]
				
				// Check for bad patterns
				for _, bad := range badPatterns {
					if bad.pattern.MatchString(message) {
						issues = append(issues, LogIssue{
							File:    path,
							Line:    lineNum,
							Level:   level,
							Message: message,
							Issue:   bad.issue,
						})
					}
				}
				
				// Check level appropriateness
				if keywords, ok := levelChecks[level]; ok {
					appropriate := false
					messageLower := strings.ToLower(message)
					for _, keyword := range keywords {
						if strings.Contains(messageLower, keyword) {
							appropriate = true
							break
						}
					}
					if !appropriate && level != "Trace" { // Trace can be more flexible
						issues = append(issues, LogIssue{
							File:    path,
							Line:    lineNum,
							Level:   level,
							Message: message,
							Issue:   fmt.Sprintf("Message may not be appropriate for %s level", level),
						})
					}
				}
				
				// Check message length
				if len(message) > 100 {
					issues = append(issues, LogIssue{
						File:    path,
						Line:    lineNum,
						Level:   level,
						Message: message,
						Issue:   "Message too long (>100 chars)",
					})
				}
			}
		}
	}
	
	return issues
}