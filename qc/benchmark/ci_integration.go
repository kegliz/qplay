// Package benchmark provides CI/CD integration utilities for automated benchmarking
package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kegliz/qplay/qc/testutil"
)

// CIConfig holds configuration for CI/CD environments
type CIConfig struct {
	Environment    string            `json:"environment"` // "github", "gitlab", "jenkins", etc.
	Branch         string            `json:"branch"`
	CommitHash     string            `json:"commit_hash"`
	CommitMessage  string            `json:"commit_message"`
	PullRequest    string            `json:"pull_request,omitempty"`
	BuildNumber    string            `json:"build_number"`
	Metadata       map[string]string `json:"metadata"`
	ResourceLimits ResourceLimits    `json:"resource_limits"`
}

// CIBenchmarkRunner provides CI/CD-friendly benchmark execution
type CIBenchmarkRunner struct {
	Config      CIConfig
	Persistence *BenchmarkPersistence
	OutputDir   string
	Verbose     bool
}

// NewCIBenchmarkRunner creates a new CI benchmark runner
func NewCIBenchmarkRunner(outputDir string) *CIBenchmarkRunner {
	config := DetectCIEnvironment()

	return &CIBenchmarkRunner{
		Config:      config,
		Persistence: NewBenchmarkPersistence(filepath.Join(outputDir, "benchmark-history")),
		OutputDir:   outputDir,
		Verbose:     os.Getenv("BENCHMARK_VERBOSE") == "true",
	}
}

// DetectCIEnvironment automatically detects CI/CD environment and extracts metadata
func DetectCIEnvironment() CIConfig {
	config := CIConfig{
		Metadata: make(map[string]string),
	}

	// GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		config.Environment = "github-actions"
		config.Branch = os.Getenv("GITHUB_REF_NAME")
		config.CommitHash = os.Getenv("GITHUB_SHA")
		config.PullRequest = os.Getenv("GITHUB_EVENT_NUMBER")
		config.BuildNumber = os.Getenv("GITHUB_RUN_NUMBER")
		config.Metadata["repository"] = os.Getenv("GITHUB_REPOSITORY")
		config.Metadata["workflow"] = os.Getenv("GITHUB_WORKFLOW")
		config.Metadata["actor"] = os.Getenv("GITHUB_ACTOR")
	}

	// GitLab CI
	if os.Getenv("GITLAB_CI") == "true" {
		config.Environment = "gitlab-ci"
		config.Branch = os.Getenv("CI_COMMIT_REF_NAME")
		config.CommitHash = os.Getenv("CI_COMMIT_SHA")
		config.CommitMessage = os.Getenv("CI_COMMIT_MESSAGE")
		config.BuildNumber = os.Getenv("CI_PIPELINE_ID")
		config.Metadata["project"] = os.Getenv("CI_PROJECT_PATH")
		config.Metadata["runner"] = os.Getenv("CI_RUNNER_DESCRIPTION")
	}

	// Jenkins
	if os.Getenv("JENKINS_URL") != "" {
		config.Environment = "jenkins"
		config.Branch = os.Getenv("GIT_BRANCH")
		config.CommitHash = os.Getenv("GIT_COMMIT")
		config.BuildNumber = os.Getenv("BUILD_NUMBER")
		config.Metadata["job_name"] = os.Getenv("JOB_NAME")
		config.Metadata["build_url"] = os.Getenv("BUILD_URL")
	}

	// Azure DevOps
	if os.Getenv("AZURE_HTTP_USER_AGENT") != "" || os.Getenv("TF_BUILD") == "True" {
		config.Environment = "azure-devops"
		config.Branch = os.Getenv("BUILD_SOURCEBRANCH")
		config.CommitHash = os.Getenv("BUILD_SOURCEVERSION")
		config.BuildNumber = os.Getenv("BUILD_BUILDNUMBER")
		config.Metadata["project"] = os.Getenv("SYSTEM_TEAMPROJECT")
		config.Metadata["definition"] = os.Getenv("BUILD_DEFINITIONNAME")
	}

	// Default/local environment
	if config.Environment == "" {
		config.Environment = "local"
		config.Branch = "main" // Default branch
		config.CommitHash = "unknown"
		config.BuildNumber = strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Set resource limits based on CI environment
	config.ResourceLimits = getCIResourceLimits(config.Environment)

	return config
}

