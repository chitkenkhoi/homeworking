package main

import(
	"fmt"

	"lqkhoi-go-http-api/internal/app"
	
	"github.com/joho/godotenv"
	
	
)

func main(){
	godotenv.Load()
	if err := app.New();err!=nil{
		fmt.Println(err)
	}
}