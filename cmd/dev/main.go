// Command dev starts the database, builds the front end and runs the server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

const webDir = "web"

type step struct {
	dir  string
	name string
	args []string
}

func main() {
	reset := flag.Bool("reset", false, "wipe the database volume to restore the sample data")
	flag.Parse()

	err := run(*reset)
	if err != nil {
		log.Fatal(err)
	}
}

func run(reset bool) error {
	err := requireTools("docker", "npm", "go")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, s := range steps(reset) {
		err = execute(ctx, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func steps(reset bool) []step {
	var planned []step
	if reset {
		planned = append(planned, step{name: "docker", args: []string{"compose", "down", "-v"}})
	}
	return append(planned,
		step{name: "docker", args: []string{"compose", "up", "-d", "--wait"}},
		step{dir: webDir, name: "npm", args: []string{"ci"}},
		step{dir: webDir, name: "npm", args: []string{"run", "build"}},
		step{name: "go", args: []string{"run", "./cmd/server"}},
	)
}

func requireTools(names ...string) error {
	for _, name := range names {
		_, err := exec.LookPath(name)
		if err != nil {
			return fmt.Errorf("%s is required, see README.md: %w", name, err)
		}
	}
	return nil
}

func execute(ctx context.Context, s step) error {
	log.Printf("> %s %v", s.name, s.args)
	command := exec.CommandContext(ctx, s.name, s.args...)
	command.Dir = s.dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		return fmt.Errorf("%s %v: %w", s.name, s.args, err)
	}
	return nil
}
