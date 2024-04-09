package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
	"webapp/pkg/data"
)

var pathToTemplates = "./templates/"
var uploadPath = "./static/img/"

func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	// start session //
	// template data
	var td = make(map[string]any)

	if app.Session.Exists(r.Context(), "test") {
		msg := app.Session.GetString(r.Context(), "test")
		td["test"] = msg
	} else {
		app.Session.Put(r.Context(), "test", "Hit this page at "+time.Now().UTC().String())
	}
	// end session //

	slog.Info("Home endpoint")
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{Data: td})
}

func (app *application) Profile(w http.ResponseWriter, r *http.Request) {

	_ = app.render(w, r, "profile.page.gohtml", &TemplateData{})
}

type TemplateData struct {
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User
}

func (app *application) render(w http.ResponseWriter, r *http.Request, t string, td *TemplateData) error {
	// parse the template from disk.
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return err
	}

	// get the template data
	td.IP = app.ipFromContext(r.Context())
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.PopString(r.Context(), "flash")

	// execute the parsed template and write to the response writer.
	if app.Session.Exists(r.Context(), "user") {
		td.User = app.Session.Get(r.Context(), "user").(data.User)
	}
	err = parsedTemplate.Execute(w, td)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// validate data
	form := NewForm(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		// redirect to login page with an error msg
		app.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		// redirect to login page with an error msg
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Println(password, user.FirstName)

	// authenticate the user with the pass
	if !app.authenticate(r, user, password) {
		// if not, put the error into the session and redirect to homepage
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// if not auth then redirect with error

	// prevent fixation attack
	_ = app.Session.RenewToken(r.Context())

	// store success message in session

	// redirect to some other page
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}

	// put the user into the session to know that the user is logged in
	app.Session.Put(r.Context(), "user", user)
	return true
}

func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "user") {
			app.Session.Put(r.Context(), "error", "Log in first!")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) UploadProfilePic(w http.ResponseWriter, r *http.Request) {
	// call a function that extracts the file from upload (request)
	files, err := app.UploadFiles(r, uploadPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get the user from the session
	user := app.Session.Get(r.Context(), "user").(data.User)

	// create a variable of type data.UserImage
	var i = data.UserImage{
		UserID:   user.ID,
		FileName: files[0].OriginalFileName,
	}

	// insert the user image into user_images
	_, err = app.DB.InsertUserImage(i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// refresh the sessional variable "user" with the new user info
	updatedUser, err := app.DB.GetUser(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Session.Put(r.Context(), "user", updatedUser)

	// redirect back to profile page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

type UploadedFile struct {
	OriginalFileName string
	FileSize         int64
}

func (app *application) UploadFiles(r *http.Request, uploadDir string) ([]*UploadedFile, error) {
	var uploadedFiles []*UploadedFile

	err := r.ParseMultipartForm(int64(1024 * 1024 * 5))
	if err != nil {
		return nil, fmt.Errorf("the uploaded file is too big, and must be less than %d bytes", 1024*1024*5)
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				uploadedFile.OriginalFileName = hdr.Filename

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.OriginalFileName)); nil != err {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}
