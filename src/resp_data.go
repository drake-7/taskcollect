package main

import "html/template"

// Primary (page, head, body)

type pageData struct {
	PageType string
	Head     headData
	Body     bodyData
}

type headData struct {
	Title string
	//CssFiles []string
}

type bodyData struct {
	ErrorData     errData
	LoginData     loginData
	TimetableData timetableData
	ResData       resData
	TasksData     tasksData
	TaskData      taskData
}

// Error Page

type errData struct {
	Heading string
	Message string
}

// Login page

type loginData struct {
	Failed bool
}

// Timetable

type timetableData struct {
	CurrentDay int
	Days       []ttDay
}

type ttDay struct {
	Day     string
	Lessons []ttLesson
}

type ttLesson struct {
	Class         string
	FormattedTime string
	Height        float64
	TopOffset     float64
	Room          string
	Teacher       string
	Notice        string
	Color         string
	BGColor       string
}

// Resources (/res page)

type resData struct {
	Heading string
	Classes []resClass
}

type resClass struct {
	Name     string
	ResItems []resItem
}

type resItem struct {
	Id       string
	Name     string
	Platform string //e.g. daymap, gclass
	Posted   string
	URL      string
}

// Resource (single resource)

type resourceData struct {
	Name     string
	Class    string
	Link     string
	Desc     template.HTML
	Posted   string
	ResLinks map[string]string
	Platform string
	Id       string
}

// Tasks

type tasksData struct {
	Heading   string
	TaskTypes []taskType
}

type taskType struct {
	Name       string
	HasDueDate bool
	Tasks      []taskItem
}

type taskItem struct {
	Id       string
	Name     string
	Platform string
	Class    string
	DueDate  string
	Posted   string
	URL      string
}

// Task (single task)

type taskData struct {
	Id           string
	Name         string
	Platform     string
	Class        string
	URL          string
	IsDue        bool
	DueDate      string
	Submitted    bool
	Desc         template.HTML
	HasResLinks  bool
	ResLinks     map[string]string
	HasWorkLinks bool
	WorkLinks    map[string]string
	HasUpload    bool
	Grade        string
	Comment      template.HTML
}

var loginPageData = pageData{
	PageType: "login",
	Head: headData{
		Title: "Login",
	},
	Body: bodyData{
		LoginData: loginData{
			Failed: false,
		},
	},
}

// TODO: Create a function for fetching these status codes then constructing the pageData

var statusNotFoundData = pageData{
	PageType: "error",
	Head: headData{
		Title: "404 Not Found",
	},
	Body: bodyData{
		ErrorData: errData{
			Heading: "404 Not Found",
			Message: "The requested resource was not found on the server.",
		},
	},
}

var statusServerErrorData = pageData{
	PageType: "error",
	Head: headData{
		Title: "500 Internal Server Error",
	},
	Body: bodyData{
		ErrorData: errData{
			Heading: "500 Internal Server Error",
			Message: "The server encountered an unexpected error and cannot continue.",
		},
	},
}

var statusForbiddenData = pageData{
	PageType: "error",
	Head: headData{
		Title: "403 Forbidden",
	},
	Body: bodyData{
		ErrorData: errData{
			Heading: "403 Forbidden",
			Message: "You do not have permission to access this resource.",
		},
	},
}
