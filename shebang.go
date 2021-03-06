package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
	var tempFifoName, tempFileName string

	aliases := getAliases(dir)
	command = aliasReplace(command, aliases)

	// If #! needs removed from the target file, copy it to a temp file without the #!
	if strings.Contains(command, "!~") || strings.Contains(command, "!+") {
		var err error
		// Read the file
		fileContents, err := ioutil.ReadFile(fullpath)
		if err != nil {
			log.Fatal(err)
		}

		var i int = 0
		// If there's no #! to remove, don't bother.
		if len(fileContents) >= 2 && string(fileContents[:2]) == "#!" {
			// Read through the file until we reach the end if the first line
			for i = 0; i < len(fileContents) && fileContents[i] != '\n'; i++ {
			}
		}

		// fifo
		if strings.Contains(command, "!~") {
			tempFifoName = getTempFilename(targetName, targetExtension)
			err = syscall.Mkfifo(tempFifoName, uint32(os.ModePerm|os.ModeNamedPipe))
			// err := syscall.Mkfifo(tempFifoName, uint32(os.ModePerm))
			if err != nil {
				log.Println("Unable to create named pipe (fifo)")
				log.Fatal(err)
			}
			defer os.Remove(tempFifoName)

			// Write everything after the #! line to the new file (the first line should
			// still be there, but empty. This is nice for debugging, since line numbers
			// all matchup between the target file and the temp file)
			// tempFifoName = tempfile.Name() + targetExtension
			go writeToFifo(tempFifoName, fileContents[i:])
			defer os.Remove(tempFifoName)
		}

		// temp file
		if strings.Contains(command, "!+") {
			tempFileName = getTempFilename(targetName, targetExtension)

			// Don't write the file in a new thread, since we want the whole file
			// to be there before compilers/whatever try to get at it
			writeToFile(tempFileName, fileContents[i:])
			defer os.Remove(tempFileName)
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
		"!~", tempFifoName,
		"!+", tempFileName)
	// Do the replacement
	command = replacer.Replace(command)

	if *verbose {
		fmt.Println(command)
	}

	// The command needs wrapped in quotes for sh -c to use properly
	command = "\"" + command + "\""

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
		"You can use the following substitution meta-strings in command:",
		"\t!(aname)\tcommand defined by aname in a .shebang_alias file",
		"\t!@\ttarget",
		"\t!-\ttarget with the file extension removed",
		"\t!<\ttarget's director",
		"\t!>\ttarget's filename",
		"\t!.\ttarget's path",
		"\t!~\ttemporary copy of target, with #! line removed. The copy is a fifo",
		"\t\tNote: shebang will only write to the fifo once, with the contents then",
		"\t\tbeing discarded when read. So don't use !~ more than once in a command,",
		"\t\tunless you really know what you're doing. Even then, you should probably",
		"\t\tjust make your own fifo.",
		"\t\tBut, I guess if you really want to, you can do things neat/hacky like",
		"\t\t\"gcc -o !- !~; echo 1$'\\n'2$'\\n' > !~ & ./!- < !~\"",
		"\t\tto send some input to stdin",
		"\t\tI can't really think of any good reasons to do it this way, rather than",
		"\t\twith regular pipes, but it's totally possible.",
		"\t!+\tsame as !~, except using a temporary file, instead of a fifo, for",
		"\t\tprograms that don't play well with fifos (here's looking at you, ",
		"\t\tScheme...)",

		"Aliases",
		"\tShebang aliases, defined in .shebang_alias files, map names to commands.",
		"\tAlias commands may contain substitution meta-strings, including other ",
		"\tsubsequently defined aliases. Recursive aliases are not supported.",
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
			// Is this line commented out? Skip it.
			if byte(scanner.Text()[0]) == '#' {
				continue
			}
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

func getTempFilename(basename string, extension string) string {
	r := rand.New(rand.NewSource(31415926535)) // Doesn't need to be particularly random, so a const seed works
	for {
		nameArr := []string{basename, strconv.FormatInt(int64(r.Int()), 10), extension}
		var name string = strings.Join(nameArr, "")
		if _, err := os.Stat(name); os.IsNotExist(err) {
			return name
		}
	}
}

func writeToFifo(filename string, data []byte) {
	err := ioutil.WriteFile(filename, data, os.ModePerm|os.ModeNamedPipe)
	if err != nil {
		log.Fatal(err)
	}
}

func writeToFile(filename string, data []byte) {
	err := ioutil.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}
