package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

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
