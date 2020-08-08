package main

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	bolt "go.etcd.io/bbolt"
)

func main() {
	// open a bolt database
	db, err := bolt.Open("notes.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("Notes"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	a := app.New()                   // create app
	w := a.NewWindow("Tappable")     // create window
	w.Resize(fyne.NewSize(500, 500)) // resize window w, h in px

	title := widget.NewEntry() // add new text entry for note title (saving)
	title.SetPlaceHolder("Shopping list")

	content := widget.NewEntry() // add new text entry for note
	content.SetPlaceHolder("Buy egg\nBuy potato")
	content.MultiLine = true

	title2 := widget.NewEntry() // add new text entry for note title (loading)
	title2.SetPlaceHolder("Shopping list")

	msg := widget.NewLabel("")  // Success/Error message.
	note := widget.NewLabel("") // Success/Error message.

	w.SetContent(widget.NewVBox( // set all these elements as content of window
		widget.NewLabel("Save Notes App (Premium Edition)"),
		title,
		content,
		// make a button with Save text and floppy disk icon,
		// on click save the note and display error/success
		widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
			if title.Text == "" || content.Text == "" {
				msg.SetText("Enter both a title and content for the note.")
				return
			}

			db.Update(func(tx *bolt.Tx) error {
				// set b to the bucket "Notes"
				b := tx.Bucket([]byte("Notes"))
				// key is title, value is note content
				err := b.Put([]byte(title.Text), []byte(content.Text))
				if err == nil {
					msg.SetText("Saved!")
				} else {
					msg.SetText("Error. Check console for details.")
				}
				return err
			})
			title.SetText("")
			content.SetText("")
		}),
		msg,
		title2,
		widget.NewButtonWithIcon("Get", theme.FileTextIcon(), func() {
			db.Update(func(tx *bolt.Tx) error {
				// set b to the bucket "Notes"
				b := tx.Bucket([]byte("Notes"))
				// key is title, value is note content
				v := b.Get([]byte(title2.Text))

				if v == nil {
					msg.SetText("This note is empty!")
					return nil
				}
				note.SetText(string(v))
				msg.SetText("Received!")
				return nil
			})
		}),
		widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
			db.Update(func(tx *bolt.Tx) error {
				// set b to the bucket "Notes"
				b := tx.Bucket([]byte("Notes"))
				// key is title, value is note content
				b.Delete([]byte(title2.Text))
				msg.SetText("Deleted!")
				return nil
			})
		}),
		note,
		widget.NewButtonWithIcon("Delete All (Permanent!!)", theme.DeleteIcon(), func() {
			db.Update(func(tx *bolt.Tx) error {
				// delete the whole ass bucket and remake
				tx.DeleteBucket([]byte("Notes"))
				_, err := tx.CreateBucket([]byte("Notes"))
				if err != nil {
					return fmt.Errorf("create bucket: %s", err)
				}
				msg.SetText("Deleted ur nasty notes 👀")
				return nil
			})
		}),
	))
	w.ShowAndRun() // show the window and run the program
}