// getCIResourceLimits returns appropriate resource limits for different CI environments
func getCIResourceLimits(environment string) ResourceLimits {
	switch environment {
	case "github-actions":
		return ResourceLimits{
			MaxMemoryMB:     1536, // GitHub Actions has ~1.75GB available
			MaxDuration:     5 * time.Minute,
			MaxCircuitDepth: 15,
			MaxQubits:       4,
		}
	case "gitlab-ci":
		return ResourceLimits{
			MaxMemoryMB:     1024, // Conservative for shared runners
			MaxDuration:     3 * time.Minute,
			MaxCircuitDepth: 12,
			MaxQubits:       4,
		}
	case "jenkins":
		return ResourceLimits{
			MaxMemoryMB:     2048, // Varies by Jenkins setup
			MaxDuration:     10 * time.Minute,
			MaxCircuitDepth: 20,
			MaxQubits:       5,
		}
	default:
		return DefaultResourceLimits
	}
}

// RunBenchmarkSuite runs a complete benchmark suite for CI/CD
func (ci *CIBenchmarkRunner) RunBenchmarkSuite() (*CIBenchmarkReport, error) {
	if ci.Verbose {
		fmt.Printf("🚀 Running quantum benchmark suite in %s environment\n", ci.Config.Environment)
		fmt.Printf("   Branch: %s\n", ci.Config.Branch)
		fmt.Printf("   Commit: %s\n", ci.Config.CommitHash)
		fmt.Printf("   Build: %s\n", ci.Config.BuildNumber)
	}

	// Create benchmark suite with CI-appropriate configuration
	suite := NewPluginBenchmarkSuite().
		WithLimits(ci.Config.ResourceLimits).
		WithConfig(testutil.QuickTestConfig)

	// Execute benchmarks
	report := &CIBenchmarkReport{
		Config:    ci.Config,
		Timestamp: time.Now(),
		Results:   make([]CIBenchmarkResult, 0),
	}

	for _, runner := range suite.runners {
		for _, circuit := range suite.circuits {
			for _, scenario := range suite.scenarios {
				if ci.Verbose {
					fmt.Printf("   Running: %s/%s/%s\n", runner, circuit, scenario)
				}

				result := ci.runSingleBenchmark(runner, circuit, scenario, suite)

				ciResult := CIBenchmarkResult{
					BenchmarkResult: result,
					ConfigHash:      ci.calculateConfigHash(runner, circuit, scenario),
				}

				report.Results = append(report.Results, ciResult)

				// Save to persistence layer
				if err := ci.Persistence.SaveResult(result, ci.Config.CommitHash, ci.Config.BuildNumber); err != nil {
					if ci.Verbose {
						fmt.Printf("   Warning: Failed to save result: %v\n", err)
					}
				}
			}
		}
	}

	// Perform regression analysis
	report.RegressionAnalysis = ci.analyzeRegressions(report.Results)

	// Generate artifacts
	if err := ci.generateArtifacts(report); err != nil {
		return report, fmt.Errorf("failed to generate artifacts: %v", err)
	}

	return report, nil
}

// CIBenchmarkReport contains comprehensive results for CI/CD
type CIBenchmarkReport struct {
	Config             CIConfig            `json:"config"`
	Timestamp          time.Time           `json:"timestamp"`
	Results            []CIBenchmarkResult `json:"results"`
	RegressionAnalysis *RegressionAnalysis `json:"regression_analysis"`
	Summary            CIBenchmarkSummary  `json:"summary"`
}

