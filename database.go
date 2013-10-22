package main

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"log"
	"os"
	"io"
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
	Access interface{}
	CreatedBy string
	CreatedDate time.Time
	ModifiedBy string
	ModifiedDate time.Time
	Comment string
}

type Database struct {
	sync.RWMutex

	filename string
	Masters *list.List
	Groups *list.List
	Users *list.List
}

func (db *Database) dump() {
	l := db.Users

	for e := l.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)
		fmt.Println(entry)
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
			} else if section == DB_USERS {
				splitLine := strings.Split(line, "\\")
				if len(splitLine) < 6 {
					// dbentry format is incorrect, skip it
					line, err = reader.ReadString('\n')
					continue
				}

				if (strings.HasPrefix(splitLine[0], "%")) {
					// This is a group -- wrong section!!
					line, err = reader.ReadString('\n')
					continue
				}
				
				db.AddEntry(splitLine[0], splitLine[1], splitLine[2], splitLine[3], splitLine[4], splitLine[5], splitLine[6])
			} else if section == DB_GROUPS {
				splitLine := strings.Split(line, "\\")
                                if len(splitLine) < 6 {
                                        // dbentry format is incorrect, skip it
                                        line, err = reader.ReadString('\n')
                                        continue
                                }

                                if (!strings.HasPrefix(splitLine[0], "%")) {
                                        // Groups must have prefix '%' -- invalid group!
					line, err = reader.ReadString('\n')
					continue
                                }

				db.AddEntry(splitLine[0], splitLine[1], splitLine[2], splitLine[3], splitLine[4], splitLine[5], splitLine[6])
			}
		}

		line, err = reader.ReadString('\n')
	}

	return nil
}

func FlagsToInt(flags string) uint32 {
	var f uint32
	flags = strings.ToUpper(flags)

	for _, c := range flags {
		f |= 1 << uint32(25 - ('Z' - c))
	}

	return f
}

func IntToFlags(flags uint32) string {
	var f string

	f = ""
	for i := 0; i < 26; i++ {
		var testflag uint32 = 1 << uint32(25 - ('Z' - ('A' + i)))
		if (flags & testflag == testflag) {
			f += string('Z' - ('Z' - ('A' + i)))
		}
	}

	return f
}

func (db *Database) EntryHasAny(name, flags string) bool {
	var entry *DbEntry

	if (strings.HasPrefix(name, "%")) {
		if (strings.ContainsAny(name, "*?")) {
			// Group names cannot contain wildcards
			return false
		}
	}

	entries := db.FindEntries(name)
	if (entries == nil) {
		return false
	} else {
		for _, e := range entries {
			if (strings.ToLower(e.Name) == strings.ToLower(name)) {
				entry = e
				break
			}
		}
	}

	if (entry == nil) {
		return false
	}

	var access uint32

	switch t := entry.Access.(type) {
	case uint32: // Entry has flag access
		access = entry.Access.(uint32)
	case string: // Entry is grouped
		if (strings.HasPrefix(name, "%")) {
			// Groups cannot be grouped
			return false
		}

		if (strings.ContainsAny(entry.Access.(string), "*?")) {
			// User's group cannot have a wildcard
			return false
		}

		entries = db.FindEntries(entry.Access.(string))
		if (entries == nil) {
			return false
		}

		entry = entries[0]
		access = entry.Access.(uint32)
	default: // wtf???
		fmt.Println("unknown user access type: ", t)
		return false
	}

	for _, c := range flags {
		testflag := FlagsToInt(string(c))
                if (access & testflag == testflag) {
                        return true
                }
        }

	return false
}

func (db *Database) UserIsAutobanned(username string) bool {
	if (db.EntryHasAny(username, "S")) {
		return false
	}

	for e := db.Users.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)
		if (WildcardCompare(username, entry.Name)) {
			if (db.EntryHasAny(entry.Name, "B")) {
				return true
			}
		}
	}

	return false
}

func (db *Database) EntryExists(name string) bool {
	db.RLock()
	defer db.RUnlock()

/*
	l := db.getDbList(entryType)
	
	for e := l.Front(); e != nil; e = e.Next() {
		if l == db.Masters {
			fmt.Println(e)
		} else {
			entry := e.Value.(*DbEntry)
			if strings.ToLower(name) == strings.ToLower(entry.Name) {
				return true
			}
		}
	}
*/
	return false
}

func (db *Database) AddEntry(name, access, createdBy, createdDate, modifiedBy, modifiedDate, comment string) {
        entry := new(DbEntry)

	entryIsGroup := strings.HasPrefix(name, "%")

        entry.Name = name
	if (strings.ContainsAny(name, "*?")) {
		if (entryIsGroup) {
			// Group name cannot have wildcard
			return
		}
	}

	entries := db.FindEntries(name)
	if (entries != nil) {
		for _, e := range entries {
			if (strings.ToLower(e.Name) == strings.ToLower(name)) {
				// Entry already exists
				return
			}
		}
	}
	
	if (strings.HasPrefix(access, "%")) {
		if (entryIsGroup) {
			// Groups cannot be grouped!
			return
		}

		if (strings.ContainsAny(access, "*?")) {
			// Group cannot contain wildcard
			return
		}

		if (strings.ContainsAny(name, "*?")) {
			// Users with wildcards be grouped
			return
		}

		entry.Access = access
	} else {
		if (strings.ContainsAny(name, "*?")) {
			if (FlagsToInt(access) != 2) {
				// Wildcarded entry cannot has more than the "B" flag
				return
			}

			entry.Access = uint32(2)
		} else {
			entry.Access = FlagsToInt(access)
		}
	}

        entry.CreatedBy = createdBy
	unixTime, _ := strconv.ParseInt(createdDate, 10, 64)
        entry.CreatedDate = time.Unix(unixTime, 0)

	entry.ModifiedBy = modifiedBy
        unixTime, _ = strconv.ParseInt(modifiedDate, 10, 64)
        entry.ModifiedDate = time.Unix(unixTime, 0)

        entry.Comment = comment

	if entryIsGroup {
		db.Groups.PushBack(entry)
	} else {
		db.Users.PushBack(entry)
	}
}

func (db *Database) FindEntries(pattern string) []*DbEntry {
	db.RLock()
	defer db.RUnlock()

	var l *list.List
	if (strings.HasPrefix(pattern, "%")) {
		l = db.Groups
	} else {
		l = db.Users
	}

	var entries []*DbEntry

	for e := l.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)

		if WildcardCompare(entry.Name, pattern) {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (db *Database) RemoveEntries(entryType DbType, pattern string) {
	db.Lock()
	defer db.Unlock()
/*
	for e := l.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*DbEntry)
		if WildcardCompare(pattern, entry.Name) {
			l.Remove(e)
		}
	}
*/
}