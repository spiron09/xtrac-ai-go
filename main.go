package main

import (
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	
	run_pipeline()
	// fmt.Println(os.Getenv("GOOGLE_API_KEY"))
	// agent_init()
	
	// r, err := srv.Users.Labels.List("me").Do()
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve labels: %v", err)
	// }
	// if len(r.Labels) == 0 {
	// 	fmt.Println("No labels found.")
	// 	return
	// }
	// fmt.Println("Labels:")
	// for _, l := range r.Labels {
	// 	fmt.Printf("- %s\n", l.Name)
	// }
}