// CIBenchmarkResult wraps benchmark results with CI metadata
type CIBenchmarkResult struct {
	BenchmarkResult BenchmarkResult `json:"benchmark_result"`
	ConfigHash      string          `json:"config_hash"`
}

// RegressionAnalysis contains regression test results
type RegressionAnalysis struct {
	TotalComparisons int                   `json:"total_comparisons"`
	Regressions      []RegressionDetection `json:"regressions"`
	Improvements     []RegressionDetection `json:"improvements"`
	OverallStatus    string                `json:"overall_status"` // "pass", "warning", "fail"
}

// RegressionDetection represents a detected regression or improvement
type RegressionDetection struct {
	ConfigHash    string            `json:"config_hash"`
	TestName      string            `json:"test_name"`
	ChangeType    string            `json:"change_type"` // "duration", "memory", "throughput"
	ChangePercent float64           `json:"change_percent"`
	Significance  string            `json:"significance"` // "minor", "major", "critical"
	Comparison    *ComparisonResult `json:"comparison,omitempty"`
}

// CIBenchmarkSummary provides high-level summary for CI reports
type CIBenchmarkSummary struct {
	TotalTests        int           `json:"total_tests"`
	PassedTests       int           `json:"passed_tests"`
	FailedTests       int           `json:"failed_tests"`
	TotalDuration     time.Duration `json:"total_duration"`
	AverageDuration   time.Duration `json:"average_duration"`
	MemoryEfficiency  string        `json:"memory_efficiency"` // "excellent", "good", "poor"
	RecommendedAction string        `json:"recommended_action"`
}

// runSingleBenchmark executes a single benchmark in CI context
func (ci *CIBenchmarkRunner) runSingleBenchmark(runnerName string, circuit CircuitType, scenario BenchmarkScenario, suite *PluginBenchmarkSuite) BenchmarkResult {
	config := BenchmarkConfig{
		CircuitType: circuit,
		Scenario:    scenario,
		Config:      suite.config,
		RunnerName:  runnerName,
		Limits:      ci.Config.ResourceLimits,
	}

	// Create a minimal testing.B for CI context
	b := &testing.B{}
	b.ResetTimer()

	return RunSingleBenchmark(b, config)
}

// CIBenchmark implements a minimal testing.B interface for CI environments
type CIBenchmark struct {
	verbose bool
	n       int
}

func (b *CIBenchmark) ReportAllocs()                       {}
func (b *CIBenchmark) ResetTimer()                         {}
func (b *CIBenchmark) StartTimer()                         {}
func (b *CIBenchmark) StopTimer()                          {}
func (b *CIBenchmark) SetBytes(n int64)                    {}
func (b *CIBenchmark) SetParallelism(p int)                {}
func (b *CIBenchmark) ReportMetric(n float64, unit string) {}

func (b *CIBenchmark) N() int {
	if b.n == 0 {
		b.n = 1 // Default iteration count for CI
	}
	return b.n
}

func (b *CIBenchmark) Skip(args ...interface{}) {
	if b.verbose {
		fmt.Printf("   Skipped: %v\n", args)
	}
}

func (b *CIBenchmark) Errorf(format string, args ...interface{}) {
	fmt.Printf("   Error: "+format+"\n", args...)
}

func (b *CIBenchmark) Helper() {}

// calculateConfigHash creates a unique hash for benchmark configuration
func (ci *CIBenchmarkRunner) calculateConfigHash(runner string, circuit CircuitType, scenario BenchmarkScenario) string {
	return fmt.Sprintf("%s-%s-%s", runner, circuit, scenario)
}

