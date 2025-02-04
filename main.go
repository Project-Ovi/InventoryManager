package main

import (
	"fyne.io/fyne/v2/app"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client
var collection *mongo.Collection

func _init() {
	// Create fyne application
	App = app.New()
	MainWindow = App.NewWindow("Part Manager")
	MainWindow.SetMaster()
}

func main() {
	// Init
	_init()

	go loginPage("")
	App.Run()

}
