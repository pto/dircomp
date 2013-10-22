package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

var (
	commandName = filepath.Base(os.Args[0])
	allFlag     = flag.Bool("all", false, "show all files, not just changes")
	debugFlag   = flag.Bool("debug", false, "print debugging information")
	helpFlag    = flag.Bool("help", false, "print this help message")
)

func main() {
	flag.Parse()
	if *helpFlag || len(os.Args) != flag.NFlag()+3 {
		usage()
		os.Exit(1)
	}

	sourceDir := os.Args[flag.NFlag()+1]
	destDir := os.Args[flag.NFlag()+2]

	// Command line arguments must be directories
	sourceFileInfo, err := os.Stat(sourceDir)
	if err == nil && !sourceFileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "%s: source %s is not a directory\n",
			commandName, sourceDir)
		os.Exit(1)
	}
	destFileInfo, err := os.Stat(destDir)
	if err == nil && !destFileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "%s: destination %s is not a directory\n",
			commandName, destDir)
		os.Exit(1)
	}

	// Read directories
	sourceInfo, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot read source %s\n",
			commandName, sourceDir)
		os.Exit(1)
	}

	destInfo, err := ioutil.ReadDir(destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot read destination %s\n",
			commandName, destDir)
		os.Exit(1)
	}

	compareDirs(sourceDir, sourceInfo, destDir, destInfo)
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [flags] sourceDirectory destinationDirectory\n\nCompare source directory tree to destination, listing differences\nin destination, based on an MD5 hash comparison.\n\nOutput indicators:\n\n+ file is added in destination\n- file is removed in destination\nM file is modified in destination\n  file is identical (only with -all flag)\n\n",
		commandName)
	flag.PrintDefaults()
}

type ByName []os.FileInfo

func (fi ByName) Len() int           { return len(fi) }
func (fi ByName) Swap(i, j int)      { fi[i], fi[j] = fi[j], fi[i] }
func (fi ByName) Less(i, j int) bool { return fi[i].Name() < fi[j].Name() }

func compareDirs(sourceBase string, sourceInfo []os.FileInfo,
	destBase string, destInfo []os.FileInfo) {

	if *debugFlag {
		fmt.Printf("debug: comparing directories \"%s\" and \"%s\"\n", sourceBase, destBase)
	}

	sort.Sort(ByName(sourceInfo))
	sort.Sort(ByName(destInfo))
	s, d := 0, 0

	for s < len(sourceInfo) || d < len(destInfo) {
		// Get names
		var sourceName, destName, fullSourceName, fullDestName string
		if s < len(sourceInfo) {
			sourceName = sourceInfo[s].Name()
			fullSourceName = filepath.Join(sourceBase, sourceName)
		}
		if d < len(destInfo) {
			destName = destInfo[d].Name()
			fullDestName = filepath.Join(destBase, destName)
		}

		if *debugFlag {
			fmt.Printf("debug: comparing names \"%s\" and \"%s\"\n",
				sourceName, destName)
		}

		// Names don't match
		if s == len(sourceInfo) || sourceName > destName {
			fmt.Printf("+ %s", fullDestName)
			if destInfo[d].IsDir() {
				fmt.Print(" (directory)")
			}
			fmt.Println()
			d++
			continue
		}
		if d == len(destInfo) || sourceName < destName {
			fmt.Printf("- %s", filepath.Join(destBase, sourceName))
			if sourceInfo[s].IsDir() {
				fmt.Print(" (directory)")
			}
			fmt.Println()
			s++
			continue
		}

		// IsDir() doesn't match
		if sourceInfo[s].IsDir() && !destInfo[d].IsDir() {
			fmt.Printf("C %s (not a directory in destination)\n",
				fullDestName)
			s++
			d++
			continue
		}
		if !sourceInfo[s].IsDir() && destInfo[d].IsDir() {
			fmt.Printf("C %s (not a directory in source)\n",
				fullDestName)
			s++
			d++
			continue
		}

		// Process a subdirectory
		if sourceInfo[s].IsDir() {
			sourceSubdirInfo, err := ioutil.ReadDir(fullSourceName)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"%s: cannot read source directory %s\n",
					commandName, fullSourceName)
				s++
				d++
				continue
			}
			destSubdirInfo, err := ioutil.ReadDir(fullDestName)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"%s: cannot read destination directory %s\n",
					commandName, fullDestName)
				s++
				d++
				continue
			}
			compareDirs(fullSourceName, sourceSubdirInfo,
				fullDestName, destSubdirInfo)
			s++
			d++
			continue
		}

		// Process a regular file
		sourceHash, err := md5sum(fullSourceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", commandName, err)
			s++
			d++
			continue
		}
		destHash, err := md5sum(fullDestName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", commandName, err)
			s++
			d++
			continue
		}
		if sourceHash != destHash {
			fmt.Printf("C %s\n", fullDestName)
		} else if *allFlag {
			fmt.Printf("  %s\n", fullDestName)
		}
		s++
		d++
		continue
	}
}

func md5sum(filename string) (hash [md5.Size]byte, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return hash, fmt.Errorf("cannot open file %s", filename)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(bufio.NewReader(file))
	if err != nil {
		return hash, fmt.Errorf("cannot read file %s", filename)
	}
	return md5.Sum(data), nil
}
