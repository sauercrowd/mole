package mole

import (
    "regexp"
    "os"
    "strings"
    "os/exec"
    "bufio"
    "errors"
    "strconv"
    "syscall"
)



type UserGroup struct {
    Uid int
    Gid int
}

func GetUidForUser(root, username string) (int, error) {
    userFormat := regexp.MustCompile(`^([^:]+):[^:]+:([0-9]+)`)

    file, err := os.Open(root + "/etc/passwd")
    if err != nil {
       return -1, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    // optionally, resize scanner's capacity for lines over 64K, see next example
    for scanner.Scan() {
        line := scanner.Text()
        matches := userFormat.FindStringSubmatch(line)

        if len(matches) == 3 {
            if matches[1] == username {

                uid, err := strconv.Atoi(matches[2])

                if err != nil {
                    return -1, err
                }

                return uid, nil
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return -1, err
    }

    return -1, nil

}


func getGidForGroup(root, group string) (int, error) {
    userFormat := regexp.MustCompile(`^([^:]+):[^:]+:([0-9]+)`)

    file, err := os.Open(root + "/etc/group")
    if err != nil {
       return -1, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        matches := userFormat.FindStringSubmatch(line)
        if len(matches) == 3 {
            if matches[1] == group {

                uid, err := strconv.Atoi(matches[2])

                if err != nil {
                    return -1, err
                }

                return uid, nil
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return -1, err
    }

    return -1, nil

}

func RunContainer(config *ImageConfig, targetDir string) error {

	cmd := exec.Command(config.Entrypoint[0], append(config.Entrypoint[1:], config.Cmd...)...)
	cmd.Stdin = os.Stdin
	cmd.Env = config.Env
	cmd.Stdout = os.Stdout

	if config.User != "" {
		splitted := strings.Split(config.User, ":")
		uid, err := GetUidForUser(targetDir, splitted[0])

		if err != nil {
			return err
		}

		if uid == -1 {
			return errors.New("User not found")
		}

		gid, err := GetUidForUser(targetDir, splitted[1])

		if err != nil {
			return err
		}

		if gid == -1 {
			return errors.New("Group not found")
		}

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
			Chroot:     targetDir,
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
		}

	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Chroot:     targetDir,
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		}
	}

	cmd.Stderr = os.Stderr

	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	} else {
		cmd.Dir = "/"
	}


	if err := cmd.Run(); err != nil {
		return err
	}

    return nil

}
