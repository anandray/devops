package main

import "os"
import "os/exec"
import "fmt"

func main() {
	
	cmd := "/usr/bin/sh"
	args := []string{"/root/t.sh"}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		//fmt.Fprintln(os.Stderr, err)
		
		fmt.Println("err: ", err)	
		
		os.Exit(1)
	}
	fmt.Println("plasma started Successfully")	
	
}
