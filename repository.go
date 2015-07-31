package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"path/filepath"
	"strconv"
	"github.com/bradhe/stopwatch"
)

type Repository struct {
	DB       *bolt.DB
}

// HELPER FUNCTIONS
func repositoryFullName() string {
	return settings.RepositoryDirectory() + string(filepath.Separator) + settings.RepositoryFile()
}

//

func (r *Repository) InitDB() {
	start := stopwatch.Start()

	var err error

	b, err := ExistsPath(settings.RepositoryDirectory())
	if err != nil {
		parrot.Error("Got error when reading repository directory", err)
	}

	if !b {
		CreatePath(settings.RepositoryDirectory())
	}

	r.DB, err = bolt.Open(repositoryFullName(), 0600, nil)
    if err != nil {
        parrot.Error("Got error creating repository directory", err)
    }
    
	watch := stopwatch.Stop(start)
    fmt.Printf("Milliseconds elappsed: %v\n", watch.Milliseconds())
}

func (r *Repository) InitSchema() error {
	err := r.DB.Update(func(tx *bolt.Tx) error {
	    _, err := tx.CreateBucket([]byte("Commands"))
	    if err != nil {
	        return fmt.Errorf("create bucket: %s", err)
	    }
	    return nil
	})
	
	return err
}

func (r *Repository) CloseDB() {
	if err := r.DB.Close(); err != nil {
		parrot.Error("Error", err)
	}
}

func (r *Repository) BackupSchema() {
	/*
	// Drop (delete) collection "Commands"
	if err := r.DB.Drop("Commands.bkp"); err != nil {
		parrot.Error("Commands.bkp collection cannot be deleted", err)
	}

	if err := r.DB.Rename("Commands", "Commands.bkp"); err != nil {
		parrot.Error("Commands.bkp collection cannot be created", err)
	}
	*/
}

// functionalities

func (r *Repository) Put(c Command) {
	err := r.DB.Update(func(tx *bolt.Tx) error {
	    b, err := tx.CreateBucketIfNotExists([]byte("Commands"))
	    if err != nil {
	        return err
	    }
	 
		encoded, err := json.Marshal(c)
	    if err != nil {
	        return err
	    }
	    //return b.Put([]byte(c.CreatedAt.Format(time.RFC3339)), encoded)
		return b.Put([]byte(c.ID), encoded)
	})	
	
	if err != nil {
	    parrot.Error("Error inserti data", err)
	}
}

/*
func (r *Repository) GetOne() Command {
	command := Command{}
	r.DB.First(&command)
	return command
}
*/

func (r *Repository) FindById(id string) Command {
	var command = Command{}
	
	err := r.DB.View(func(tx *bolt.Tx) error {
    	b := tx.Bucket([]byte("Commands"))
   		v := b.Get([]byte(id))
    	
		fmt.Printf("The answer is: %s\n", v)
    	
		err := json.Unmarshal(v, &command)
	    if err != nil {
	        return err
	    }
		
		return nil
	})
	
	if err != nil {
	    parrot.Error("Error getting data", err)
	}
	
	return command
}

func (r *Repository) GetAllCommands() []Command {
	commands := []Command{}

	r.DB.View(func(tx *bolt.Tx) error {
	    b := tx.Bucket([]byte("Commands"))
	    c := b.Cursor()
	
	    for k, v := c.First(); k != nil; k, v = c.Next() {
	        fmt.Printf("key=%s, value=%s\n", k, v)
	    }
	
	    return nil
	})

	return commands
}

func (r *Repository) GetHistory(count int) []Command {
	commands := []Command{}
	//r.DB.Order("terminated_at desc").Find(&commands).Count(&count)
	return commands
}

func (r *Repository) GetExecutedCommands(count int) []ExecutedCommand {
	commands := r.GetAllCommands()

	parrot.Info("Count is: " + strconv.Itoa(count))

	//r.DB.Limit(count).Order("terminated_at desc").Find(&commands)

	executedCommands := make([]ExecutedCommand, len(commands))

	for i := 0; i < len(commands); i++ {
		executedCommands[i] = commands[i].AsExecutedCommand(i)
	}

	return executedCommands
}

func extend(slice []Command, element Command) []Command {
	n := len(slice)
	if n == cap(slice) {
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]Command, len(slice), 2*len(slice)+1)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}

// Append appends the items to the slice.
// First version: just loop calling Extend.
func append(slice []Command, items ...Command) []Command {
	for _, item := range items {
		slice = extend(slice, item)
	}
	return slice
}
