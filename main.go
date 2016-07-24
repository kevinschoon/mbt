package main

import (
	"encoding/json"
	"fmt"
	"github.com/gambol99/go-marathon"
	"github.com/jawher/mow.cli"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

var (
	app      = cli.App("marathon-bk", "Marathon Backup")
	endpoint = app.StringOpt("e endpoint", "http://localhost:8080", "Marathon endpoint e.g. http://localhost:8080")
	user     = app.StringOpt("u user", "", "HTTP Basic Auth user:password")
	force    = app.BoolOpt("f force", false, "Force by overwriting existing files")
	path     = app.StringOpt("p path", *endpoint, "Path to save to or restore from")
)

func failOnError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func getClient() (marathon.Marathon, error) {
	config := marathon.NewDefaultConfig()
	config.URL = *endpoint
	u := strings.Split(*user, ":")
	if len(u) == 2 {
		config.HTTPBasicAuthUser = u[0]
		config.HTTPBasicPassword = u[1]
	}
	return marathon.NewClient(config)
}

func mkdir(path string) (err error) {
	fmt.Printf("mkdir %s\n", path)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0775)
	}
	return err
}

func write(path string, data []byte) (err error) {
	fmt.Printf("Writing to %s\n", path)
	if _, err = os.Stat(path); os.IsExist(err) && !*force {
		return fmt.Errorf("Data already saved at %s", path)
	}
	return ioutil.WriteFile(path, data, 0644)
}

func read(path string) (data []byte, err error) {
	fmt.Printf("Reading from %s\n", path)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return data, fmt.Errorf("Path %s not found", path)
	}
	return ioutil.ReadFile(path)
}

func save(path string, app marathon.Application, client marathon.Marathon) (err error) {
	name := strings.Replace(app.ID, "/", "", -1)
	dir := fmt.Sprintf("%s/%s", path, name)
	if err := mkdir(dir); err != nil {
		return err
	}
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}
	if err := write(dir+"/current", data); err != nil {
		return err
	}
	versions, err := client.ApplicationVersions(name)
	if err != nil {
		return err
	}
	for _, version := range versions.Versions {
		prev, err := client.ApplicationByVersion(name, version)
		if err != nil {
			return err
		}
		data, err := json.Marshal(prev)
		if err != nil {
			return err
		}
		if err = write(dir+"/"+version, data); err != nil {
			return err
		}
	}
	return nil
}

func backup() {
	client, err := getClient()
	failOnError(err)
	applications, err := client.Applications(url.Values{})
	failOnError(err)
	failOnError(mkdir(*path))
	for _, app := range applications.Apps {
		failOnError(save(*path, app, client))
	}
}

func restore() {
	client, err := getClient()
	failOnError(err)
	file, err := ioutil.ReadDir(*path)
	failOnError(err)
	for _, file := range file {
		if file.IsDir() {
			app := &marathon.Application{}
			data, err := read(fmt.Sprintf("%s/%s/current", *path, file.Name()))
			failOnError(err)
			failOnError(json.Unmarshal(data, app))
			*app.Instances = 0
			*app.Fetch = make([]marathon.Fetch, 0)
			if _, err := client.CreateApplication(app); err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

func main() {
	app.Command("backup", "Backup the given Marathon endpoint", func(cmd *cli.Cmd) { cmd.Action = backup })
	app.Command("restore", "Restore the given Marathon endpoint", func(cmd *cli.Cmd) { cmd.Action = restore })
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
