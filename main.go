// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joshdk/retry/retry"
)

// version is used to hold the version string. Will be replaced at go build
// time with -ldflags.
var version = "development"

// cmdFlags represents the assorted command line flags that can be passed.
type cmdFlags struct {
	retry.Spec
	version bool
}

func main() {
	var flags cmdFlags
	flag.IntVar(&flags.Attempts, "attempts", 999, "maximum number of attempts (0 = unlimited)")
	flag.BoolVar(&flags.Backoff, "backoff", false, "use exponential backoff when sleeping")
	flag.IntVar(&flags.Consecutive, "consecutive", 1, "required number of back to back successes")
	flag.DurationVar(&flags.InitialDelay, "delay", 0, "initial delay period before tasks are run")
	flag.BoolVar(&flags.Invert, "invert", false, "wait for task to fail rather than succeed")
	flag.DurationVar(&flags.Jitter, "jitter", 0, "time range randomly added to sleep")
	flag.DurationVar(&flags.TotalTime, "max-time", 0, "maximum total time to run tasks (0 = unlimited)")
	flag.BoolVar(&flags.Quiet, "quiet", false, "silence all output")
	flag.DurationVar(&flags.Sleep, "sleep", 5*time.Second, "time to sleep between attempts")
	flag.DurationVar(&flags.TaskTime, "task-time", 0, "maximum time for a single attempt to take (0 = unlimited)")
	flag.BoolVar(&flags.version, "version", false, fmt.Sprintf("print the version %q and exit", version))
	flag.Usage = usage
	flag.Parse()

	if err := mainCmd(flags); err != nil {
		if !flags.Quiet {
			fmt.Fprintf(os.Stderr, "retry: %v\n", err)
		}
		os.Exit(1)
	}
}

func mainCmd(flags cmdFlags) error {
	// If the version flag (-version) was given, print the version and exit.
	if flags.version {
		fmt.Println(version)
		return nil
	}

	// If no arguments were given, there's nothing to do.
	if flag.NArg() == 0 {
		usage()
		return nil
	}

	var (
		task    retry.Task
		command = flag.Args()[0]
		args    = flag.Args()[1:]
	)

	if strings.HasPrefix(command, "http://") || strings.HasPrefix(command, "https://") {
		// The command looks like it references a url (starts with http:// or
		// https://).
		task = retry.HTTPTask{URL: command}
	} else {
		// Otherwise, assume the command references a (shell) command.
		task = retry.ExecTask{Name: command, Args: args, Quiet: flags.Quiet}
	}

	return retry.Retry(flags.Spec, task)
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: retry [flags] command|url\n")
	flag.PrintDefaults()
}