// analyzeRegressions performs regression analysis on benchmark results
func (ci *CIBenchmarkRunner) analyzeRegressions(results []CIBenchmarkResult) *RegressionAnalysis {
	analysis := &RegressionAnalysis{
		TotalComparisons: 0,
		Regressions:      make([]RegressionDetection, 0),
		Improvements:     make([]RegressionDetection, 0),
		OverallStatus:    "pass",
	}

	for _, result := range results {
		// Skip failed benchmarks
		if !result.BenchmarkResult.Success {
			continue
		}

		// Try to compare with baseline
		comparison, err := ci.Persistence.CompareWithBaseline(
			result.BenchmarkResult,
			ci.Config.CommitHash,
			ci.Config.BuildNumber)

		if err != nil {
			continue // No baseline available
		}

		analysis.TotalComparisons++

		// Check for regressions
		if len(comparison.Regressions) > 0 {
			detection := RegressionDetection{
				ConfigHash:   result.ConfigHash,
				TestName:     GetBenchmarkName(result.BenchmarkResult.RunnerName, result.BenchmarkResult.CircuitType, result.BenchmarkResult.Scenario),
				Comparison:   comparison,
				Significance: ci.determineSeverity(comparison),
			}

			// Determine primary change type
			if comparison.Summary.DurationChange > 10 {
				detection.ChangeType = "duration"
				detection.ChangePercent = comparison.Summary.DurationChange
			} else if comparison.Summary.MemoryChange > 20 {
				detection.ChangeType = "memory"
				detection.ChangePercent = comparison.Summary.MemoryChange
			}

			analysis.Regressions = append(analysis.Regressions, detection)

			// Update overall status
			if detection.Significance == "critical" {
				analysis.OverallStatus = "fail"
			} else if detection.Significance == "major" && analysis.OverallStatus == "pass" {
				analysis.OverallStatus = "warning"
			}
		}

		// Check for improvements
		if len(comparison.Improvements) > 0 {
			detection := RegressionDetection{
				ConfigHash: result.ConfigHash,
				TestName:   GetBenchmarkName(result.BenchmarkResult.RunnerName, result.BenchmarkResult.CircuitType, result.BenchmarkResult.Scenario),
				Comparison: comparison,
			}

			if comparison.Summary.DurationChange < -10 {
				detection.ChangeType = "duration"
				detection.ChangePercent = comparison.Summary.DurationChange
			} else if comparison.Summary.MemoryChange < -20 {
				detection.ChangeType = "memory"
				detection.ChangePercent = comparison.Summary.MemoryChange
			}

			analysis.Improvements = append(analysis.Improvements, detection)
		}
	}

	return analysis
}

// determineSeverity determines the severity of a regression
func (ci *CIBenchmarkRunner) determineSeverity(comparison *ComparisonResult) string {
	// Critical: >50% duration regression or >100% memory regression
	if comparison.Summary.DurationChange > 50 || comparison.Summary.MemoryChange > 100 {
		return "critical"
	}

	// Major: >20% duration regression or >50% memory regression
	if comparison.Summary.DurationChange > 20 || comparison.Summary.MemoryChange > 50 {
		return "major"
	}

	return "minor"
}

// generateArtifacts creates CI artifacts (reports, badges, etc.)
func (ci *CIBenchmarkRunner) generateArtifacts(report *CIBenchmarkReport) error {
	// Ensure output directory exists
	if err := os.MkdirAll(ci.OutputDir, 0755); err != nil {
		return err
	}

	// Generate JSON report
	jsonPath := filepath.Join(ci.OutputDir, "benchmark-report.json")
	if err := ci.writeJSONReport(report, jsonPath); err != nil {
		return err
	}

	// Generate human-readable summary
	summaryPath := filepath.Join(ci.OutputDir, "benchmark-summary.txt")
	if err := ci.writeSummaryReport(report, summaryPath); err != nil {
		return err
	}

	// Generate badge data (for README badges)
	badgePath := filepath.Join(ci.OutputDir, "benchmark-badge.json")
	if err := ci.writeBadgeData(report, badgePath); err != nil {
		return err
	}

	return nil
}

