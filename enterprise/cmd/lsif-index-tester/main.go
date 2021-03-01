package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
)

// TODO:
//   We need to check for lsif-clang, lsif-validate at start time.

type ProjectResult struct {
	name       string
	success    bool
	usage      UsageStats
	output     string
	testResult TestSuiteResult
}

type IndexerResult struct {
	usage  UsageStats
	output string
}

type UsageStats struct {
	// Memory usage in kilobytes by child process.
	memory int64
}

type PassedTest struct {
	Name string
}

type FailedTest struct {
	Name string
	Diff string
}

type TestFileResult struct {
	Name   string
	Passed []PassedTest
	Failed []FailedTest
}

type TestSuiteResult struct {
	FileResults []TestFileResult
}

var directory string
var raw_indexer string
var debug bool

// TODO: Do more monitoring of the process.
// var monitor bool

func logFatal(msg string, args ...interface{}) {
	log15.Error(msg, args...)
	os.Exit(1)
}

func main() {
	flag.StringVar(&directory, "dir", ".", "The directory to run the test harness over")
	flag.StringVar(&raw_indexer, "indexer", "", "The name of the indexer that you want to test")
	flag.BoolVar(&debug, "debug", false, "Enable debugging")

	flag.Parse()

	if debug {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlDebug, log15.StdoutHandler))
	} else {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.StdoutHandler))
	}

	if raw_indexer == "" {
		logFatal("Indexer is required. Pass with --indexer")
	}

	var indexer = strings.Split(raw_indexer, " ")

	log15.Info("Starting Execution: ", "base directory", directory, "indexer", raw_indexer)

	testContext := context.Background()
	err := testDirectory(testContext, indexer, directory)
	if err != nil {
		logFatal("Failed with", "err", err)
	}
}

func testDirectory(ctx context.Context, indexer []string, directory string) error {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	type ChannelResult struct {
		result ProjectResult
		err    error
	}

	resultChan := make(chan ChannelResult, len(files))
	var wg sync.WaitGroup

	for _, f := range files {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			projResult, err := testProject(ctx, indexer, directory+"/"+name, name)

			resultChan <- ChannelResult{
				result: projResult,
				err:    err,
			}
		}(f.Name())

	}

	wg.Wait()
	close(resultChan)

	allPassed := true
	for chanResult := range resultChan {
		if chanResult.err != nil {
			log15.Warn("Failed to run test. Got err:", "error", chanResult.err)
			continue
		}

		for _, fileResult := range chanResult.result.testResult.FileResults {
			if len(fileResult.Failed) > 0 {
				allPassed = false
				for _, failed := range fileResult.Failed {
					fmt.Printf("Failed test: File: %s, Name: %s\nDiff: %s\n", fileResult.Name, failed.Name, failed.Diff)
				}
			}
		}
	}

	if !allPassed {
		return errors.New("Failed some tests. Try again later :)")
	}

	return nil
}

func testProject(ctx context.Context, indexer []string, project, name string) (ProjectResult, error) {
	output, err := setupProject(project)
	if err != nil {
		return ProjectResult{name: name, success: false, output: string(output)}, err
	} else {
		log15.Debug("... Generated compile_commands.json")
	}

	result, err := runIndexer(ctx, indexer, project, name)
	if err != nil {
		return ProjectResult{
			name:    name,
			success: false,
			output:  string(result.output),
		}, err
	}

	log15.Debug("... \t Resource Usage:", "usage", result.usage)

	output, err = validateDump(project)
	if err != nil {
		fmt.Println("Not valid")
		return ProjectResult{
			name:    name,
			success: false,
			usage:   result.usage,
			output:  string(output),
		}, err
	} else {
		log15.Debug("... Validated dump.lsif")
	}

	bundle, err := readBundle(1, project)
	if err != nil {
		return ProjectResult{
			name:    name,
			success: false,
			usage:   result.usage,
			output:  string(output),
		}, err
	}

	testResult, _ := validateTestCases(project, bundle)

	return ProjectResult{
		name:       name,
		success:    true,
		usage:      result.usage,
		output:     string(output),
		testResult: testResult,
	}, nil
}

