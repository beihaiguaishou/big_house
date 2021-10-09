package main

import "log"

func main() {
	registrationHouses, _ := InRegistrationHouses()
	for index := range registrationHouses {
		house, _ := HouseInfo(registrationHouses[index].Name)
		log.Printf("%+v %+v", registrationHouses[index], house)
	}
}
