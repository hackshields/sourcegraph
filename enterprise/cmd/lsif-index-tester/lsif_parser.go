package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/pathexistence"
)

func readBundle(dumpID int, root string) (*conversion.GroupedBundleDataMaps, error) {
	dumpPath := path.Join(root, "dump.lsif")
	getChildrenFunc := makeExistenceFunc(root)
	file, err := os.Open(dumpPath)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't open file")
	}
	defer file.Close()

	bundle, err := conversion.Correlate(context.Background(), file, dumpID, "", getChildrenFunc)
	if err != nil {
		fmt.Println("conversion failed")
		return nil, err
	}

	return conversion.GroupedBundleDataChansToMaps(context.Background(), bundle), nil
}

func makeExistenceFunc(directory string) pathexistence.GetChildrenFunc {
	return func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		// NOTE: We're using find because it allows us to look for things outside of git directories (you might just have a bunch
		// of code somewhere, outside of a git repo). But if you're experiencing big slowdowns, you may want to try and consider
		// making this a bit smarter to prune away things (like in node_modules for example... or even just using git ls-files like
		// in repl/main.go)

		cmd := exec.Command("bash", "-c", fmt.Sprintf("find %s -maxdepth 1", strings.Join(dirnames, " ")))
		cmd.Dir = directory

		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("find(%v) failed: %v", strings.Join(dirnames, ", "), string(out)))
		}

		return parseDirectoryChildren(dirnames, strings.Split(string(out), "\n")), nil
	}
}

func parseDirectoryChildren(dirnames, paths []string) map[string][]string {
	childrenMap := map[string][]string{}

	// Ensure each directory has an entry, even if it has no children
	// listed in the gitserver output.
	for _, dirname := range dirnames {
		childrenMap[dirname] = nil
	}

	// Order directory names by length (biggest first) so that we assign
	// paths to the most specific enclosing directory in the following loop.
	sort.Slice(dirnames, func(i, j int) bool {
		return len(dirnames[i]) > len(dirnames[j])
	})

	for _, path := range paths {
		if strings.Contains(path, "/") {
			for _, dirname := range dirnames {
				if strings.HasPrefix(path, dirname) {
					childrenMap[dirname] = append(childrenMap[dirname], path)
					break
				}
			}
		} else {
			// No need to loop here. If we have a root input directory it
			// will necessarily be the last element due to the previous
			// sorting step.
			if len(dirnames) > 0 && dirnames[len(dirnames)-1] == "" {
				childrenMap[""] = append(childrenMap[""], path)
			}
		}
	}

	return childrenMap
}
