package pkg

import(
	"os"
	"strconv"
)

func GetenvStringValue(key string, default_values string)string{
	if value := os.Getenv(key);value != ""{
		return value
	}else{
		return default_values
	}
}
func GetenvIntValue(key string, default_values int)int{
	if value,err := strconv.Atoi(os.Getenv(key));err!=nil{
		return default_values
	}else{
		return value
	}
}