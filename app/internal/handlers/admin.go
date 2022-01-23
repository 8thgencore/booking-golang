package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/8thgencore/bookings/internal/constants"
	"github.com/8thgencore/bookings/internal/forms"
	"github.com/8thgencore/bookings/internal/helpers"
	"github.com/8thgencore/bookings/internal/models"
	"github.com/8thgencore/bookings/internal/render"
	"github.com/go-chi/chi/v5"
)

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminAllReservations shows all reservations inu admin tool
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows the reservation in the admin tool
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	// get reservation from the database
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminPostShowReservation posts a reservation
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// AdminProcessReservation  marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		log.Println(err)
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	// http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// AdminDeleteReservation deletes a reservation
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	_ = m.DB.DeleteReservation(id)

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")
	// http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) getCalendarTime(r *http.Request) (time.Time, error) {
	var yearString, monthString string
	now := time.Now()

	if r.URL.Query().Get("y") == "" {
		yearString = m.App.Session.GetString(r.Context(), "calendar_current_year")
		monthString = m.App.Session.GetString(r.Context(), "calendar_current_month")
	} else {
		yearString = r.URL.Query().Get("y")
		monthString = r.URL.Query().Get("m")
	}

	currentDate, err := time.Parse("2006-01", fmt.Sprintf("%s-%s", yearString, monthString))

	if err != nil {
		return now, err
	}

	return currentDate.UTC(), nil
}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	currentDate, err := m.getCalendarTime(r)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	currentYear := currentDate.Format("2006")
	currentMonth := currentDate.Format("01")

	if r.URL.Query().Get("y") == "" || err != nil {
		http.Redirect(w, r,
			fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", currentYear, currentMonth),
			http.StatusSeeOther,
		)
		return
	}

	m.App.Session.Put(r.Context(), "calendar_current_year", currentYear)
	m.App.Session.Put(r.Context(), "calendar_current_month", currentMonth)

	next := currentDate.AddDate(0, 1, 0)
	previous := currentDate.AddDate(0, -1, 0)

	currentLocation := currentDate.Location()
	firstDayOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentLocation)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := map[string]interface{}{
		"now":         time.Now().UTC(),
		"currentDate": currentDate,
		"rooms":       rooms,
		"weeks":       helpers.GetMonthWeeks(firstDayOfMonth, lastDayOfMonth, time.Sunday),
		"weekDays":    helpers.GetWeekDays(),
	}

	for _, room := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for day := firstDayOfMonth; !day.After(lastDayOfMonth); day = helpers.NextDay(day) {
			reservationMap[day.Format(constants.DefaultDateFormat)] = 0
			blockMap[day.Format(constants.DefaultDateFormat)] = 0
		}

		// get all the restriction for the current room
		restrictions, err := m.DB.GetRoomRestrictionsByDate(room.ID, firstDayOfMonth, lastDayOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, restriction := range restrictions {
			for day := restriction.StartDate; !day.After(restriction.EndDate); day = helpers.NextDay(day) {
				if restriction.ReservationID > 0 {
					//it's a reservation
					reservationMap[day.Format(constants.DefaultDateFormat)] = restriction.ReservationID
				} else {
					// it's a block
					blockMap[day.Format(constants.DefaultDateFormat)] = restriction.ID
				}
			}
		}

		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: helpers.GetCalendarStringMap(previous, currentDate, next),
		Data:      data,
	})
}
