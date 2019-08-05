package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var debug bool

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug logging. Default : 'false'")
	flag.Parse()

	lvl, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		ll, err := logrus.ParseLevel(lvl)
		if err == nil {
			logrus.SetLevel(ll)
		}
	} else {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}
	}
}

func main() {
	logrus.Debug("Starting dockersh")

	logrus.Debug("Loading all config files")
	config, err := loadAllConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load config: %v\n", err)
		return
	}
	logrus.Debugf("Config dump: %+v", config)

	logrus.Debugf("Checking for container: name=%v", config.ContainerName)
	id, err := isContainerRunning(config.ContainerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not check container status: %v\n", err)
		return
	}
	logrus.Debugf("Container running? %v", id != "")

	if id == "" {
		logrus.Debug("Container is not running, starting it")
		id, err = startContainer(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not start container: %s\n", err)
			return
		}
	}

	logrus.Debugf("Container ID: %v", id)
	logrus.Debug("Exec into the container")

	err = execContainer(id, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not exec into container: %v\n", err)
	}
	// _, _, groups, _, err := user.GetUserGroupSupplementaryHome(username, 65536, 65536, "/")
	// err = nsenterexec(config.ContainerName, uid, gid, groups, config.UserCwd, config.Shell)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error starting shell in new container: %v\n", err)
	// 	return 1
	// }

	return
}

func tmplConfigVar(template string, v *configInterpolation) string {
	shell := "/bin/bash"
	r := strings.NewReplacer("%h", v.Home, "%u", v.User, "%s", shell) // Arguments are old, new ...
	return r.Replace(template)
}

func getInterpolatedConfig(config *Configuration, configInterpolations configInterpolation) error {
	config.ContainerUsername = tmplConfigVar(config.ContainerUsername, &configInterpolations)
	config.MountHomeTo = tmplConfigVar(config.MountHomeTo, &configInterpolations)
	config.MountHomeFrom = tmplConfigVar(config.MountHomeFrom, &configInterpolations)
	config.ImageName = tmplConfigVar(config.ImageName, &configInterpolations)
	config.Shell = tmplConfigVar(config.Shell, &configInterpolations)
	config.UserCwd = tmplConfigVar(config.UserCwd, &configInterpolations)
	config.ContainerName = tmplConfigVar(config.ContainerName, &configInterpolations)

	for i, o := range config.DockerOpt {
		config.DockerOpt[i] = tmplConfigVar(o, &configInterpolations)
	}

	return nil
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

// func realMain() int {
// 	err := dockerVersionCheck()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Docker version error: %v", err)
// 		return 1
// 	}
// 	username, homedir, uid, gid, err := getCurrentUser()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "could not get current user: %v", err)
// 		return 1
// 	}
// 	config, err := loadAllConfig(username, homedir)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Could not load config: %v\n", err)
// 		return 1
// 	}
// 	config.UserId = uid
// 	configInterpolations := configInterpolation{homedir, username}
// 	err = getInterpolatedConfig(&config, configInterpolations)
// 	if err != nil {
// 		panic(fmt.Sprintf("Cannot interpolate config: %v", err))
// 	}

// 	_, err = dockerpid(config.ContainerName)
// 	if err != nil {
// 		_, err = dockerstart(config)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "could not start container: %s\n", err)
// 			return 1
// 		}
// 	}
// 	_, _, groups, _, err := user.GetUserGroupSupplementaryHome(username, 65536, 65536, "/")
// 	err = nsenterexec(config.ContainerName, uid, gid, groups, config.UserCwd, config.Shell)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error starting shell in new container: %v\n", err)
// 		return 1
// 	}
// 	return 0
// }
