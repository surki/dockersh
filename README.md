  * ![Build status](https://travis-ci.org/Yelp/dockersh.svg)
  * [![Coverage Status](https://coveralls.io/repos/Yelp/dockersh/badge.png)](https://coveralls.io/r/Yelp/dockersh)
  * [Dockerhub container](https://registry.hub.docker.com/u/yelp/dockersh/)

dockersh
========

A user shell for isolated, containerized environments.

What is this?
=============

dockersh is designed to be used as a login shell on machines with multiple interactive users.

When a user invokes dockersh, it will bring up a [Docker](https://docker.com/) container (if not already running), and
then spawn a new interactive shell in the container's namespace.

dockersh can be used as a shell in ``/etc/passwd`` or as an ssh ``ForceCommand``.


This allows you to have a single ssh process on the normal ssh port which places user
sessions into their own individual docker containers in a secure and locked down manner.

Why do I want this?
===================

You want to allow multiple users to ssh onto a single box, but you'd like some isolation
between those users. With dockersh each user enters their
own individual docker container (acting like a lightweight virtual machine), with their home directory mounted from the host
system (so that user data is persistent between container restarts), but with its own kernel namespaces for
processes and networking.


This means that the user is isolated from the rest of the system, and they can only see their own processes,
and have their own network stack. This gives better privacy between users, and can also be used for more easily
separating each user's processes from the rest of the system with per user constraints.


Normally to give users individual containers you have to run an ssh daemon in each
container, and either have have a different port for each user to ssh to or some nasty
Forcecommand hacks (which only work with agent forwarding from the client).


Dockersh eliminates the need for any of these techniques by acting like a regular
shell which can be used in ``/etc/passwd`` or as an ssh
[ForceCommand](http://www.openbsd.org/cgi-bin/man.cgi/OpenBSD-current/man5/sshd_config.5?query=sshd_config).  
This allows you to have a single ssh process, on the normal ssh port, and gives
a secure way to connect users into their own individual docker
containers.

SECURITY WARNING
================

dockersh tries hard to drop all privileges as soon as possible, including disabling 
the suid, sgid, raw sockets and mknod capabilities of the target process (and all children),
however this doesn't mean that it is safe enough to allow public access to dockersh containers!

*WARNING:* Whilst this project tries to make users inside containers have lowered privileges
and drops capabilities to limit users ability to escalate their privilege level, it is not certain
to be completely secure. Notably when Docker adds user namespace support, this can be used
to further lock down privileges.

*SECOND WARNING:* The dockersh binary needs the suid bit set so that it can make the syscalls to adjust
kernel namespaces, so any security issues in this code are likely to be exploitable to root.

Requirements
============

Linux >= 3.8

Docker >= 1.2.0

If you want to build it locally (rather than in a docker container), Go >= 1.2

Installation
============

With docker
-----------

(This is the recommended method).

Build the Dockerfile in the local directory into an image, and run it like this:

    $ docker build -t dockersh:build . && docker run --rm -i -v /usr/local/bin:/target  -it dockersh:build

Without docker
--------------

You need to install golang (>= 1.12), then you should just be able to run:

    go build

and a 'dockersh' binary will be generated in your current working
directory. N.B. This binary needs to be moved to where you would like to
install it (recommended ``/usr/local/bin``). This is done automatically if
you use the Docker based installed, but you need to do it manually if you're
compiling the binary yourself.

Invoking dockersh
=================

There are two main methods of invoking dockersh. Either:

1. Put the path to dockersh into ``/etc/shells``, and then change the users shell
   in /etc/passwd (e.g. ``chsh myuser -s /usr/local/bin/dockersh``)
1. Set dockersh as the ssh ``ForceCommand`` in the users ``$HOME/.ssh/config``, or
   globally in ``/etc/ssh/ssh_config``

*Note:* The dockersh binary needs the suid bit set to operate!

Configuration
=============

We use [gcfg](https://code.google.com/p/gcfg/) to read configs in an ini style format.

The global config file, ``/etc/dockershrc`` has a ``[dockersh]`` block in it, and zero or more ``[user "foo"]`` blocks.

This can be used to set settings globally or per user, and also to enable the setting
of settings in the (optional) per user configuration file (``~/.dockersh``), if enabled.

Config file values
------------------

Setting name  | Type | Description | Default value | Example value
------------- | ---- | ----------- | ------------- | -------------
imagename  | String | The name of the container image to launch for the user. The %u sequence will interpolate the username | busybox | ubuntu, or %u/mydockersh
containername | String | The name of the container (per user) which is launched. | %u_dockersh | %u-dsh
mounthome | Bool | If the users home directory should be mounted in the target container | false | true
mounttmp | Bool | If /tmp should be mounted into the target container (so that ssh agent forwarding works). N.B. Security risk | false | true
mounthometo | String | Where to map the user's home directory inside the container. | %h | /opt/home/myhomedir
mounthomefrom | String | Where to map the user's home directory from on the host. | %h | /opt/home/%u
usercwd | String | Where to chdir into the container when starting a shell. | %h | /
containerusername | String | Username which should be used inside the container. | %u | root
shell | String | The shell that should be started for the user inside the container. | /bin/ash | /bin/bash
mountdockersocket | Bool | If to mount the docker socket from the host. (DANGEROUS) | false | true
dockersocket | String | The location of the docker socket from the host. | /var/run/docker.sock | /opt/docker/var/run/docker.sock
entrypoint | String | The entrypoint for the persistent process to keep the container running | internal | /sbin/yoursupervisor
cmd | Array of Strings | Additional parameters to pass when launching the container as the command line | | -c'/echo foo'
env | Array of Strings | Environment variables to pass to docker when launching the container | | IAM_ROLE=%u
enableuserconfig | Bool | Set to true to enable reading of per user ``~/.dockersh`` files | false | true
enableuserimagename | Bool | Set to true to enable reading of imagename parameter from ``~/.dockersh`` files | false | true
enableusercontainername | Bool | Set to true to enable reading of containername parameter from ``~/.dockersh`` files. (Dangerous!) | false | true
enableusermounthome | Bool | Set to true to enable reading of mounthome parameter from ``~/.dockersh`` files | false | true
enableusermounttmp | Bool | Set to true to enable reading of mounttmp parameter from ``~/.dockersh`` files | false | true
enableusermounthometo | Bool | Set to true to enable reading of mounthometo parameter from ``~/.dockersh`` files | false | true
enableusermounthomefrom | Bool | Set to true to enable reading of mounthomefrom parameter from ``~/.dockersh`` files | false | true
enableuserusercwd | Bool | Set to true to enable reading of usercwd parameter from ``~/.dockersh`` files | false | true
enableusercontainerusername | bool | Set to true to enable reading of containerusername parameter from ``~/.dockersh`` files | false | true
enableusershell | Bool | Set to true to enable reading of shell parameter from ``~/.dockersh`` files | false | true
enableuserentrypoint | Bool | Set to true to enable users to set their own supervisor daemon / entry point to the container for PID 1 | false | true
enableusercmd | Bool | Set to true to enable users to set the additional command parameters to the entry point | false | true
enableuserenv | Bool | Set to true to enable users to set additional options to the docker container that's started. (Dangerous!) | false | true

Notes:

  * Boolean settings are set by just putting the setting name in the config (see examples below).
  * You must set both ``enableuserconfig`` and the specific ``enableuserxxx`` setting that you want in ``/etc/dockersh`` to
    get any values parsed from ``~/.dockersh``
  * Array values are represented by having the same config key appear multiple times, once per value.

Config interpolations
---------------------

The following sequences are interpolated if found in configuration variables:

Sequence | Interpolation
---------|--------------
%u | The username of the user running dockersh
%h | The homedirectory (from /etc/passwd) of the user running dockersh

Example configs
---------------

A very restricted environment, with only the busybox container, limited to 32M of memory, ``/etc/dockersh`` looks like this:

    [dockersh]
    imagename = busybox
    shell = /bin/ash
    usercwd = /

A fairly restricted shell environment, but with homedirectories and one admin user being allowed additional privs, set the following ``/etc/dockersh``

    [dockersh]
    imagename = ubuntu:precise
    shell = /bin/bash
    mounthome

    [user "someadminguy"]
    mounttmp
    mountdockersocket
    
In a less restrictive environment, you may allow users to choose their own container and shell, from a 'shell' container
they have uploaded to the registry, and have ssh agent forwarding working, with the following ``/etc/dockersh``

    [dockersh]
    imagename = "%u/shell"
    mounthome
    mounttmp
    enableuserconfig
    enableusershell

    [user "someadminguy"]
    mountdockersocket

And an example user's ``~/.dockersh``

    [dockersh]
    shell = /bin/zsh

Or just allowing your users to run whatever container they want:

    [dockersh]
    mounthome
    mounttmp
    enableuserconfig
    enableuserimagename

Caveats
=======

  * User namespaces are not supported (yet) so if users escalate to root inside the container, they can probably escape
  * Tty/Pty handling is not great - whilst things appear to work, they don't go well in unusual circumstances (e.g. your process being killed due to OOM).
  * This code *has not* been audited by a 3rd party or a container expert, there are probably issues waiting to be found!

TODO
====

 * How do we deal with changed settings (i.e. when to recycle the container)
    * Document just kill 1 inside the container?
 * Fix up go panics when exiting the root container.
 * getpwnam so that we can interpolate the user's shell from /etc/shells (if used in ForceCommand mode!)
 * Decent test cases
 * Use libcontainer a lot more, in favour of our code:
    * https://github.com/docker/libcontainer/pull/143 - better nsenter with cgroups
    * https://github.com/docker/libcontainer/pull/150 - better forkexec
 * Find a better way to make ssh agent sockets work than to bind /tmp

Contributing
============

Patches are very very welcome!

This is our first real Go project, so we apologise about the shoddy quality of the code.

Please make a branch and send us a pull request.

Please ensure that you use the supplied pre-commit hook to correctly format your code
with go fmt:

    ln -s hooks/pre-commit .git/hooks/pre-commit

Copyright
=========

Copyright (c) 2014 Yelp. Some rights are reserved (see the LICENSE file for more details).