func setupProject(directory string) ([]byte, error) {
	cmd := exec.Command("./setup_indexer.sh")
	cmd.Dir = directory

	return cmd.CombinedOutput()
}

func runIndexer(ctx context.Context, indexer []string, directory, name string) (ProjectResult, error) {
	command := indexer[0]
	args := indexer[1:]

	log15.Debug("... Generating dump.lsif")
	cmd := exec.Command(command, args...)
	cmd.Dir = directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ProjectResult{}, err
	}

	sysUsage := cmd.ProcessState.SysUsage()
	mem, _ := MaxMemoryInKB(sysUsage)

	return ProjectResult{
		name:    name,
		success: false,
		usage:   UsageStats{memory: mem},
		output:  string(output),
	}, err
}

func validateDump(directory string) ([]byte, error) {
	// TODO: Eventually this should use the package, rather than the installed module
	//       but for now this will have to do.

	// validator.validate(directory + "/" + "dump.lsif")
	cmd := exec.Command("lsif-validate", "dump.lsif")
	cmd.Dir = directory

	return cmd.CombinedOutput()
}

func validateTestCases(directory string, bundle *conversion.GroupedBundleDataMaps) (TestSuiteResult, error) {
	testFiles, err := os.ReadDir(filepath.Join(directory, "lsif_tests"))
	if err != nil {
		log15.Warn("No lsif test directory exists here", "directory", directory)
		return TestSuiteResult{}, err
	}

	var fileResults []TestFileResult
	for _, file := range testFiles {
		filePath := filepath.Ext(file.Name())

		if filePath != "json" {
			continue
		}

		fileResult, err := runOneTestFile(filepath.Join(directory, "lsif_tests", file.Name()), bundle)
		if err != nil {
			logFatal("Had an error while we do the test file", "err", err)
		}

		// log15.Info("Test Results:", "result", fileResult)
		fileResults = append(fileResults, fileResult)
	}

	return TestSuiteResult{FileResults: fileResults}, nil
}

func runOneTestFile(file string, bundle *conversion.GroupedBundleDataMaps) (TestFileResult, error) {
	doc, err := ioutil.ReadFile(file)
	if err != nil {
		return TestFileResult{}, errors.Wrap(err, "Failed to read file")
	}

	testCase := LsifTest{}
	if err := json.Unmarshal(doc, &testCase); err != nil {
		return TestFileResult{}, errors.Wrap(err, "Malformed JSON")
	}

	fileResult := TestFileResult{Name: file}

	for _, definitionRequest := range testCase.Definitions {
		path := definitionRequest.Request.TextDocument
		line := definitionRequest.Request.Position.Line
		character := definitionRequest.Request.Position.Character

		results, err := conversion.Query(bundle, path, line, character)

		if err != nil {
			return TestFileResult{}, err
		}

		// TODO: We probably can have more than one result and have that make sense...
		//		should allow testing that
		if len(results) > 1 {
			logFatal("Had too many results", "results", results)
		} else if len(results) == 0 {
			logFatal("Found no results", "results", results)
		}

		definitions := results[0].Definitions

		if len(definitions) > 1 {
			logFatal("Had too many definitions", "definitions", definitions)
		} else if len(definitions) == 0 {
			logFatal("Found no definitions", "definitions", definitions)
		}

		response := transformLocationToResponse(definitions[0])
		if diff := cmp.Diff(response, definitionRequest.Response); diff != "" {
			fileResult.Failed = append(fileResult.Failed, FailedTest{
				Name: definitionRequest.Name,
				Diff: diff,
			})
		} else {
			fileResult.Passed = append(fileResult.Passed, PassedTest{
				Name: definitionRequest.Name,
			})
		}
	}

	return fileResult, nil
}

func transformLocationToResponse(location semantic.LocationData) DefinitionResponse {
	return DefinitionResponse{
		TextDocument: location.URI,
		Range: Range{
			Start: Position{
				Line:      location.StartLine,
				Character: location.StartCharacter,
			},
			End: Position{
				Line:      location.EndLine,
				Character: location.EndCharacter,
			},
		},
	}

}
