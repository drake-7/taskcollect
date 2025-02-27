package daymap

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~kvo/libgo/defs"
	"git.sr.ht/~kvo/libgo/errors"

	"main/plat"
)

type dmJsonEntry struct {
	Text   string
	Type   string
	Id     int
	Start  string
	Finish string
	Title  string
}

// Get a list of lessons for the week from DayMap for a user.
func GetLessons(creds User) ([][]plat.Lesson, errors.Error) {
	var weekStartIdx, weekEndIdx int
	t := time.Now().In(creds.Timezone)

	now := time.Date(
		t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0,
		creds.Timezone,
	)

	weekday := now.Weekday()

	switch weekday {
	case 6:
		weekStartIdx = 2
		weekEndIdx = 6
	default:
		weekStartIdx = 1 - int(weekday)
		weekEndIdx = 5 - int(weekday)
	}

	weekStart := now.AddDate(0, 0, weekStartIdx)
	weekEnd := now.AddDate(0, 0, weekEndIdx)

	lessonsUrl := "https://gihs.daymap.net/daymap/DWS/Diary.ashx"
	lessonsUrl += "?cmd=EventList&from="
	lessonsUrl += weekStart.Format("2006-01-02") + "&to="
	lessonsUrl += weekEnd.Format("2006-01-02")

	client := &http.Client{}

	req, e := http.NewRequest("GET", lessonsUrl, nil)
	if e != nil {
		err := errors.New(e.Error(), nil)
		return nil, errors.New("GET request for lessonsUrl failed", err)
	}

	req.Header.Set("Cookie", creds.Token)

	resp, e := client.Do(req)
	if e != nil {
		err := errors.New(e.Error(), nil)
		return nil, errors.New("failed to get resp", err)
	}

	dmJson := []dmJsonEntry{}

	e = json.NewDecoder(resp.Body).Decode(&dmJson)
	if e != nil {
		err := errors.New(e.Error(), nil)
		return nil, errors.New("failed to decode JSON", err)
	}

	lessons := make([][]plat.Lesson, 5)

	for _, l := range dmJson {
		if l.Type != "Lesson" {
			continue
		}

		lesson := plat.Lesson{}
		lesson.Start, e = time.ParseInLocation("2006-01-02T15:04:05.0000000", l.Start, creds.Timezone)

		if e != nil {
			startIdx := strings.Index(l.Start, "(") + 1
			endIdx := strings.Index(l.Start, "000-")

			if startIdx == 0 || endIdx == -1 {
				return nil, plat.ErrInvalidDmJson.Here()
			}

			startStr := l.Start[startIdx:endIdx]
			startInt, e := strconv.Atoi(startStr)
			if e != nil {
				err := errors.New(e.Error(), nil)
				return nil, errors.New("(1) string -> int conversion failed", err)
			}

			lesson.Start = time.Unix(int64(startInt), 0)

			startIdx = strings.Index(l.Finish, "(") + 1
			endIdx = strings.Index(l.Finish, "000-")

			if startIdx == 0 || endIdx == -1 {
				return nil, plat.ErrInvalidDmJson.Here()
			}

			finishStr := l.Finish[startIdx:endIdx]
			finishInt, e := strconv.Atoi(finishStr)
			if e != nil {
				err := errors.New(e.Error(), nil)
				return nil, errors.New("(2) string -> int conversion failed", err)
			}

			lesson.End = time.Unix(int64(finishInt), 0)
		} else {
			lesson.End, e = time.ParseInLocation("2006-01-02T15:04:05.0000000", l.Finish, creds.Timezone)
			if e != nil {
				err := errors.New(e.Error(), nil)
				return nil, errors.New("failed to parse time", err)
			}
		}

		class := l.Title
		class = strings.TrimSpace(class)

		re, e := regexp.Compile("[0-9][A-Z]+[0-9]+")
		if e != nil {
			err := errors.New(e.Error(), nil)
			return nil, errors.New("failed to compile regex", err)
		}

		lesson.Room = re.FindString(class)
		roomIdx := re.FindStringIndex(class)
		if len(class) > 0 && len(roomIdx) > 0 && roomIdx[0] > 0 {
			lesson.Class = class[:roomIdx[0]-1]
		} else {
			lesson.Class = class
		}

		if !strings.HasPrefix(l.Text, "<div") && len(l.Text) > 0 {
			if strings.Contains(l.Text, "<div") {
				lesson.Notice = l.Text[:strings.Index(l.Text, "<div")]
			} else {
				lesson.Notice = l.Text
			}

			lesson.Notice = strings.ReplaceAll(
				lesson.Notice,
				`<img src="/daymap/images/buttons/roomChange.gif"/>&nbsp;`,
				"",
			)
		}

		if strings.Contains(class, "Mentor Group") {
			continue
		}

		day := time.Date(
			lesson.Start.In(creds.Timezone).Year(),
			lesson.Start.In(creds.Timezone).Month(),
			lesson.Start.In(creds.Timezone).Day(),
			0, 0, 0, 0,
			creds.Timezone,
		)

		i := 0

		for day.After(weekStart.AddDate(0, 0, i)) {
			i++
		}

		if i > 4 {
			continue
		}

		if _, err := defs.Get(lessons, i); err != nil {
			return nil, errors.New("number of days with lessons exceeded 5", err)
		}

		lessons[i] = append(lessons[i], lesson)
	}

	return lessons, nil
}
