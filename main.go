package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func livecheck(ip string, stopChan chan struct{}) {
	filename := fmt.Sprintf("/tmp/%s-livecheck.csv", ip)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	header := []string{"#ts", ip}
	writer.Write(header)
	writer.Flush()

	stamp := func(value string) {
		timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
		data := []string{timestamp, value}
		fmt.Println(data, ip)
		writer.Write(data)
		writer.Flush()
	}

	ping := func(ip string) ([]byte, error) {
		cmd := exec.Command("ping", "-w", "1", ip)
		return cmd.CombinedOutput()
	}

	for {
		select {
		case <-stopChan:
			fmt.Println("Received stop signal for IP:", ip)
			return
		case <-time.After(10 * time.Second):
			_, err := ping(ip)
			if err != nil {
				stamp("0")
			} else {
				stamp("1")
			}
		}
	}
}

func main() {
	ips := os.Args[1:]

	if len(ips) != 1 {
		log.Fatal("Please specify a comma-separated list of IP addresses as command line arguments")
	}

	ipSlice := strings.Split(ips[0], ",")

	done := make(chan struct{})
	for _, ip := range ipSlice {
		go livecheck(ip, done)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	done <- struct{}{}
}
