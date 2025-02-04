package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var App fyne.App
var MainWindow fyne.Window

func loginPage(previousError string) {
	// Make warning banner
	warningBanner := container.New(layout.NewStackLayout(),
		canvas.NewRectangle(theme.Color(theme.ColorNameError)),
		container.NewHScroll(widget.NewLabel(previousError)),
	)
	if previousError == "" {
		warningBanner.RemoveAll()
	}

	// Make account icon
	accountIcon := canvas.NewImageFromResource(theme.Icon(theme.IconNameAccount))
	accountIcon.FillMode = canvas.ImageFillContain
	accountIcon.ScaleMode = canvas.ImageScaleSmooth
	accountIcon.SetMinSize(fyne.NewSquareSize(200))

	// Make username entry
	usernameEntry := widget.NewEntry()
	usernameEntry.PlaceHolder = "username"
	usernameEntry.Validator = validation.NewAllStrings(
		validation.NewRegexp("^.{4,}$", "Username must be longer than 4 characters"),
		validation.NewRegexp("^\\S*$", "Password can't contain spaces"),
	)

	// Make password entry
	passwordEntry := widget.NewEntry()
	passwordEntry.Password = true
	passwordEntry.PlaceHolder = "******"

	// Make login button
	loginButton := widget.NewButton("Login", func() {})

	// Create login action
	loginAction := func(_ string) {
		// Validate fields
		if err := usernameEntry.Validate(); err != nil {
			loginPage(err.Error())
			return
		}
		if err := passwordEntry.Validate(); err != nil {
			loginPage(err.Error())
			return
		}

		// Login
		if err := loginMongo(usernameEntry.Text, passwordEntry.Text); err != nil {
			loginPage(err.Error())
			return
		}

		// Display homepage
		homePage()
	}

	// Assign login function
	passwordEntry.OnSubmitted = loginAction
	usernameEntry.OnSubmitted = loginAction
	loginButton.OnTapped = func() { loginAction("BTN") }

	// Set content
	MainWindow.SetContent(container.New(layout.NewVBoxLayout(),
		warningBanner,
		accountIcon,
		layout.NewSpacer(),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("Username"), usernameEntry,
			widget.NewLabel("Password"), passwordEntry,
		),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), loginButton, layout.NewSpacer()),
		layout.NewSpacer(),
	))

	MainWindow.Show()
}

func homePage() {
	partList := container.New(layout.NewVBoxLayout())

	// Make search bar
	catalogSearchBar := widget.NewEntry()
	catalogSearchBar.PlaceHolder = "Search..."
	catalogSearchBar.OnSubmitted = func(s string) {
		// Display normally when search bar becomes empty
		if s == "" {
			homePage()
			return
		}

		// Search db
		searchTerms := strings.Split(s, ", ")
		filter := bson.M{"tags": bson.M{"$all": searchTerms}}
		cursor, err := collection.Find(context.TODO(), filter)
		if err != nil {
			return
		}

		// Get results
		var result []Part
		err = cursor.All(context.TODO(), &result)
		if err != nil {
			return
		}

		// Create UI
		partList.RemoveAll()
		parts := displayPartList(result)
		for _, val := range parts {
			partList.Add(val)
		}
	}

	// Make add button
	newItemButton := widget.NewButtonWithIcon("", theme.Icon(theme.IconNameContentAdd), addNewPart)

	// Make part list
	go func() {
		// Find all items
		cursor, err := collection.Find(context.TODO(), bson.D{})
		if err != nil {
			log.Println(err)
			return
		}

		// Retrieve all items
		var result []Part
		err = cursor.All(context.TODO(), &result)
		if err != nil {
			log.Println(err)
			return
		}

		// Create UI
		parts := displayPartList(result)
		for _, val := range parts {
			partList.Add(val)
		}
	}()

	// Make topbar
	topBar := container.New(layout.NewBorderLayout(nil, nil, nil, newItemButton),
		catalogSearchBar, newItemButton,
	)

	// Make layout
	MainWindow.SetContent(container.New(layout.NewBorderLayout(topBar, nil, nil, nil),
		topBar,
		container.NewVScroll(partList),
	))
}

func displayPartList(partList []Part) []*fyne.Container {
	var content []*fyne.Container

	for _, val := range partList {
		// Make name
		name := widget.NewLabel(val.Name)

		// Make tags
		tags := widget.NewLabel(strings.Join(val.Tags, ", "))
		tags.Alignment = fyne.TextAlignTrailing
		tagsObj := container.NewHScroll(tags)

		// Make explore button
		exploreBTN := widget.NewButtonWithIcon("", theme.Icon(theme.IconNameFolderOpen), func() { displayPart(val) })

		obj := container.New(layout.NewStackLayout(),
			canvas.NewRectangle(theme.Color(theme.ColorNameHeaderBackground)),
			container.New(layout.NewBorderLayout(nil, nil, name, exploreBTN),
				name, exploreBTN, tagsObj,
			),
		)

		content = append(content, obj)
	}

	return content
}

