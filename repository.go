package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type Repository struct {
	DB gorm.DB
}

// HELPER FUNCTIONS
func repositoryFullName() string {
	return settings.RepositoryDirectory() + string(filepath.Separator) + settings.RepositoryFile()
}

//

func (r *Repository) InitDB() {
	var err error

	b, err := ExistsPath(settings.RepositoryDirectory())
	if err != nil {
		parrot.Error("Got error when reading repository directory", err)
	}

	if !b {
		CreatePath(settings.RepositoryDirectory())
	}

	r.DB, err = gorm.Open("sqlite3", repositoryFullName())
	if err != nil {
		parrot.Error("Got error when connect database", err)
	}
	r.DB.LogMode(settings.RepositoryLogMode())

	/*
		r.DB.Ping()
		r.DB.SetMaxIdleConns(10)
		r.DB.SetMaxOpenConns(100)
	*/

	// Disable table name's pluralization
	r.DB.SingularTable(true)
}

func (r *Repository) InitSchema() {
	r.DB.AutoMigrate(&Command{})
}

func (r *Repository) BackupSchema() {
	b, _ := ExistsPath(settings.RepositoryDirectory())
	if !b {
		return
	}

	err := os.Rename(repositoryFullName(), repositoryFullName()+".bkp")

	if err != nil {
		parrot.Error("Warning", err)
	}
}

// functionalities

func (r *Repository) Put(c Command) {
	r.DB.Create(&c)
}

func (r *Repository) GetOne() Command {
	command := Command{}
	r.DB.First(&command)
	return command
}

func (r *Repository) FindById(id int) Command {
	command := Command{}
	r.DB.Where("Id = ?", id).Find(&command)
	return command
}

func (r *Repository) GetAllCommands() []Command {
	commands := []Command{}
	r.DB.Find(&commands)
	return commands
}

func (r *Repository) GetHistory(count int) []Command {
	commands := []Command{}
	r.DB.Order("terminated_at desc").Find(&commands).Count(&count)
	return commands
}

func (r *Repository) GetExecutedCommands(count int) []ExecutedCommand {
	commands := []Command{}

	r.DB.Order("terminated_at desc").Find(&commands).Count(&count)

	executedCommands := make([]ExecutedCommand, len(commands))

	for i := 0; i < len(commands); i++ {
		executedCommands[i] = commands[i].AsExecutedCommand(i)
	}

	return executedCommands
}
