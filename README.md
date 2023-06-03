# Mole

**A simple tool to Docker images with minimal isolation**

Docker has a lot of stuff built in, providing a great interface that feels like a VM.

That's great, but sometimes you just want to run it. No need for additional isolation if your running it on a VM.
No need to isolate the filesystem and make it hard to copy things in and out. No need for extra network interfaces and configuration.

That's what mole does. It's goal is to get as close as possible to "what if your docker image just runs on the VM instead of in a container".

Mole is there to run the image as easily as possible, deferring further isolation to other tools.

It integrates well with existing unix tools for other purposes, as it simply exposes the container as a directory to you.

## Features
- container is just a directory, you can copy files in and out any time
- add mounts using `$ mount --bind`
- get a shell with `$ chroot`
- uses host network interfaces
- no 3rd party dependencies

Mole is great when you're using docker images purely as a packaging format,
e.g. on a VM which whole purpose it is to run a docker image.


## Usage

### Installation

```
curl -Lso ./mole https://github.com/sauercrowd/mole/releases/download/v0.0.1/mole
chmod +x ./mole
```

### Running/Restarting a container

```
sudo ./mole run elasticsearch:8.7.1 ./es
```

That downloaded and setups the container in the directory `./es`, and
executes the default entrypoint & command using the user the image specifies.

It will only download & setup the directory if it doesn't yet exist.

Once a directory exist you can also directly run it by dropping the `image:tag` part, e.g.

```
sudo ./mole run ./es
```

### Deleting a container

```
sudo ./mole rm ./es
```


### Running a command in the environment

You can just use `chroot`

```
sudo chroot ./es
```

Currently the environment variables, user and namespaces wont be used

### Mounting another directory

If you want to the directory `./to_mount` to `/mount`, you can run

```
mount --bind ./to_mount ./es/mount
```
