# Mole

**A simole tool to run containers with minimal isolation**

Docker has a lot of stuff built in, provide an environment that feels like a VM.
But sometimes - we don't need that.


Mole is a simole tool to execute confainers, following the unix philosopher.
It integrates well with existing unix tools for other purposes, as it simply exposes the container as a directory to you.

It deliberately isolates as little as possible, focusing on simplicity

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
