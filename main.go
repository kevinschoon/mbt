package main

import (
	"encoding/json"
	"fmt"
	"github.com/gambol99/go-marathon"
	"github.com/jawher/mow.cli"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func maybe(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func mkdir(path string) (err error) {
	split := strings.Split(path, "/")
	for i, p := range split {
		if i > 0 {
			p = strings.Join(split[:i], "/") + "/" + p
		}
		if _, err = os.Stat(p); os.IsNotExist(err) {
			if err = os.Mkdir(p, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func write(path string, data []byte, force bool) (err error) {
	fmt.Printf("Writing to %s\n", path)
	if _, err = os.Stat(path); os.IsExist(err) && !force {
		return fmt.Errorf("Data already saved at %s", path)
	}
	return ioutil.WriteFile(path, data, 0644)
}

func save(path string, app marathon.Application, client marathon.Marathon, force bool) (err error) {
	dir := fmt.Sprintf("%s%s", path, app.ID)
	if err := mkdir(dir); err != nil {
		return err
	}
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}
	if err := write(dir+"/current", data, force); err != nil {
		return err
	}
	versions, err := client.ApplicationVersions(app.ID)
	if err != nil {
		return err
	}
	for _, version := range versions.Versions {
		prev, err := client.ApplicationByVersion(app.ID, version)
		if err != nil {
			return err
		}
		data, err := json.Marshal(prev)
		if err != nil {
			return err
		}
		if err = write(dir+"/"+version, data, force); err != nil {
			return err
		}
	}
	return nil
}

func ReadDir(dir string) (apps []*marathon.Application, err error) {
	walk := func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			split := strings.Split(path, "/")
			if split[len(split)-1] == "current" {
				raw, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				app := &marathon.Application{}
				if err = json.Unmarshal(raw, app); err != nil {
					return err
				}
				apps = append(apps, app)
			}
		}
		if info.Mode().IsDir() && path != dir {
			_, err = ReadDir(path)
		}
		return err
	}
	if err = filepath.Walk(dir, walk); err != nil {
		return nil, err
	}
	return apps, nil
}

func main() {
	app := cli.App("mbt", "Marathon Backup Tool")
	var (
		client       marathon.Marathon
		endpointFlag = app.StringOpt("e endpoint", "http://localhost:8080", "Marathon endpoint e.g. http://localhost:8080")
		userFlag     = app.StringOpt("u user", "", "HTTP Basic Auth user:password")
	)

	app.Spec = "[OPTIONS]"

	app.Before = func() {
		config := marathon.NewDefaultConfig()
		config.URL = *endpointFlag
		u := strings.Split(*userFlag, ":")
		if len(u) == 2 {
			config.HTTPBasicAuthUser = u[0]
			config.HTTPBasicPassword = u[1]
		}
		c, err := marathon.NewClient(config)
		maybe(err)
		client = c
	}
	app.Command("backup", "Backup the given Marathon endpoint", func(cmd *cli.Cmd) {
		var (
			path  = cmd.StringArg("PATH", "", "path to restore from")
			force = cmd.BoolOpt("f force", false, "force writing backup files")
		)
		cmd.Spec = "[OPTIONS] PATH"
		cmd.Action = func() {
			applications, err := client.Applications(url.Values{})
			maybe(err)
			maybe(mkdir(*path))
			for _, app := range applications.Apps {
				maybe(save(*path, app, client, *force))
			}
		}
	})
	app.Command("restore", "Restore the given Marathon endpoint", func(cmd *cli.Cmd) {
		var (
			path = cmd.StringArg("PATH", "", "path to restore from")
		)
		cmd.Spec = "[OPTIONS] PATH"
		cmd.Action = func() {
			apps, err := ReadDir(*path)
			maybe(err)
			for _, app := range apps {
				*app.Instances = 0
				*app.Fetch = make([]marathon.Fetch, 0)
				if _, err := client.CreateApplication(app); err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	})
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
