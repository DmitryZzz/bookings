package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DmitryZzz/bookings/internal/models"
	"github.com/go-playground/assert/v2"
)

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{

	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"non-existent", "/green/eggs/and/ham", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new res", "/admin/reservations-new", "GET", http.StatusOK},
	{"all res", "/admin/reservations-all", "GET", http.StatusOK},
	{"show res", "/admin/reservations/all/1/show", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %q expected %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General`s Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCTX(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d wanted %d", rr.Code, http.StatusOK)
	}

	// test case where reservation is not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test with non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d wanted %d", rr.Code, http.StatusSeeOther)
	}

}

func TestRepository_PostReservation(t *testing.T) {

	testCases := []struct {
		name         string
		reqBody      string
		needReader   bool
		expectedCode int
	}{
		{
			name: "parse form ok",
			reqBody: "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=1",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name: "missing post data",
			reqBody: "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=1",
			needReader:   false,
			expectedCode: http.StatusSeeOther,
		},
		{
			name: "invalid start date",
			reqBody: "start_date=invalid&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=1",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name: "invalid end date",
			reqBody: "start_date=2050-01-01&end_date=invalid&first_name=John&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=1",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name: "invalid room id",
			reqBody: "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=invalid",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name: "invalid data",
			reqBody: "start_date=2050-01-01&end_date=2050-01-02&first_name=J&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=1",
			needReader:   true,
			expectedCode: http.StatusOK,
		},
		{
			name: "failure to insert reservation into database",
			reqBody: "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=2",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name: "failure to insert restriction into database",
			reqBody: "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com" +
				"&phone=123456789&room_id=1000",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.needReader {
				req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(tc.reqBody))
			} else {
				req, _ = http.NewRequest("POST", "/make-reservation", nil)
			}
			ctx := getCTX(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(Repo.PostReservation)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}

}
func TestRepository_PostAvailability(t *testing.T) {

	testCases := []struct {
		name         string
		reqBody      string
		needReader   bool
		expectedCode int
	}{
		{
			name:         "parse form ok",
			reqBody:      "start=2050-01-01&end=2050-01-02",
			needReader:   true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "missing post data",
			reqBody:      "start=2050-01-01&end=2050-01-02",
			needReader:   false,
			expectedCode: http.StatusSeeOther,
		},
		{
			name:         "invalid start date",
			reqBody:      "start=invalid&end=2050-01-02",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name:         "invalid end date",
			reqBody:      "start=2050-01-01&end=invalid",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name:         "invalid availability",
			reqBody:      "start=2050-01-01&end=2050-12-31",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
		{
			name:         "no rooms",
			reqBody:      "start=2050-01-01&end=2050-12-30",
			needReader:   true,
			expectedCode: http.StatusSeeOther,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.needReader {
				req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(tc.reqBody))
			} else {
				req, _ = http.NewRequest("POST", "/search-availability", nil)
			}
			ctx := getCTX(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(Repo.PostAvailability)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	var j jsonResponse

	// first case - rooms are available
	postedData := url.Values{}
	postedData.Add("start", "2050-01-01")
	postedData.Add("end", "2050-01-02")
	postedData.Add("room_id", "1")

	// create request
	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))

	// get context with session
	ctx := getCTX(req)
	req = req.WithContext(ctx)

	// set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// make handler handlerfunc
	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	// make response recorder
	rr := httptest.NewRecorder()

	// make request to our handler
	handler.ServeHTTP(rr, req)

	err := json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if !j.OK {
		t.Errorf("AvailableJson handler room available returned wrong answer: got %t wanted %t", j.OK, true)
	}

	// second case - rooms are not available
	postedData = url.Values{}
	postedData.Add("start", "2050-12-31")
	postedData.Add("end", "2050-12-31")
	postedData.Add("room_id", "1")

	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if j.OK {
		t.Errorf("AvailableJson handler room not available returned wrong answer: got %t wanted %t", j.OK, false)
	}

	// third case - missing post data
	req, _ = http.NewRequest("POST", "/search-availability-json", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if j.OK {
		t.Errorf("AvailableJson handler missing post data returned wrong answer: got %t wanted %t", j.OK, true)
	}
}

func TestRepository_ReservationSummary(t *testing.T) {

	// no reservation in session data
	req, _ := http.NewRequest("GET", "/reservation-summary", nil)

	ctx := getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(rr, req)

	expectedCode := http.StatusSeeOther

	assert.Equal(t, expectedCode, rr.Code)

	// got reservation in session data
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, "2050-01-01")
	endDate, _ := time.Parse(layout, "2050-01-02")
	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    1,
	}

	req, _ = http.NewRequest("GET", "/reservation-summary", nil)

	ctx = getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(rr, req)

	expectedCode = http.StatusOK

	assert.Equal(t, expectedCode, rr.Code)

}

func TestRepository_ChooseRoom(t *testing.T) {

	// missing URL parameter
	req, _ := http.NewRequest("GET", "/choose-room/invalid", nil)

	ctx := getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RequestURI = "/choose-room/invalid"
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	expectedCode := http.StatusSeeOther

	assert.Equal(t, expectedCode, rr.Code)

	// no reservation data in context
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)

	ctx = getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RequestURI = "/choose-room/1"
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	expectedCode = http.StatusSeeOther

	assert.Equal(t, expectedCode, rr.Code)

	// no reservation data in context
	reservation := models.Reservation{}

	req, _ = http.NewRequest("GET", "/choose-room/1", nil)

	ctx = getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RequestURI = "/choose-room/1"
	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	expectedCode = http.StatusSeeOther

	assert.Equal(t, expectedCode, rr.Code)

}

func TestRepository_BookRoom(t *testing.T) {

	// invalid room id
	req, _ := http.NewRequest("GET", "/book-room", nil)
	q := req.URL.Query()
	q.Add("s", "2050-01-01")
	q.Add("e", "2050-01-02")
	q.Add("id", "777")
	req.URL.RawQuery = q.Encode()

	ctx := getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(rr, req)

	expectedCode := http.StatusSeeOther

	assert.Equal(t, expectedCode, rr.Code)

	// all data ok
	req, _ = http.NewRequest("GET", "/book-room", nil)
	q = req.URL.Query()
	q.Add("s", "2050-01-01")
	q.Add("e", "2050-01-02")
	q.Add("id", "1")
	req.URL.RawQuery = q.Encode()

	ctx = getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(rr, req)

	expectedCode = http.StatusSeeOther

	assert.Equal(t, expectedCode, rr.Code)
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{"valid-credentials",
		"me@here.ca",
		http.StatusSeeOther,
		"",
		"/",
	},
	{"invalid-credentials",
		"jack@nimble.com",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{"invalid-data",
		"j",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

func TestLogin(t *testing.T) {
	// range through all test
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		// create a request
		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			// get the URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		// checking for expected values in HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			fmt.Println(e.expectedHTML)
			fmt.Println(rr.Body)
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s, but did not", e.name, e.expectedHTML)
			}
		}
	}
}

func getCTX(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
