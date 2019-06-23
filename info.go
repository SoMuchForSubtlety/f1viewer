package main

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// takes struct reflect Types and values and draws them as a table
func getTableValuesFromInterface(stru interface{}) ([]string, [][]string) {
	titles := reflect.TypeOf(stru)
	values := reflect.ValueOf(stru)
	t := make([]string, 1)
	v := make([][]string, 1)

	// iterate through titles and values and add them to the slices
	for i := 0; i < titles.NumField(); i++ {
		title := titles.Field(i)
		value := values.Field(i)

		if value.Kind() == reflect.Slice {
			lines := make([]string, value.Len())
			for j := 0; j < value.Len(); j++ {
				if value.Index(j).Kind() == reflect.String {
					lines[j] = value.Index(j).String()
				} else if value.Index(j).Kind() == reflect.Struct {
					a, b := getTableValuesFromInterface(value.Index(j).Interface())
					t = append(t, title.Name)
					v = append(v, []string{"================================"})
					t = append(t, a...)
					v = append(v, b...)
				}
			}
			t = append(t, title.Name)
			v = append(v, lines)
		} else if time, ok := value.Interface().(time.Time); ok {
			t = append(t, title.Name)
			v = append(v, []string{time.Format("2006-01-02 15:04:05")})
		} else if number, ok := value.Interface().(int); ok {
			t = append(t, title.Name)
			v = append(v, []string{strconv.Itoa(number)})
		} else if b, ok := value.Interface().(bool); ok {
			t = append(t, title.Name)
			v = append(v, []string{strconv.FormatBool(b)})
		} else if s, ok := value.Interface().(string); ok {
			lineArray := strings.FieldsFunc(s, func(r rune) bool { return r == '\n' || r == '\r' })
			t = append(t, title.Name)
			v = append(v, lineArray)
		} else {
			if !strings.Contains(strings.ToLower(title.Name), "winner") {
				t = append(t, title.Name)
				v = append(v, []string{value.String()})
			}
		}
	}
	return t, v
}

// TODO add channel to abort
// takes title and values slices and draws them as table
func (session *viewerSession) fillTableFromSlices(titles []string, values [][]string, abort chan bool) {
	select {
	case <-abort:
		// aborts previous call
	default:
		// so it doesn't lock
	}
	aborted := make(chan bool)
	go func() {
		// waits for abort signal
		abort <- true
		aborted <- true
	}()
	session.infoTable.Clear()
	rowIndex := 0
	for index, title := range titles {
		// convert supported API IDs to reasonable strings
		lines := session.convertIDs(values[index])
		select {
		case <-aborted:
			return
		default:
			if len(values[index]) > 0 && len(values[index][0]) > 1 {
				// print title
				session.infoTable.SetCell(rowIndex, 1, tview.NewTableCell(title).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorBlue))
				// print values
				for _, line := range lines {
					session.infoTable.SetCell(rowIndex, 2, tview.NewTableCell(line))
					rowIndex++
				}
				rowIndex++
			}
		}
	}
	session.infoTable.ScrollToBeginning()
	session.app.Draw()
}

// checks for driver or team IDs for the info table
func (session *viewerSession) convertIDs(lines []string) []string {
	if len(lines) < 1 {
		return lines
	}
	if len(lines[0]) > 12 && lines[0][:12] == "/api/driver/" {
		lines = session.substituteDriverNames(lines)
	} else if len(lines[0]) > 12 && lines[0][:10] == "/api/team/" {
		lines = session.substituteTeamNames(lines)
	}
	return lines
}

// turns slice of driver IDs to their names
func (session *viewerSession) substituteDriverNames(lines []string) []string {
	var wg sync.WaitGroup
	wg.Add(len(lines))
	// iterate over all lines
	for j := 0; j < len(lines); j++ {
		go func(j int) {
			defer wg.Done()
			// check if driver metadata is already cached
			session.driverMapMutex.RLock()
			driver, ok := session.driverMap[lines[j]]
			session.driverMapMutex.RUnlock()
			if !ok {
				var err error
				// load driver metadata if not already cached
				driver, err = getDriver(lines[j])
				if err != nil {
					return
				}
				// add metadata to cache
				session.driverMapMutex.Lock()
				session.driverMap[lines[j]] = driver
				session.driverMapMutex.Unlock()
			}
			// change string to driver name + number from metadata
			name := fmt.Sprintf("%4v "+driver.FirstName+" "+driver.LastName, "("+strconv.Itoa(driver.DriverRacingnumber)+")")
			lines[j] = name
		}(j)
	}
	wg.Wait()
	sort.Strings(lines)
	return lines
}

// turns array of team IDs to their names
func (session *viewerSession) substituteTeamNames(lines []string) []string {
	var wg sync.WaitGroup
	wg.Add(len(lines))
	// iterate over all lines
	for j := 0; j < len(lines); j++ {
		go func(j int) {
			defer wg.Done()
			// check if team metadata is already cached
			session.teamMapMutex.RLock()
			team, ok := session.teamMap[lines[j]]
			session.teamMapMutex.RUnlock()
			if !ok {
				// load team metadata if not already cached
				var err error
				team, err = getTeam(lines[j])
				if err != nil {
					return
				}
				// add metadata to cache
				session.teamMapMutex.Lock()
				session.teamMap[lines[j]] = team
				session.teamMapMutex.Unlock()
			}
			lines[j] = team.Name
		}(j)
	}
	wg.Wait()
	sort.Strings(lines)
	return lines
}
