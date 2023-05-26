package main

import (
	"fmt"
	"github.com/sauercrowd/mole/mole"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "help":
		printHelp()
		return

	case "run":
		doRun(os.Args)
		return

	case "rm":
		doRm(os.Args)
		return
	}

	printHelp()

}

func getBackingDir(targetDir string) string {
	filteredDst := strings.TrimSuffix(targetDir, "/")

	splitted := strings.Split(filteredDst, "/")

	splitted[len(splitted)-1] = fmt.Sprintf(".%s.backing", splitted[len(splitted)-1])

	return strings.Join(splitted, "/")
}

func doRun(args []string) {
	if len(args) == 3 {
		targetDir := args[2]

		config, err := mole.GetConfigFromDir(targetDir)

		if err != nil {
			log.Fatal(err)
		}

		err = mole.RunContainer(config, targetDir)

		if err != nil {
			log.Fatal(err)
		}

	} else if len(args) == 4 {
		imageTagCombo := args[2]
		targetDir := args[3]
		splittedImageCombo := strings.SplitN(imageTagCombo, ":", 2)

		image := splittedImageCombo[0]

		if strings.Count(image, "/") == 0 {
			image = fmt.Sprintf("library/%s", image)
		}

		run(image, splittedImageCombo[1], targetDir)
	} else {
		printHelp()
		return
	}
}

func doRm(args []string) {
	if len(args) == 3 {
		targetDir := args[2]

		umount(path.Join(targetDir, "proc"))

		umount(targetDir)

		if err := os.RemoveAll(targetDir); err != nil {
			log.Fatal(err)
		}
		if err := os.RemoveAll(getBackingDir(targetDir)); err != nil {
			log.Fatal(err)
		}

		if err := os.Remove(mole.GetConfigPath(targetDir)); err != nil {
			log.Fatal(err)
		}
	} else {
		printHelp()
		return
	}
}

func printHelp() {
	fmt.Println("Usage: mole <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  run <image>:<tag> <targetDir>")
	fmt.Println("  run <targetDir>")
	fmt.Println("  rm <targetDir>")
	fmt.Println("  help")
}

func run(image, tag, targetDir string) {
	targetDirBacking := getBackingDir(targetDir)

	config, err := mole.GetConfig(image, tag)
	if err != nil {
		log.Fatal(err)
	}

	if err := mole.StoreConfigForDir(targetDir, config.Config); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Preparing image")

	err = mole.GetImage(targetDirBacking, image, config.Manifest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting image")

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Fatal(err)
		}

		mountBind(targetDirBacking, targetDir)

		mountProc(path.Join(targetDir, "proc"))
	}

	err = mole.RunContainer(config.Config, targetDir)
	if err != nil {
		log.Fatal(err)
	}
}

func mountProc(mountPoint string) {
	cmd := exec.Command("mount", "-t", "proc", "none", mountPoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to mount proc filesystem: %s", err)
	}
}

func mountBind(src, dst string) {
	cmd := exec.Command("mount", "--bind", src, dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to mount bind filesystem: %s", err)
	}
}

func umount(mountPoint string) {
	cmd := exec.Command("umount", mountPoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to umount filesystem: %s", err)
	}
}
