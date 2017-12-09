package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	clog "github.com/morriswinkler/cloudglog"
	"github.com/rakyll/globalconf"
)

var (
	logFileName = flag.String("logfile", "/var/log/ssh-guard", "logfile path")
	cmds        = flag.String("allowd_commands", "", "Semi-colon separated list of allowed commads, use add and del to midify")
	addCMD      = flag.String("add", "", "add command")
	delCMD      = flag.Int("del", 0, "del command")
	listCMDS    = flag.Bool("list", false, "list all allowed comands")
)

func init() {
	clog.FormatStyle(clog.ModernFormat)
}

func cmdLineAdd(conf *globalconf.GlobalConf) {

	list := strings.Split(*cmds, ";")
	list = append(list, *addCMD)

	cFlag := flag.Lookup("allowd_commands")
	err := cFlag.Value.Set(strings.Join(list, ":"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Added command to list of allowed commands\n")
	conf.Set("", cFlag)
}

func cmdLineDel(conf *globalconf.GlobalConf) {

	list := strings.Split(*cmds, ":")
	if *delCMD >= len(list) {
		fmt.Printf("Error: can not delete element %d, the list contains only %d elements\n\n", *delCMD, len(list)-1)
		cmdLineList(conf)
		os.Exit(1)
	}

	list = append(list[:*delCMD], list[*delCMD+1:]...)

	cFlag := flag.Lookup("allowd_commands")
	err := cFlag.Value.Set(strings.Join(list, ":"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Removed command from list of allowed commands\n")
	conf.Set("", cFlag)
}

func cmdLineList(conf *globalconf.GlobalConf) {

	list := strings.Split(*cmds, ":")
	for i := 1; i < len(list); i++ {
		fmt.Printf("[%d] %s\n", i, list[i])
	}
	fmt.Printf("\n")

}

func cmdLine(conf *globalconf.GlobalConf) {

	if *addCMD != "" {
		cmdLineAdd(conf)
		cmdLineList(conf)
		os.Exit(0)
	}

	if *delCMD != 0 {
		cmdLineDel(conf)
		cmdLineList(conf)
		os.Exit(0)
	}
	if *listCMDS {
		cmdLineList(conf)
		os.Exit(0)
	}

	flag.Usage()
}

func main() {

	conf, err := globalconf.New("ssh-guard")
	if err != nil {
		clog.Fatal(err)
	}
	conf.ParseAll()

	logFile, err := os.OpenFile(*logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		clog.Fatal(err)
	}
	defer logFile.Close()

	//w := bufio.NewWriter(logFile)
	//defer w.Flush()

	clog.LogFile(logFile)

	sshOrigCMD := os.Getenv("SSH_ORIGINAL_COMMAND")
	if sshOrigCMD == "" {
		cmdLine(conf)
	}

	allowed_cmds := strings.Split(*cmds, ":")
	envStr := strings.Join(os.Environ(), ", ")

	var allowed bool
	var allowedCMD string

	for i := range allowed_cmds {
		if sshOrigCMD == allowed_cmds[i] {
			allowed = true
			allowedCMD = allowed_cmds[i]
		}
	}

	if allowed {

		clog.Infof("Executing allowd command: cmd: [ %s ] env: [ %s ]\n", allowedCMD, envStr)

		argv := strings.Split(allowedCMD, " ")
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			clog.Errorf("Error while Executing command: err: [ %s] cmd: [ %s] env: [ %s ]\n", err, allowedCMD, envStr)
			return
		}
		clog.Infof("Command finished: cmd: [ %s ] env: [ %s ]\n", allowedCMD, envStr)
	} else {
		clog.Infof("Unknown command ignored: cmd: [ %s ] env: [ %s ]\n", sshOrigCMD, envStr)
	}
}