func displayPart(part Part) {
	window := App.NewWindow("Part \"" + part.Name + "\"")

	// Make name
	nameEntry := widget.NewEntry()
	nameEntry.Text = part.Name

	// Make tags
	tagsEntry := widget.NewEntry()
	tagsEntry.MultiLine = true
	tagsEntry.Wrapping = fyne.TextWrapBreak
	tagsEntry.Text = strings.Join(part.Tags, ", ")

	// Make Location
	locationEntry := widget.NewEntry()
	locationEntry.Text = part.Location

	// Make quantity
	quantityEntry := widget.NewEntry()
	quantityEntry.Text = fmt.Sprint(part.Qty)

	// Make form
	form := container.New(layout.NewFormLayout(),
		widget.NewLabel("Name: "), nameEntry,
		widget.NewLabel("Tags: "), tagsEntry,
		widget.NewLabel("Location: "), locationEntry,
		widget.NewLabel("Quantity: "), quantityEntry,
	)

	// Make save button
	saveBTN := widget.NewButtonWithIcon("Save", theme.Icon(theme.IconNameDocumentSave), func() {
		// Update local instance
		part.Name = nameEntry.Text
		part.Tags = strings.Split(tagsEntry.Text, ", ")
		part.Location = locationEntry.Text
		qty, err := strconv.ParseFloat(quantityEntry.Text, 64)
		if err == nil {
			part.Qty = qty
		}

		// Make updates
		filter := bson.D{{Key: "_id", Value: part.ID}}
		updateName := bson.D{{Key: "$set", Value: bson.D{{Key: "part-name", Value: part.Name}}}}
		updateTags := bson.D{{Key: "$set", Value: bson.D{{Key: "tags", Value: part.Tags}}}}
		updateLocation := bson.D{{Key: "$set", Value: bson.D{{Key: "location", Value: part.Location}}}}
		updateQTY := bson.D{{Key: "$set", Value: bson.D{{Key: "qty", Value: part.Qty}}}}

		// Push updates to cloud
		collection.UpdateOne(context.TODO(), filter, updateName)
		collection.UpdateOne(context.TODO(), filter, updateTags)
		collection.UpdateOne(context.TODO(), filter, updateLocation)
		collection.UpdateOne(context.TODO(), filter, updateQTY)

		// Update interface
		homePage()
	})

	// Make delete button
	deleteBTN := widget.NewButtonWithIcon("Delete", theme.Icon(theme.IconNameDelete), func() {
		// Make filter
		filter := bson.M{"_id": part.ID}

		// Delete entry
		collection.DeleteOne(context.TODO(), filter)

		// Reload pages
		homePage()
		window.Close()
	})

	// Make button collection
	BTNCollection := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), deleteBTN, layout.NewSpacer(), saveBTN, layout.NewSpacer())

	window.SetContent(container.New(layout.NewBorderLayout(nil, BTNCollection, nil, nil),
		BTNCollection, form,
	))
	window.Show()
}

func addNewPart() {
	window := MainWindow

	// Make name
	nameEntry := widget.NewEntry()

	// Make tags
	tagsEntry := widget.NewEntry()
	tagsEntry.MultiLine = true
	tagsEntry.Wrapping = fyne.TextWrapBreak

	// Make Location
	locationEntry := widget.NewEntry()

	// Make quantity
	quantityEntry := widget.NewEntry()

	// Make form
	form := container.New(layout.NewFormLayout(),
		widget.NewLabel("Name: "), nameEntry,
		widget.NewLabel("Tags: "), tagsEntry,
		widget.NewLabel("Location: "), locationEntry,
		widget.NewLabel("Quantity: "), quantityEntry,
	)

	// Make save button
	saveBTN := widget.NewButtonWithIcon("Create", theme.Icon(theme.IconNameDocumentSave), func() {
		// Create local instance
		var part Part

		// Update local instance
		part.ID = primitive.NewObjectID()
		part.Name = nameEntry.Text
		part.Tags = strings.Split(tagsEntry.Text, ", ")
		part.Location = locationEntry.Text
		qty, err := strconv.ParseFloat(quantityEntry.Text, 64)
		if err == nil {
			part.Qty = qty
		}

		// Push local instance to cloud
		_, err = collection.InsertOne(context.TODO(), part)
		if err != nil {
			log.Println(err)
		}

		// Update interface
		homePage()
	})

	// Make button collection
	BTNCollection := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), saveBTN, layout.NewSpacer())

	window.SetContent(container.New(layout.NewBorderLayout(nil, BTNCollection, nil, nil),
		BTNCollection, form,
	))
}
