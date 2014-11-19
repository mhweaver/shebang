package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	// Set up flags
	verbose := flag.Bool("d", false, "Display command being run")
	removeShebang := flag.Bool("r", false, "Remove the #! line from the target file")
	flag.Parse()

	// Adjust the start index of the command argument in os.Args, if flags were set
	var commandStart int = 1
	if *verbose {
		commandStart++
	}
	if *removeShebang {
		commandStart++
	}

	// #! tends to mess up quoted arguments, so deal with that
	var command, target string = parseArgs(os.Args[commandStart:])

	dir, filename := filepath.Split(target)
	fullpath := dir + filename
	targetName := removeExtension(fullpath)

	// If #! needs removed from the target file, copy it to a temp file without the #!
	if *removeShebang {
		var err error
		// Read the file
		fileContents, err := ioutil.ReadFile(fullpath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to read target file")
			os.Exit(1)
		}

		// If there's no #! to remove, don't bother.
		if len(fileContents) >= 2 && string(fileContents[:2]) == "#!" {
			var i int

			// io/ioutils has a nice TempFile fuctions, so we might as well use it
			// (really, I'm just using it for the name, because I'm lazy...)
			tempfile, err := ioutil.TempFile(dir, filename)
			defer os.Remove(tempfile.Name())

			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to create temp file")
				os.Exit(1)
			}

			// Read through the file until we reach the end if the first line
			for i = 0; i < len(fileContents) && fileContents[i] != '\n'; i++ {
			}

			// Write everything after the #! line to the new file (the first line should
			// still be there, but empty. This is nice for debugging, since line numbers
			// all matchup between the target file and the temp file)
			err = ioutil.WriteFile(tempfile.Name()+".c", fileContents[i:], os.ModePerm)
			defer os.Remove(tempfile.Name() + ".c")
			if err != nil {
				fmt.Println(tempfile.Name() + ".c")
				fmt.Fprintln(os.Stderr, "Unable to write to temp file")
				os.Exit(1)
			}

			// Update the variables, so they get substituted correctly
			target = tempfile.Name() + ".c"
			dir, filename = filepath.Split(target)
		}

	}

	// targetName := removeExtension(filename)
	fullpath = dir + filename

	// List of meta-strings and what they are replaced with
	replacer := strings.NewReplacer(
		"!@", target,
		"!-", targetName,
		"!>", filename,
		"!<", dir,
		"!.", fullpath)
	// Do the replacement
	command = replacer.Replace(command)

	// The command needs wrapped in quotes for sh -c to use properly
	command = "\"" + command + "\""

	if *verbose {
		fmt.Println(command)
	}

	// Run sh -c command, using stdio
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func removeExtension(filename string) string {
	extension := filepath.Ext(filename)
	return strings.TrimSuffix(filename, extension)
}

// Under certain conditions, the quoted command arg gets split up weirdly.
// This turns the args into a string, then pulls it back apart
// Returns command, target - command is all but the last space-separated token
//						   - target is the last space-separated token
func parseArgs(argv []string) (command, target string) {
	argStr := strings.Join(argv, " ")
	splitArgs := strings.Split(argStr, " ")
	numToks := len(splitArgs)

	command = strings.Join(splitArgs[:numToks-1], " ")
	target = splitArgs[numToks-1]

	return command, target

}

func printUsage() {
	fmt.Println("Usage:\tshebang [-d] [-r] \"command\" target\n")
	fmt.Println("Optional Parameters:")
	fmt.Println("\t-d\tDisplay the final command getting run")
	fmt.Println("\t-r\tRemove the #! line from the target file\n")
	fmt.Println("You can use the following substitution meta-strings in command:")
	fmt.Println("\t!@\ttarget")
	fmt.Println("\t!-\ttarget with the file extension removed")
	fmt.Println("\t!<\ttarget's filename")
	fmt.Println("\t!>\ttarget's directory")
	fmt.Println("\t!.\ttarget's path")
}
