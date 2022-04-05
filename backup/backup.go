package backup

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"time"
)

func BackupInit(id string, port string) {
	count := backupCheck(id)
	go primaryInit(count, id, port)

}

func primaryInit(count int, id string, port string) {
	fmt.Println("--- Primary process ---")
	fmt.Println("... initializing new backup")
	err := exec.Command("gnome-terminal", "-x", "go", "run", "main.go", "-id="+id, "-port="+port).Run()
	// err := exec.Command(
	// 	"cmd", "/C", "start", "powershell", "-noexit", "go", "run",
	// 	"./main.go", "-id="+id, "-port="+port).Run()
	if err != nil {
		fmt.Println(err)
	}
	for {
		count++
		filePath := "backup/data/counter" + id + ".data"
		err := ioutil.WriteFile(filePath, []byte(strconv.Itoa(count)), 0777)

		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}
}

func backupCheck(id string) int {
	fmt.Println("--- Backup check ---")
	filePath := "backup/data/counter" + id + ".data"
	countOld, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("... done")
		return 0
	}
	for {
		time.Sleep(1500 * time.Millisecond)
		filePath := "backup/data/counter" + id + ".data"
		countNew, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Println(err)
		}
		if string(countNew) == string(countOld) {
			fmt.Println("... done")
			count, _ := strconv.Atoi(string(countNew))
			return count
		}
		countOld = countNew
	}
}
