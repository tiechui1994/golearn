package main

import "log"

// log.Fatal和log.Panic
// log.Fatal 调用os.Exit() 导致应用退出
// log.Panic 调用panic() 导致应用退出
func Log()  {
	log.Fatal()
	log.Panic()
}
