package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
)

// func dockerpid(name string) (pid int, err error) {
// 	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Pid}}", name)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return -1, errors.New(err.Error() + ":\n" + string(output))
// 	}

// 	pid, err = strconv.Atoi(strings.TrimSpace(string(output)))

// 	if err != nil {
// 		return -1, errors.New(err.Error() + ":\n" + string(output))
// 	}
// 	if pid == 0 {
// 		return -1, errors.New("Invalid PID")
// 	}
// 	return pid, nil
// }

func isContainerRunning(name string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	filter := filters.NewArgs()
	filter.Add("name", name)
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: filter})
	if err != nil {
		return "", err
	}

	if len(containers) >= 1 {
		return containers[0].ID, nil
	}

	return "", nil
}

func containerID(name string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	filter := filters.NewArgs()
	filter.Add("name", name)
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: filter})
	if err != nil {
		return "", err
	}

	if len(containers) >= 1 {
		return containers[0].ID, nil
	}

	return "", nil
}

// func dockersha(name string) (sha string, err error) {
// 	cmd := exec.Command("docker", "inspect", "--format", "{{.Id}}", name)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return sha, errors.New(err.Error() + ":\n" + string(output))
// 	}
// 	sha = strings.TrimSpace(string(output))
// 	if sha == "" {
// 		return "", errors.New("Invalid SHA")
// 	}
// 	return sha, nil
// }

