package main

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"log"
	"os"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DbType string
const (
	DB_MASTERS DbType = "masters"
	DB_GROUPS = "groups"
	DB_USERS = "users"
)

var (
	validSections = []DbType { DB_MASTERS, DB_GROUPS, DB_USERS }
	
	ErrDatabaseLoaded = errors.New("database: this database has already been loaded")
	ErrSectionStart = errors.New("database: start of new section before ending previous section")
	ErrInvalidDbType = errors.New("database: invalid database type specified")
)

type DbEntry struct {
	Name string
	Access string
	CreatedBy string
	CreatedDate time.Time
	ModifiedBy string
	ModifiedDate time.Time
	Comments string
}

type Database struct {
	sync.RWMutex

	filename string
	Masters *list.List
	Groups *list.List
	Users *list.List
}

func WildcardCompare(s1 string, s2 string) bool {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	s1 = createRegex(s1)
	regex := regexp.MustCompile(s1)

	return regex.MatchString(s2)
}

func createRegex(s string) string {
	var wc string
	for i := 0; i < len(s); i++ {
		if s[i] == '*' {
			wc += ".*"
		} else if s[i] == '?' {
			wc += "."
		} else if s[i] >= '0' && s[i] <= '9' {
			wc += string(s[i])
		} else if s[i] >= 'a' && s[i] <= 'z' {
			wc += string(s[i])
		} else {
			wc += "\\" + string(s[i])
		}
	}

	return "^" + wc + "$"
}

func (db *Database)dump() {
	l := db.Users
	for e := l.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)
		fmt.Printf("%s %s %s\n", entry.Name, entry.Access, entry.CreatedDate.Format(time.RFC1123))
	}
}

func LoadDatabase(dbList *list.List, filename string) *Database {
	// Make sure we don't load duplicate databases
	for db := dbList.Front(); db != nil; db = db.Next() {
		f := db.Value.(*Database).filename
		if strings.ToLower(filename) == strings.ToLower(f) {
			return db.Value.(*Database)
		}
	}

	log.Printf("Loading database %s...\n", filename)
	file, err := os.Open(filename)
	if err != nil {
		log.Printf(err.Error())
		return nil
	}
	defer file.Close()

	db := new(Database)
	db.filename = filename
	db.Masters = list.New()
	db.Groups = list.New()
	db.Users = list.New()

	reader := bufio.NewReader(file)
	for err != io.EOF {
		var line string
		line, err = reader.ReadString('\n')

		line = strings.Trim(line, " \n")
		if strings.ToLower(line) == "masters {" {
			db.loadSection(reader, DB_MASTERS)
			log.Printf("  - loaded %d masters\n", db.Masters.Len())
		} else if strings.ToLower(line) == "groups {" {
			db.loadSection(reader, DB_GROUPS)
			log.Printf("  - loaded %d groups\n", db.Groups.Len())
		} else if strings.ToLower(line) == "users {" {
			db.loadSection(reader, DB_USERS)
			log.Printf("  - loaded %d users\n", db.Users.Len())
		}

	}

	dbList.PushBack(db)

	return db
}

func lineIsSection(line string) bool {
	for i := range validSections {
		if strings.ToLower(line) == string(validSections[i]) + " {" {
			return true
		}
	}

	return false
}

func (db *Database) loadSection(reader *bufio.Reader, section DbType) error {
	line, err := reader.ReadString('\n')
	for err != io.EOF {
		line = strings.Trim(line, " \t\n")
		
		// make sure we're not starting a new section before finishing the current one
		if lineIsSection(line) {
			return ErrSectionStart
		} else if line == "}" {
			return nil
		}

		if line != "" {
			if section == DB_MASTERS {
				db.Masters.PushBack(line)
			} else if section == DB_GROUPS || section == DB_USERS {
				splitLine := strings.Split(line, "\\")
				if len(splitLine) < 6 {
					// dbentry format is incorrect, skip it
					line, err = reader.ReadString('\n')
					continue
				}

				e := new(DbEntry)
				e.Name = splitLine[0]
				e.Access = splitLine[1]
				e.CreatedBy = splitLine[2]
				unixTime, _ := strconv.ParseInt(splitLine[3], 10, 64)
				e.CreatedDate = time.Unix(unixTime, 0)
				e.ModifiedBy = splitLine[4]
				unixTime, _ = strconv.ParseInt(splitLine[5], 10, 64)
				e.ModifiedDate = time.Unix(unixTime, 0)
				e.Comments = splitLine[6]
			
				if section == DB_GROUPS {
					if db.FindEntry(DB_GROUPS, e.Name) == nil {
						db.Groups.PushBack(e)
					}
				} else if section == DB_USERS {
					if db.FindEntry(DB_USERS, e.Name) == nil {
						db.Users.PushBack(e)
					}
				}
			}
		}

		line, err = reader.ReadString('\n')
	}

	return nil
}

func (db *Database)getDbList(entryType DbType) *list.List {
	switch (entryType) {
	case DB_MASTERS:
		return db.Masters
	case DB_GROUPS:
		return db.Groups
	case DB_USERS:
		return db.Users
	}
	
	return nil
}

func (db *Database) UserHasFlag(user string, flag byte) bool {
	// Does the user exist in the database?
	u := db.FindEntry(DB_USERS, user)
	if u == nil {
		return false
	}
	
	// Check to see if the user is in a group
	if strings.HasPrefix(u.Access, "%") {
		g := db.FindEntry(DB_GROUPS, u.Access)
		if g == nil {
			return false
		}
	
		// Does the group have the specified flag?
		if strings.Contains(g.Access, string(flag)) {
			return true
		}
	} else {
		// Does the user have the specified flag?
		if strings.Contains(u.Access, string(flag)) {
			return true
		}
	}

	return false
}

func (db *Database) FindEntry(entryType DbType, name string) *DbEntry {
	l := db.getDbList(entryType)
	if l == db.Masters {
		return nil
	}

	db.RLock()
	defer db.RUnlock()

	for e := l.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)
		if strings.ToLower(name) == strings.ToLower(entry.Name) {
			return entry
		}
	}

	return nil
}

func (db *Database) FindEntries(entryType DbType, name string) []DbEntry {
	l := db.getDbList(entryType)
	if l == db.Masters || l == nil {
		return nil
	}

	db.RLock()
	defer db.RUnlock()

	var entries []DbEntry

	for e := l.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)
		if WildcardCompare(name, entry.Name) {
			if entries == nil {
				entries = make([]DbEntry, 1)
				entries[0] = *entry
			} else {
				entries = append(entries, *entry)
			}
		}
	}

	return entries
}

func (db *Database) RemoveEntry(entryType DbType, name string) {
	db.Lock()
	defer db.Unlock()

	l := db.getDbList(entryType)
	if l == db.Masters {
		return
	}

	for e := l.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)
		if strings.ToLower(name) == strings.ToLower(entry.Name) {
			l.Remove(e)
			return
		}
	}
}