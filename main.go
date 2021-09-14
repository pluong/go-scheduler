package main

import (
	"flag"
	"fmt"
	"os/exec"
	s "strings"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/robfig/cron"
)

func main() {
	script := flag.String("script", "./script.ksh", "script that will be triggered")
	redisServer := flag.String("redis", "192.168.0.139:6379", "redis instance")
	cronInput := flag.String("cron", "5,10,15,20,25,30,35,40,45,50,55 * * * *", "cron expression")

	flag.Parse()

	strArr := s.Split(*cronInput, "-")
	cronExp := s.Join(strArr, " ")

	fmt.Println("Configuration values: [redis-server] =", *redisServer, " [script]] = ", *script, "[cron expr] =", *cronInput, strArr, cronExp)

	cancel := make(chan bool)
	c := cron.New()
	c.AddFunc(cronExp, func() { getMutexAndExec(*redisServer, *script, cancel) })
	c.Start()

	for {
		select {
		// case <-cancel:
		// 	fmt.Println("Scheduler has Stopped!")
		// 	c.Stop()
		default:
			fmt.Println("---------------- tick --------- ", time.Now().Format(time.ANSIC), "\n")
			time.Sleep(1 * time.Second)
		}

	}

}

func getMutexAndExec(redisServer string, script string, c chan bool) {
	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	client := goredislib.NewClient(&goredislib.Options{
		Addr: redisServer,
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	// Obtain a new mutex by using the same name for all instances wanting the
	// same lock.
	mutexname := "global-mutex"
	mutex := rs.NewMutex(mutexname, redsync.WithTries(1))

	// Obtain a lock for our given mutex. After this is successful, no one else
	// can obtain the same lock (the same mutex name) until we unlock it.
	if err := mutex.Lock(); err != nil {
		fmt.Println("!!! Lock is not available !!!")
		c <- true
		return
		// panic(err)
	} else {
		fmt.Println("Acquired lock", mutex.Name())
	}

	// Do your work that requires the lock.
	executeScript(script)
	time.Sleep(2 * time.Second)

	// Release the lock so other processes or threads can obtain a lock.
	if ok, err := mutex.Unlock(); !ok || err != nil {
		panic("unlock failed")
	}
}

func executeScript(script string) {
	cmd := exec.Command(script)
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