func startContainer(config Configuration) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	id, err := containerID(config.ContainerName)
	logrus.Debugf("Checking if container with name %v already exists: %v", config.ContainerName, id != "")

	if id != "" {
		logrus.Debugf("Removing container, name: %v id: %v", config.ContainerName, id)
		err := cli.ContainerRemove(context.Background(), id,
			types.ContainerRemoveOptions{})
		if err != nil {
			return "", err
		}
	}

	binds := []string{"/etc/passwd:/etc/passwd:ro", "/etc/group:/etc/group:ro"}

	var init []string
	if config.Entrypoint == "internal" {
		init = []string{"/bin/sh", "-c", "trap : TERM INT; (while true; do sleep 1000; done) & wait"}
	} else {
		init = []string{config.Entrypoint}
	}
	logrus.Debugf("Entry point is: %v", init)

	// if len(config.DockerOpt) > 0 {
	// 	for _, element := range config.DockerOpt {
	// 		cmdtxt = append(cmdtxt, element)
	// 	}
	// }

	if config.MountTmp {
		logrus.Debugf("Bind mounting /tmp")
		binds = append(binds, "/tmp:/tmp:rw")
	}
	if config.MountHome {
		h := fmt.Sprintf("%s:%s:rw", config.MountHomeFrom, config.MountHomeTo)
		logrus.Debugf("Bind mounting home: %v", h)
		binds = append(binds, h)
	}
	if config.MountDockerSocket {
		logrus.Debugf("Bind mounting %v", config.DockerSocket)
		binds = append(binds, config.DockerSocket+":/var/run/docker.sock")
	}

	// if len(config.ReverseForward) > 0 {
	// 	cmdtxt, err = setupReverseForward(cmdtxt, config.ReverseForward)
	// 	if err != nil {
	// 		return []string{}, err
	// 	}
	// }

	// cmdtxt = append(cmdtxt, "--name", config.ContainerName, "--entrypoint", init, config.ImageName)
	// if len(config.Cmd) > 0 {
	// 	for _, element := range config.Cmd {
	// 		cmdtxt = append(cmdtxt, element)
	// 	}
	// } else {
	// 	cmdtxt = append(cmdtxt, "")
	// }

	hostname, _ := os.Hostname()

	ctx := context.Background()

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Hostname:     hostname,
			User:         fmt.Sprintf("%d", config.UserId), // User that will run the command(s) inside the container, also support user:group
			AttachStdin:  false,                            // Attach the standard input, makes possible user interaction
			AttachStdout: false,                            // Attach the standard output
			AttachStderr: false,                            // Attach the standard error
			//ExposedPorts:    "",                               // List of exposed ports
			Tty:             false,                   // Attach standard streams to a tty, including stdin if it is not closed.
			OpenStdin:       false,                   // Open stdin
			StdinOnce:       false,                   // If true, close stdin after the 1 attached client disconnects.
			Env:             []string{"hello=world"}, // List of environment variable to set in the container
			Healthcheck:     nil,                     // Healthcheck describes how to check the container is healthy
			ArgsEscaped:     false,                   // True if command is already escaped (meaning treat as a command line) (Windows specific).
			Image:           config.ImageName,        // Name of the image as it was passed by the operator (e.g. could be symbolic)
			Volumes:         map[string]struct{}{},   // List of volumes (mounts) used for the container
			WorkingDir:      "/",                     // Current directory (PWD) in the command will be launched
			Entrypoint:      init,                    // Entrypoint to run when starting the container
			NetworkDisabled: false,                   // Is network disabled
			MacAddress:      "",                      // Mac Address of the container
			Labels:          map[string]string{},     // List of labels set to this container
			StopSignal:      "",                      // Signal to stop a container
			StopTimeout:     nil,                     // Timeout (in seconds) to stop a container
			Shell:           []string{"/bin/bash"},   // Shell for shell-form of RUN, CMD, ENTRYPOINT
			//Cmd:             []string{"/init"},       // Command to run when starting the container
		},
		&container.HostConfig{
			Binds:           binds, // List of volume bindings for this container
			ContainerIDFile: "",    // File (path) where the containerId is written
			//LogConfig:       LogConfig,     // Configuration of the logs for this container
			//NetworkMode:     NetworkMode,   // Network mode to use for the container
			//PortBindings:    nat.PortMap,   // Port mapping between the exposed port (container) and the host
			//RestartPolicy:   RestartPolicy, // Restart policy to be used for the container
			AutoRemove:   true,       // Automatically remove container when it exits
			VolumeDriver: "",         // Name of the volume driver used to mount volumes
			VolumesFrom:  []string{}, // List of volumes to take from other container

			// 	"-v", "/etc/passwd:/etc/passwd:ro", "-v", "/etc/group:/etc/group:ro",

			// Applicable to UNIX platforms
			CapAdd:       []string{},                                       // List of kernel capabilities to add to the container
			CapDrop:      []string{"SETUID", "SETGID", "NET_RAW", "MKNOD"}, // List of kernel capabilities to remove from the container
			Capabilities: []string{},                                       // List of kernel capabilities to be available for container (this overrides the default set)
			//CgroupnsMode:    CgroupnsMode,                                     // Cgroup namespace mode to use for the container
			DNS:        []string{}, // List of DNS server to lookup
			DNSOptions: []string{}, // List of DNSOption to look for
			DNSSearch:  []string{}, // List of DNSSearch to look for
			ExtraHosts: []string{}, // List of extra hosts
			GroupAdd:   []string{}, // List of additional groups that the container process will run as
			//IpcMode:         IpcMode,                                          // IPC namespace to use for the container
			//Cgroup:          CgroupSpec,                                       // Cgroup to use for the container
			Links:       []string{}, // List of links (in the name:alias form)
			OomScoreAdj: 0,          // Container preference for OOM-killing
			//PidMode:         PidMode,                                          // PID namespace to use for the container
			Privileged:      false,               // Is the container in privileged mode
			PublishAllPorts: false,               // Should docker publish all exposed port for the container
			ReadonlyRootfs:  false,               // Is the container root filesystem in read-only
			SecurityOpt:     []string{},          // List of string values to customize labels for MLS systems, such as SELinux.
			StorageOpt:      map[string]string{}, // Storage driver options per container.
			Tmpfs:           map[string]string{}, // List of tmpfs (mounts) used for the container
			//UTSMode:         UTSMode,                                          // UTS namespace to use for the container
			//UsernsMode:      UsernsMode,                                       // The user namespace to use for the container
			ShmSize: 0,                   // Total shm memory usage
			Sysctls: map[string]string{}, // List of Namespaced sysctls used for the container
			Runtime: "",                  // Runtime to use with this container

			//Isolation: Isolation, // Isolation technology of the container (e.g. default, hyperv)

			// Contains container's resources (cgroups, ulimits)
			//Resources: a,

			// Mounts specs used by the container
			//Mounts: []mount.Mount{},

			// MaskedPaths is the list of paths to be masked inside the container (this overrides the default set of paths)
			MaskedPaths: []string{},

			// ReadonlyPaths is the list of paths to be set as read-only inside the container (this overrides the default set of paths)
			ReadonlyPaths: []string{},

			// Run a custom init inside the container, if null, use the daemon's configured settings
			Init: nil,
		},
		nil, config.ContainerName)
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func execContainer(id string, config Configuration) error {
	shell := "/bin/bash"

	// TODO: Move to proper docker client API

	args := []string{"/usr/bin/docker"}
	args = append(args, "exec")
	args = append(args, "--user")
	args = append(args, fmt.Sprintf("%d", config.UserId))
	args = append(args, "--workdir")
	args = append(args, config.MountHomeTo)
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		args = append(args, "--tty")
	}
	// TODO: Handle scp etc
	args = append(args, "--interactive")
	args = append(args, id)
	args = append(args, shell)
	args = append(args, "--login")

	if err := syscall.Exec(args[0], args, os.Environ()); err != nil {
		log.Fatal(err)
	}

	// pid, err := syscall.ForkExec(args[0], args,
	// 	&syscall.ProcAttr{
	// 		Env: os.Environ(),
	// 		Dir: "/",
	// 		//sys.Setsid
	// 		//sys.Setpgid
	// 		//sys.Setctty && sys.Ctty
	// 		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	// 		Sys:   &syscall.SysProcAttr{},
	// 	})
	// if err != nil {
	// 	return err
	// }

	// var wstatus syscall.WaitStatus
	// _, err = syscall.Wait4(pid, &wstatus, 0, nil)
	// if err != nil {
	// 	return err
	// }

	return nil
	// if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
	// 	log.Fatal(err)
	// }

	// pid, err := ForkExec(shell, []string{shell, "--login"}, &ProcAttr{
	// 	Env: os.Environ(),
	// 	Dir: wd,
	// 	//sys.Setsid
	// 	//sys.Setpgid
	// 	//sys.Setctty && sys.Ctty
	// 	Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	// 	Sys: &SysProcAttr{
	// 		Chroot:     fmt.Sprintf("/proc/%s/root", strconv.Itoa(containerpid)),
	// 		Credential: &Credential{Uid: uint32(uid), Gid: uint32(gid)}, //, Groups: []uint32(groups)},
	// 	},
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// if err != nil {
	// 	return err
	// }

	// ctx := context.Background()

	// if _, err := cli.ContainerInspect(ctx, id); err != nil {
	// 	return err
	// }

	// if !execConfig.Detach {
	// 	if err := dockerCli.In().CheckTty(execConfig.AttachStdin, execConfig.Tty); err != nil {
	// 		return err
	// 	}
	// }

	// response, err := client.ContainerExecCreate(ctx, options.container, *execConfig)
	// if err != nil {
	// 	return err
	// }

	// execID := response.ID
	// if execID == "" {
	// 	return errors.New("exec ID empty")
	// }

	// if execConfig.Detach {
	// 	execStartCheck := types.ExecStartCheck{
	// 		Detach: execConfig.Detach,
	// 		Tty:    execConfig.Tty,
	// 	}
	// 	return client.ContainerExecStart(ctx, execID, execStartCheck)
	// }
	// return interactiveExec(ctx, dockerCli, execConfig, execID)
	// return nil

	return nil
}

