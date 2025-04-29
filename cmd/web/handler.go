package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/obzva/snippetbox/internal/model"
	"github.com/obzva/snippetbox/internal/validator"
)

func ping(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			app.serverError(w, r, err.Error())
		}
	}
}

func getHome(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		snippets, err := app.snippetModel.Latest(r.Context())
		if err != nil {
			app.serverError(w, r, err.Error())
			return
		}

		td := newTemplateData(app, r)
		td.Snippets = snippets

		app.render(w, r, http.StatusOK, "home.tmpl", td)
	}
}

func getSnippetView(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get id
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			app.serverError(w, r, err.Error())
			return
		}
		if id < 1 {
			app.serverError(w, r, "id should be larger than or equal to 0", slog.Int("id", id))
			return
		}

		// get snippet
		s, err := app.snippetModel.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, model.ErrNoRecord) {
				app.clientError(w, http.StatusNotFound)
			} else {
				app.serverError(w, r, err.Error())
			}
			return
		}

		td := newTemplateData(app, r)
		td.Snippet = s

		app.render(w, r, http.StatusOK, "view.tmpl", td)
	}
}

type snippetCreateForm struct {
	Title   string
	Content string
	Expires int
}

func getSnippetCreate(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		td := newTemplateData(app, r)
		td.Form = snippetCreateForm{
			Expires: 365,
		}

		app.render(w, r, http.StatusOK, "create.tmpl", td)
	}
}

const (
	fieldTitle   = "title"
	fieldContent = "content"
	fieldExpires = "expires"
)

func postSnippetCreate(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		expires, err := strconv.Atoi(r.PostForm.Get(fieldExpires))
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		v := validator.NewValidator()
		form := snippetCreateForm{
			Title:   r.PostForm.Get(fieldTitle),
			Content: r.PostForm.Get(fieldContent),
			Expires: expires,
		}

		v.CheckField(validator.StringNotBlank(form.Title), fieldTitle, "this field cannot be blank")
		v.CheckField(validator.RunesMax(form.Title, 100), fieldTitle, "this field cannot be more than 100 characters long")
		v.CheckField(validator.StringNotBlank(form.Content), fieldContent, "this field cannot be blank")
		v.CheckField(validator.CheckPermitted(expires, 1, 7, 365), fieldExpires, "this field must be one of 1, 7, or 365")

		if !v.CheckValidity() {
			td := newTemplateData(app, r)
			td.Form = form
			td.FieldErrors = v.FieldErrors
			app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl", td)
			return
		}

		id, err := app.snippetModel.Insert(r.Context(), form.Title, form.Content, form.Expires)
		if err != nil {
			app.serverError(w, r, err.Error())
			return
		}

		app.sessionManager.Put(r.Context(), sessionKeyFlash, "Snippet was successfully created!")

		http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
	}
}

type userSignupForm struct {
	Name, Email, Password string
}

func getUserSignup(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		td := newTemplateData(app, r)
		td.Form = userSignupForm{}
		app.render(w, r, http.StatusOK, "signup.tmpl", td)
	}
}

const (
	fieldName     = "name"
	fieldEmail    = "email"
	fieldPassword = "password"
)

func postUserSignup(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		v := validator.NewValidator()
		form := userSignupForm{
			Name:     r.PostForm.Get(fieldName),
			Email:    r.PostForm.Get(fieldEmail),
			Password: r.PostForm.Get(fieldPassword),
		}

		v.CheckField(validator.StringNotBlank(form.Name), fieldName, "this field cannot be blank")
		v.CheckField(validator.StringNotBlank(form.Email), fieldEmail, "this field cannot be blank")
		v.CheckField(validator.StringMatch(form.Email, validator.EmailRegexp), fieldEmail, "this field must be a valid email address")
		v.CheckField(validator.StringNotBlank(form.Password), fieldPassword, "this field cannot be blank")
		v.CheckField(validator.RunesMin(form.Password, 8), fieldPassword, "this field must be at least 8 runes long")

		if !v.CheckValidity() {
			td := newTemplateData(app, r)
			td.Form = form
			td.FieldErrors = v.FieldErrors
			app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", td)
			return
		}

		err = app.userModel.Insert(r.Context(), form.Name, form.Email, form.Password)
		if err != nil {
			if errors.Is(err, model.ErrDuplicateEmail) {
				v.AddFieldError(fieldEmail, "email address is already in use")
				td := newTemplateData(app, r)
				td.Form = form
				td.FieldErrors = v.FieldErrors
				app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", td)
			} else {
				app.serverError(w, r, err.Error())
			}
			return
		}

		app.sessionManager.Put(r.Context(), sessionKeyFlash, "Your signup was successful. Please log in.")

		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	}
}

type userLoginForm struct {
	Email, Password string
}

func getUserLogin(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		td := newTemplateData(app, r)
		td.Form = userLoginForm{}

		app.render(w, r, http.StatusOK, "login.tmpl", td)
	}
}

func postUserLogin(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		v := validator.NewValidator()
		form := userLoginForm{
			Email:    r.PostForm.Get(fieldEmail),
			Password: r.PostForm.Get(fieldPassword),
		}

		v.CheckField(validator.StringNotBlank(form.Email), fieldEmail, "this field cannot be blank")
		v.CheckField(validator.StringMatch(form.Email, validator.EmailRegexp), fieldEmail, "this field must be a valid email address")
		v.CheckField(validator.StringNotBlank(form.Password), fieldPassword, "this field cannot be blank")

		if !v.CheckValidity() {
			td := newTemplateData(app, r)
			td.Form = form
			td.FieldErrors = v.FieldErrors
			td.NonFieldErrors = v.NonFieldErrors

			app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", td)
			return
		}

		id, err := app.userModel.Authenticate(r.Context(), form.Email, form.Password)
		if err != nil {
			if errors.Is(err, model.ErrInvalidCredentials) {
				v.AddNonFieldError("Email or password is incorrect")

				td := newTemplateData(app, r)
				td.Form = form
				td.FieldErrors = v.FieldErrors
				td.NonFieldErrors = v.NonFieldErrors

				app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", td)
			} else {
				app.serverError(w, r, err.Error())
			}
			return
		}

		err = app.sessionManager.RenewToken(r.Context())
		if err != nil {
			app.serverError(w, r, err.Error())
			return
		}

		app.sessionManager.Put(r.Context(), sessionKeyAuth, id)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func postUserLogout(app *application) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := app.sessionManager.RenewToken(r.Context())
		if err != nil {
			app.serverError(w, r, err.Error())
			return
		}

		app.sessionManager.Remove(r.Context(), sessionKeyAuth)

		app.sessionManager.Put(r.Context(), sessionKeyFlash, "You've been logged out successfully!")

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
