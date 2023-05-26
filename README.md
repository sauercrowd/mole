# Mole

**The minimal way to run docker images.**

Docker has a lot of stuff built in, provide an environment that feels like a VM.
But sometimes - we don't need that.

Docker images are great for packaging - and that's the only thing that mole focuses on.
Mole is applying the minimal amount of isolation necessary to run the image - that's it.
Great for just running an image on a VM.

Mole is following the Unix philosophy: a simple, clear purpose and interface.
It integrates well with existing unix tools for other purposes.

## Usage

### Installation

```
curl -s -o ./mole https://github.com/sauercrowd/mole/releases/download/v0.0.1/mole
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