func dockerstart(config Configuration) (pid int, err error) {
	// cmd := exec.Command("docker", "rm", config.ContainerName)
	// _ = cmd.Run()
	// cmdtxt, err := dockercmdline(config)
	// if err != nil {
	// 	return -1, err
	// }
	// //fmt.Fprintf(os.Stderr, "docker %s\n", strings.Join(cmdtxt, " "))
	// cmd = exec.Command("docker", cmdtxt...)
	// var output bytes.Buffer
	// cmd.Stdout = &output
	// cmd.Stderr = &output
	// err = cmd.Run()
	// if err != nil {
	// 	return -1, errors.New(err.Error() + ":\n" + output.String())
	// }
	// return dockerpid(config.ContainerName)

	return 0, nil
}

// func dockercmdline(config Configuration) ([]string, error) {
// 	var err error
// 	bindSelfAsInit := false
// 	init := config.Entrypoint
// 	if init == "internal" {
// 		init = "/init"
// 		bindSelfAsInit = true
// 	}
// 	thisBinary := "/usr/local/bin/dockersh"
// 	if os.Getenv("SHELL") != "/usr/local/bin/dockersh" {
// 		thisBinary, _ = filepath.Abs(os.Args[0])
// 	}
// 	var cmdtxt = []string{"run", "-d", "-u", fmt.Sprintf("%d", config.UserId),
// 		"-v", "/etc/passwd:/etc/passwd:ro", "-v", "/etc/group:/etc/group:ro",
// 		"--cap-drop", "SETUID", "--cap-drop", "SETGID", "--cap-drop", "NET_RAW",
// 		"--cap-drop", "MKNOD"}
// 	if len(config.DockerOpt) > 0 {
// 		for _, element := range config.DockerOpt {
// 			cmdtxt = append(cmdtxt, element)
// 		}
// 	}
// 	if config.MountTmp {
// 		cmdtxt = append(cmdtxt, "-v", "/tmp:/tmp")
// 	}
// 	if config.MountHome {
// 		cmdtxt = append(cmdtxt, "-v", fmt.Sprintf("%s:%s:rw", config.MountHomeFrom, config.MountHomeTo))
// 	}
// 	if bindSelfAsInit {
// 		cmdtxt = append(cmdtxt, "-v", thisBinary+":/init")
// 	} else {
// 		if len(config.ReverseForward) > 0 {
// 			return []string{}, errors.New("Cannot configure ReverseForward with a custom init process")
// 		}
// 	}
// 	if config.MountDockerSocket {
// 		cmdtxt = append(cmdtxt, "-v", config.DockerSocket+":/var/run/docker.sock")
// 	}
// 	if len(config.ReverseForward) > 0 {
// 		cmdtxt, err = setupReverseForward(cmdtxt, config.ReverseForward)
// 		if err != nil {
// 			return []string{}, err
// 		}
// 	}
// 	cmdtxt = append(cmdtxt, "--name", config.ContainerName, "--entrypoint", init, config.ImageName)
// 	if len(config.Cmd) > 0 {
// 		for _, element := range config.Cmd {
// 			cmdtxt = append(cmdtxt, element)
// 		}
// 	} else {
// 		cmdtxt = append(cmdtxt, "")
// 	}

// 	return cmdtxt, nil
// }

// func validatePortforwardString(element string) error {
// 	parts := strings.Split(element, ":")
// 	if len(parts) != 2 {
// 		return errors.New("Number of parts must be 2")
// 	}
// 	if _, err := strconv.Atoi(parts[0]); err != nil {
// 		return (err)
// 	}
// 	if _, err := strconv.Atoi(parts[1]); err != nil {
// 		return (err)
// 	}
// 	return nil
// }

// func setupReverseForward(cmdtxt []string, reverseForward []string) ([]string, error) {
// 	for _, element := range reverseForward {
// 		err := validatePortforwardString(element)
// 		if err != nil {
// 			return cmdtxt, err
// 		}
// 	}
// 	cmdtxt = append(cmdtxt, "--env=DOCKERSH_PORTFORWARD="+strings.Join(reverseForward, ","))
// 	return cmdtxt, nil
// }
