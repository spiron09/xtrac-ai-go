package main

import (
	"github.com/joho/godotenv"
)

func main() {
	//Loading Env
	godotenv.Load()
	//Pipeline Run
	run_pipeline()	
}
