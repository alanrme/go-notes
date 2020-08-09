package main

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
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
	w.Resize(fyne.NewSize(800, 500)) // resize window w, h in px

	title := widget.NewEntry() // add new text entry for note title (saving)
	title.SetPlaceHolder("Shopping list")

	content := widget.NewEntry() // add new text entry for note
	content.SetPlaceHolder("Buy egg\nBuy potato")
	content.MultiLine = true

	title2 := widget.NewEntry() // add new text entry for note title (loading)
	title2.SetPlaceHolder("Shopping list")

	msg := widget.NewLabel("")  // Success/Error message
	note := widget.NewLabel("") // retrieved note

	// list of all the notes stored
	notes := widget.NewVBox()

	db.View(func(tx *bolt.Tx) error {
		// bucket is Notes
		b := tx.Bucket([]byte("Notes"))
		c := b.Cursor()
		// cycle through every key (k)
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			key := string(k) // cache this value so it exists when the button is clicked
			// append the key to the list, on click it fills it into the textbox
			notes.Append(widget.NewButton(key, func() {
				title2.SetText(key)
			}))
		}
		return nil
	})

	w.SetContent(fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		widget.NewVBox( // set all these elements as content of window
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
						key := title.Text // cache the value so it exists when button clicked
						notes.Append(widget.NewButton(key, func() {
							title2.SetText(key)
						}))
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
					title2.SetText("")
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
					title2.SetText("")
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
		),
		widget.NewScrollContainer(
			notes,
		),
	))
	w.ShowAndRun() // show the window and run the program
}
