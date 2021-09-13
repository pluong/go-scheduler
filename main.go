package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/robfig/cron"
)

func main() {
	c := cron.New()
	c.AddFunc("* * * * *", executeScript)
	c.Start()

	for {
		fmt.Println("---------------- tick --------- %s\n", time.Now().Format(time.RFC3339))
		time.Sleep(1 * time.Second)
	}
}

func executeScript() {
	cmd := exec.Command("./script.ksh")
	// cmd.Dir = "/Users/paulluong/go/src/github.com/pluong/scheduler/"
	// cmd.Path = "."
	// fmt.Printf("current dir: %s \n", cmd.Dir)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Printf("could not execute command; %s\n", err.Error())
		return
	}

	fmt.Printf("%s", stdout)
}
