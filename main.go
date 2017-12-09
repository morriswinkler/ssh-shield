/*
Command ssh-guard is a simple too to manage allowed ssh commands via the authorized_keys command parameter.

Install

       go get github.com/morriswinkler/ssh-guard

Setup

Upload a public ssh rsa key to the user folder on your ssh server, for my user id would be at:

       /home/morriswinkler/.ssh/authorized_keys

Prepend the key with the command parameter:

      cat /home/morriswinkler/.ssh/authorized_keys
      commadn="/home/morriswinkler/go/bin/ssh-guard" ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDyGB4+u1qRBNOpDGtQm1LgJXJMmRo+Dvu4WKbpwq29aSM+1KulQw+sJ9vhpKXZt5bqCCkv/2W+ScqSBP87AaFqT8tQ45f4tq6IYibYLjWT492qL948B7Yd2EEvVmP1K81uPvLLzgiuZ3Ci/1pa7kBEmxqI7itrD7g1A9BRixq74X3S/KvhEti/Nm8BGQBrg+8h05qyHG7qtQtwajbQDZsxAEN3OseZpI2n0WFBcJ84ic5lK8f01CBtRLPvwcu8/lpn7bW5MzC0ShyBT1OMBaUwzwfAfn9Tw9aoziAzmGFbW5OkuBObQKG6pSo2Th2C40fhTO1WoefHv2FT4BxhgpVv morriswinkler@ssh_server

Add more options to make it more secure:

      cat /home/morriswinkler/.ssh/authorized_keys
      from="8.8.8.8",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,commadn="/home/morriswinkler/go/bin/ssh-guard" ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDyGB4+u1qRBNOpDGtQm1LgJXJMmRo+Dvu4WKbpwq29aSM+1KulQw+sJ9vhpKXZt5bqCCkv/2W+ScqSBP87AaFqT8tQ45f4tq6IYibYLjWT492qL948B7Yd2EEvVmP1K81uPvLLzgiuZ3Ci/1pa7kBEmxqI7itrD7g1A9BRixq74X3S/KvhEti/Nm8BGQBrg+8h05qyHG7qtQtwajbQDZsxAEN3OseZpI2n0WFBcJ84ic5lK8f01CBtRLPvwcu8/lpn7bW5MzC0ShyBT1OMBaUwzwfAfn9Tw9aoziAzmGFbW5OkuBObQKG6pSo2Th2C40fhTO1WoefHv2FT4BxhgpVv morriswinkler@ssh_server

Thats it.

CommandManagement

By default no command is allowed, you can add commands by running:

         ssh-guard -add "ls /"

To list them use:

         ssh-guard -list

And finaly to remove them use:

        ssh-guard -del 1

Commands are stored as a colon separated list in ~/.config/ssh-guard/config.ini:

       allowed_commands = ls /: ps aux:ls

The log file contains ignored and allowed commands and invoations from the sshd, you can use it to find commands to add if needed.
You can change the location inside the config.

      logfile = ~/ssh-gaurd.log
*/
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
	logFileName = flag.String("logfile", "~/log/ssh-guard.log", "logfile path")
	cmds        = flag.String("allowd_commands", "", "Colon separated list of allowed commads, use add / del / list")
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
