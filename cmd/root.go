package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	//"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
//var rootCmd = &cobra.Command{}

func Execute() {
	args := os.Args

	// Print other command-line arguments
	fmt.Println("Command-line arguments:")
	for i, arg := range args[0:] {
		fmt.Printf("%d: %s\n", i+1, arg)
	}
	ballerinaHome := "/path/to/ballerina/home"
	//javaPath := filepath.Join(ballerinaHome, "../../dependencies/jdk-17.0.7+7-jre")
	javaPath := filepath.Join(ballerinaHome, "..", "..", "dependencies", "jdk-17.0.7+7-jre")

	if stat, err := os.Stat(javaPath); err == nil && stat.IsDir() {
		javaHome := javaPath
		// Set JAVA_HOME environment variable
		os.Setenv("JAVA_HOME", javaHome)
	}
	fmt.Println(runtime.GOOS)
	var cygwin, mingw, os400, darwin bool
	switch runtime.GOOS {
	case "darwin":
		darwin = true
		javaHome := os.Getenv("JAVA_HOME")
		javaVersion := os.Getenv("JAVA_VERSION")
		if javaHome == "" {
			if javaVersion == "" {
				// If both JAVA_HOME and JAVA_VERSION are not set, get the default Java home
				//javaMac := path.Join(usr, libexec, javaHome)
				javaHomeCmd, _ := exec.Command("/usr/libexec/java_home").Output()
				javaHome = string(javaHomeCmd)
				os.Setenv("JAVA_HOME", javaHome)
			} else {
				// If JAVA_HOME is not set, but JAVA_VERSION is set, get the Java home for the specified version
				javaHomeCmd := exec.Command("/usr/libexec/java_home", "-v", javaVersion)
				javaHomeOutput, err := javaHomeCmd.Output()
				if err == nil {
					javaHome = string(javaHomeOutput)
					os.Setenv("JAVA_HOME", javaHome)
				} else {
					fmt.Printf("Error determining Java home for version %s: %v\n", javaVersion, err)
					os.Exit(1)
				}
			}
		}
	case "os400":
		os400 = true
	}

	if _, ok := os.LookupEnv("CYGWIN"); ok {
		cygwin = true // Running in Cygwin environment
	}

	if _, ok := os.LookupEnv("MINGW"); ok {
		mingw = true // Running in MinGW environment
	}

	fmt.Println("Cygwin:", cygwin)
	fmt.Println("MinGW:", mingw)
	fmt.Println("OS400:", os400)
	fmt.Println("Darwin:", darwin)

	prg, err := os.Executable()
	prg = "/usr/lib/ballerina/distributions/ballerina-2201.8.4/bin/bal"
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	// Resolve symbolic links
	for {
		link, err := filepath.EvalSymlinks(prg)
		if err != nil {
			fmt.Println("Error resolving symbolic link:", err)
			return
		}

		if link == prg {
			break
		}

		prg = link
	}

	// Get the directory of the script
	prgDir := filepath.Dir(prg)

	// Set BALLERINA_HOME
	ballerinaHome, err = filepath.Abs(filepath.Join(prgDir, ".."))
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return
	}

	// Output the result
	fmt.Println("BALLERINA_HOME:", ballerinaHome)

	//Setting Java CMD

	javaCmd := os.Getenv("JAVACMD")
	javaHome := os.Getenv("JAVA_HOME")
	//fmt.Println("JavaCmD= ", javaCmd)

	if javaCmd == "" {
		if javaHome != "" {
			// If JAVA_HOME is set, determine the appropriate JAVACMD
			var javacmdPath string

			// Check if the specified Java home is for IBM's JDK on AIX
			if _, err := os.Stat(filepath.Join(javaHome, "jre", "sh", "java")); err == nil {
				javacmdPath = filepath.Join(javaHome, "jre", "sh", "java")
			} else {
				// Use the standard Java executable path
				javacmdPath = filepath.Join(javaHome, "bin", "java")
			}

			// Check if the determined JAVACMD path is executable
			if _, err := os.Stat(javacmdPath); err == nil {
				javaCmd = javacmdPath
			} else {
				fmt.Printf("Error: Unable to find executable JAVACMD in specified JAVA_HOME: %s\n", javaHome)
				os.Exit(1)
			}
		} else {
			// If neither JAVACMD nor JAVA_HOME is set, use the default 'java' command
			javaCmd = "java"
		}
	}

	javaCmd = strings.TrimSpace(javaCmd)

	//fmt.Printf("JAVACMD: %s\n", javaCmd)
	//fmt.Printf("JAVAHOME: %s\n", javaHome)

	//assign value to ballerinaCliWidth
	var ballerinaCLIWidth string

	// Check if tput is available
	if !commandExists("tput") {
		ballerinaCLIWidth = "80"
	} else {
		if cols, err := getTerminalColumns(); err != nil {
			ballerinaCLIWidth = "80"
		} else {
			ballerinaCLIWidth = strconv.Itoa(cols)
		}
	}

	// Export the BALLERINA_CLI_WIDTH environment variable
	//os.Setenv("BALLERINA_CLI_WIDTH", ballerinaCLIWidth)

	// Print the result for demonstration
	fmt.Println("BALLERINA_CLI_WIDTH:", ballerinaCLIWidth)
	//Setting Ballerina debug port

	balJavaDebug := os.Getenv("BAL_JAVA_DEBUG")
	javaOpts := ""
	if balJavaDebug != "" {
		fmt.Println("BAL_JAVA_DEBUG is set")

		if balJavaDebug == "" {
			fmt.Println("Please specify the debug port for the BAL_JAVA_DEBUG variable")
			os.Exit(1)
		} else {
			javaOpts = os.Getenv("JAVA_OPTS")

			if javaOpts != "" {
				fmt.Println("Warning !!! User specified JAVA_OPTS may interfere with BAL_JAVA_DEBUG")
			}

			balDebugOpts := os.Getenv("BAL_DEBUG_OPTS")

			if balDebugOpts != "" {
				javaOpts = javaOpts + " " + balDebugOpts
			} else {
				javaOpts = javaOpts + " -Xdebug -Xnoagent -Djava.compiler=NONE -Xrunjdwp:transport=dt_socket,server=y,suspend=y,address=" + balJavaDebug
			}

			fmt.Println("Java Options:", javaOpts)
		}
	} else {
		fmt.Println("BAL_JAVA_DEBUG is not set")
	}
	//Ballerina jarpath setting
	//ballerinaHome = "/usr/lib/ballerina/distributions/ballerina-2201.8.4"
	jarPath := filepath.Join(ballerinaHome, "bre", "lib")
	var ballerinaJarPath string

	// Get a list of all JAR files in the jarPath directory
	files, err := filepath.Glob(filepath.Join(jarPath, "*.jar"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	pattern := filepath.Join(ballerinaHome, "bre", "lib", "*.jar")

	// Iterate over the JAR files
	for _, f := range files {
		matched, err := filepath.Match(pattern, f)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if !matched {
			ballerinaJarPath += ":" + f
		}
	}

	//fmt.Printf("BALLERINA_JARPATH: %s\n", ballerinaJarPath)

	//Ballerina Classpath Setting
	var ballerinaClasspath string
	toolsJarPath := filepath.Join(ballerinaHome, "bre", "lib", "tools.jar")
	if _, err := os.Stat(toolsJarPath); err == nil {
		ballerinaClasspath = filepath.Join(javaHome, "lib", "tools.jar")
	}
	// Add JAR files from bre/lib
	//brePath := filepath.Join(ballerinaHome, "bre", "lib")
	breFiles, err := filepath.Glob(filepath.Join(jarPath, "*.jar"))
	if err == nil {
		//fmt.Println("Files in bre folder")
		for _, f := range breFiles {
			//fmt.Println(f)
			ballerinaClasspath += string(filepath.ListSeparator) + f
		}
	}

	// Add JAR files from lib/tools/lang-server/lib
	langServerPath := filepath.Join(ballerinaHome, "lib", "tools", "lang-server", "lib")
	langServerFiles, err := filepath.Glob(filepath.Join(langServerPath, "*.jar"))
	if err == nil {
		for _, f := range langServerFiles {
			ballerinaClasspath += string(filepath.ListSeparator) + f
		}
	}

	// Add JAR files from lib/tools/debug-adapter/lib
	debugAdapterPath := filepath.Join(ballerinaHome, "lib", "tools", "debug-adapter", "lib")
	debugAdapterFiles, err := filepath.Glob(filepath.Join(debugAdapterPath, "*.jar"))
	if err == nil {
		for _, f := range debugAdapterFiles {
			ballerinaClasspath += string(filepath.ListSeparator) + f
		}
	}

	//Define Ballerina CLI Arguments
	cmdLineArgs := []string{
		"-Xms256m", "-Xmx1024m",
		"-XX:+HeapDumpOnOutOfMemoryError",
		"-classpath", ballerinaClasspath,
		"-Dballerina.home=" + ballerinaHome,
		"-Dballerina.version=2201.8.4",
		"-Denable.nonblocking=false",
		"-Djava.security.egd=file:/dev/./urandom",
		"-Dfile.encoding=UTF8",
		"-Dballerina.target=jvm",
		"-Djava.command=" + javaCmd,
	}
	fmt.Println("length", len(os.Args))
	if javaOpts != "" {
		JAVA_OPTS := strings.Fields(javaOpts)
		cmdLineArgs = append(cmdLineArgs, JAVA_OPTS...)
	}
	//Implementation of bal run <*.jar>
	if len(os.Args) >= 3 && os.Args[1] == "run" && isJarFile(os.Args[2]) {
		//jarFilePath := os.Args[2]
		cmdArgs := append(cmdLineArgs, "-jar")
		cmdArgs = append(cmdArgs, os.Args[2:]...)

		// Execute the command
		cmd := exec.Command(javaCmd, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err := cmd.Run()
		if err != nil {
			fmt.Println("Error running bal run jar command:", err)
			os.Exit(1)
		}
	} else if len(os.Args) == 4 && os.Args[1] == "run" && isJarFile(os.Args[3]) && os.Args[2][:8] == "--debug=" {
		debugPort, err := extractDebugPort(os.Args[2]) //Implementation of bal run --debug=<PORT> <*.jar>
		if err != nil {
			fmt.Println("Error: Invalid debug port number specified.")
			os.Exit(1)
		}
		fmt.Println(debugPort)
		if validateDebugPort(debugPort) {

			cmdArgs := append(cmdLineArgs,
				"-agentlib:jdwp=transport=dt_socket,server=y,suspend=y,address="+strconv.Itoa(debugPort),
				"-jar",
				os.Args[3],
			)
			cmd := exec.Command(javaCmd, cmdLineArgs...)
			cmd.Args = append(cmd.Args, cmdArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Println("Error running else if 1 command:", err)
				os.Exit(1)
			}
		}
	} else if len(os.Args) == 5 && os.Args[1] == "run" && os.Args[2] == "--debug" && isJarFile(os.Args[4]) {
		debugPort, err := strconv.Atoi(os.Args[3]) // handles "bal run --debug <PORT> <JAR_PATH>" command
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if validateDebugPort(debugPort) {
			cmdArgs := append(cmdLineArgs,
				"-agentlib:jdwp=transport=dt_socket,server=y,suspend=y,address="+strconv.Itoa(debugPort),
				"-jar")
			cmdArgs = append(cmdArgs, os.Args[4:]...)
			cmd := exec.Command(javaCmd, cmdArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Println("Error running else if 2 command:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Error: Invalid debug port number specified.")
			os.Exit(1)
		}
	} else {

		cmdArgs := append(cmdLineArgs, "io.ballerina.cli.launcher.Main")
		cmdArgs = append(cmdArgs, os.Args[1:]...)
		// Execute the command
		cmd := exec.Command(javaCmd, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()

		// Run the command

		if err != nil {
			fmt.Println("Error running else command:", err)
			os.Exit(1)
		}
	}
	//         /home/pabadhi/myGoProjects/sample2/app/build/libs/app.jar

}

func isJarFile(filePath string) bool {
	return len(filePath) > 4 && filePath[len(filePath)-4:] == ".jar"
}

func extractDebugPort(debugArg string) (int, error) {
	re := regexp.MustCompile(`--debug=([0-9]+)`)
	matches := re.FindStringSubmatch(debugArg)
	if len(matches) != 2 {
		return 0, fmt.Errorf("invalid --debug argument format")
	}

	return strconv.Atoi(matches[1])
}
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func getTerminalColumns() (int, error) {
	cmd := exec.Command("tput", "cols")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	cols, err := strconv.Atoi(string(output))
	if err != nil {
		return 0, err
	}

	return cols, nil
}

func validateDebugPort(port int) bool {
	return port > 0
}

func init() {
}
