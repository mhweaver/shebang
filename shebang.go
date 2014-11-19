package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type alias struct {
	name    string
	command string
}

func main() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	// Set up flags
	verbose := flag.Bool("d", false, "Display command being run")
	flag.Parse()

	// Adjust the start index of the command argument in os.Args, if flags were set
	var commandStart int = 1
	if *verbose {
		commandStart++
	}

	// #! tends to mess up quoted arguments, so deal with that
	var command, target string = parseArgs(os.Args[commandStart:])

	dir, filename := filepath.Split(target)
	fullpath := dir + filename
	targetName := removeExtension(fullpath)
	targetExtension := filepath.Ext(fullpath)
	var tempTargetName string

	aliases := getAliases(dir)
	command = aliasReplace(command, aliases)

	// If #! needs removed from the target file, copy it to a temp file without the #!
	if strings.Contains(command, "!~") {
		var err error
		// Read the file
		fileContents, err := ioutil.ReadFile(fullpath)
		if err != nil {
			log.Fatal(err)
		}

		// If there's no #! to remove, don't bother.
		if len(fileContents) >= 2 && string(fileContents[:2]) == "#!" {
			var i int

			// io/ioutils has a nice TempFile fuctions, so we might as well use it
			// (really, I'm just using it for the name, because I'm lazy...)
			tempfile, err := ioutil.TempFile(dir, filename)
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(tempfile.Name())

			// Read through the file until we reach the end if the first line
			for i = 0; i < len(fileContents) && fileContents[i] != '\n'; i++ {
			}

			// Write everything after the #! line to the new file (the first line should
			// still be there, but empty. This is nice for debugging, since line numbers
			// all matchup between the target file and the temp file)
			tempTargetName = tempfile.Name() + targetExtension
			err = ioutil.WriteFile(tempTargetName, fileContents[i:], os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(tempTargetName)
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
		"!.", fullpath,
		"!~", tempTargetName)
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
	usage := [...]string{
		//																				 |
		"Usage:\tshebang [-r] \"command\" target\n",
		"Optional Parameters",
		"\t-d\tDisplay the final command getting run",
		"\t\tNote: this may (and probably will) contain extra quotes. Ignore those.",
		"You can use the following substitution meta-strings in command:",
		"\t!(aname)\tcommand defined by aname in a .shebang_alias file",
		"\t!@\ttarget",
		"\t!-\ttarget with the file extension removed",
		"\t!<\ttarget's filename",
		"\t!>\ttarget's directory",
		"\t!.\ttarget's path",
		"\t!~\ttemporary copy of target, with #! line removed",

		"Aliases",
		"\tShebang aliases, defined in .shebang_alias files, map names to commands.",
		"\tAlias commands may contain substitution meta-strings, including other ",
		"\tpreviously defined aliases. Recursive aliases are not supported.",
		"\tAliases can be defined in the following locations, and will be parsed in",
		"\tthis order:",
		"\t\t<target's directory>/.shebang_alias",
		"\t\t./.shebang_alias",
		"\t\t~/.shebang_alias",
		"\t\t/etc/.shebang_alias",
		"\t.shebang_alias files contain 1 alias per line, in the following format:",
		"\t\talias_name = command",
		"\tFor example, to run/compile a C program you might use the following alias:",
		"\t\tc = gcc -o !- !~ && !-; rm !-",
		"\tThis creates a temporary copy of the target without the #! line (!~), ",
		"\tcompiles it, runs the output file (!-), then removes the output file.",
		"\tTo use this, the #! in the .c file would look something like this:",
		"\t#!/usr/bin/shebang \"!(c)\"",
	}
	for _, line := range usage {
		fmt.Println(line)
	}
}

func getAliases(targetDir string) []alias {
	aliasLocations := [...]string{targetDir, "./", "~/", "/etc/"}
	aliasFilename := ".shebang_alias"
	// aliases := list.New()
	aliases := []alias{}

	for i := 0; i < len(aliasLocations); i++ {
		filename := aliasLocations[i] + aliasFilename

		// Open the file
		f, err := os.Open(filename)
		if err != nil {
			// Ignore the error and move on to the next file location
		} else {
			defer f.Close()
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			al := parseAlias(scanner.Text())
			// aliases.PushBack(al)
			aliases = append(aliases, al)
		}
	}

	return aliases
}

// name = command
func parseAlias(raw string) alias {
	tokens := strings.SplitN(raw, "=", 2) // Split into "name " and " command"
	for index, _ := range tokens {
		tokens[index] = strings.Trim(tokens[index], " \t")
	}
	return alias{name: tokens[0], command: tokens[1]}
}

func aliasReplace(command string, aliases []alias) string {
	cmd := command
	for _, al := range aliases {
		name := al.name
		command := al.command
		cmd = strings.Replace(cmd, "!("+name+")", command, -1)
	}
	return cmd
}