// writeJSONReport writes the full JSON report
func (ci *CIBenchmarkRunner) writeJSONReport(report *CIBenchmarkReport, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// writeSummaryReport writes a human-readable summary
func (ci *CIBenchmarkRunner) writeSummaryReport(report *CIBenchmarkReport, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	summary := ci.generateSummary(report)

	content := fmt.Sprintf(`Quantum Benchmark Report
=======================

Environment: %s
Branch: %s
Commit: %s
Build: %s
Timestamp: %s

Results Summary:
- Total Tests: %d
- Passed: %d
- Failed: %d
- Total Duration: %v
- Average Duration: %v

Regression Analysis:
- Status: %s
- Regressions: %d
- Improvements: %d

%s
`,
		report.Config.Environment,
		report.Config.Branch,
		report.Config.CommitHash,
		report.Config.BuildNumber,
		report.Timestamp.Format(time.RFC3339),
		summary.TotalTests,
		summary.PassedTests,
		summary.FailedTests,
		summary.TotalDuration,
		summary.AverageDuration,
		report.RegressionAnalysis.OverallStatus,
		len(report.RegressionAnalysis.Regressions),
		len(report.RegressionAnalysis.Improvements),
		ci.formatDetailedResults(report))

	_, err = file.WriteString(content)
	return err
}

// generateSummary creates summary statistics
func (ci *CIBenchmarkRunner) generateSummary(report *CIBenchmarkReport) CIBenchmarkSummary {
	var totalDuration time.Duration
	passedTests := 0
	failedTests := 0

	for _, result := range report.Results {
		totalDuration += result.BenchmarkResult.Duration
		if result.BenchmarkResult.Success {
			passedTests++
		} else {
			failedTests++
		}
	}

	avgDuration := time.Duration(0)
	if len(report.Results) > 0 {
		avgDuration = totalDuration / time.Duration(len(report.Results))
	}

	return CIBenchmarkSummary{
		TotalTests:      len(report.Results),
		PassedTests:     passedTests,
		FailedTests:     failedTests,
		TotalDuration:   totalDuration,
		AverageDuration: avgDuration,
	}
}

// formatDetailedResults formats detailed results for the summary
func (ci *CIBenchmarkRunner) formatDetailedResults(report *CIBenchmarkReport) string {
	var builder strings.Builder

	if len(report.RegressionAnalysis.Regressions) > 0 {
		builder.WriteString("\nRegressions Detected:\n")
		for _, reg := range report.RegressionAnalysis.Regressions {
			builder.WriteString(fmt.Sprintf("- %s: %s %.2f%% (%s)\n",
				reg.TestName, reg.ChangeType, reg.ChangePercent, reg.Significance))
		}
	}

	if len(report.RegressionAnalysis.Improvements) > 0 {
		builder.WriteString("\nImprovements:\n")
		for _, imp := range report.RegressionAnalysis.Improvements {
			builder.WriteString(fmt.Sprintf("- %s: %s %.2f%%\n",
				imp.TestName, imp.ChangeType, imp.ChangePercent))
		}
	}

	return builder.String()
}

// writeBadgeData writes badge information for README shields
func (ci *CIBenchmarkRunner) writeBadgeData(report *CIBenchmarkReport, path string) error {
	summary := ci.generateSummary(report)

	badgeData := map[string]interface{}{
		"schemaVersion": 1,
		"label":         "benchmarks",
		"message":       fmt.Sprintf("%d/%d passed", summary.PassedTests, summary.TotalTests),
		"color":         ci.getBadgeColor(summary, report.RegressionAnalysis),
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(badgeData)
}

// getBadgeColor determines the appropriate badge color
func (ci *CIBenchmarkRunner) getBadgeColor(summary CIBenchmarkSummary, analysis *RegressionAnalysis) string {
	if summary.FailedTests > 0 || analysis.OverallStatus == "fail" {
		return "red"
	}
	if analysis.OverallStatus == "warning" {
		return "yellow"
	}
	return "green"
}
