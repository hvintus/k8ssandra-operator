package kubectl

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Options struct {
	Namespace  string
	Context    string
	ServerSide bool
}

var logOutput = false

func LogOutput(enabled bool) {
	logOutput = enabled
}

func Apply(opts Options, arg interface{}) error {
	cmd := exec.Command("kubectl")

	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}

	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}

	cmd.Args = append(cmd.Args, "apply")

	if opts.ServerSide {
		cmd.Args = append(cmd.Args, "--server-side", "--force-conflicts")
	}

	cmd.Args = append(cmd.Args, "-f")

	if buf, ok := arg.(*bytes.Buffer); ok {
		cmd.Stdin = buf
		cmd.Args = append(cmd.Args, "-")
	} else if s, ok := arg.(string); ok {
		cmd.Args = append(cmd.Args, s)
	} else {
		return errors.New("Expected arg to be a *bytes.Buffer or a string")
	}

	fmt.Println(cmd)

	output, err := cmd.CombinedOutput()

	if logOutput || err != nil {
		fmt.Println(string(output))
	}

	return err
}

func Delete(opts Options, arg interface{}) error {
	cmd := exec.Command("kubectl")

	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}

	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}

	cmd.Args = append(cmd.Args, "delete", "-f")

	if buf, ok := arg.(*bytes.Buffer); ok {
		cmd.Stdin = buf
		cmd.Args = append(cmd.Args, "-")
	} else if s, ok := arg.(string); ok {
		cmd.Args = append(cmd.Args, s)
	} else {
		return errors.New("Expected arg to be a *bytes.Buffer or a string")
	}

	fmt.Println(cmd)

	output, err := cmd.CombinedOutput()

	if logOutput || err != nil {
		fmt.Println(string(output))
	}

	return err
}

func DeleteByName(opts Options, kind, name string, ignoreNotFound bool) error {
	cmd := exec.Command("kubectl")
	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}
	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}
	cmd.Args = append(cmd.Args, "delete", kind, name)
	if ignoreNotFound {
		cmd.Args = append(cmd.Args, "--ignore-not-found")
	}
	fmt.Println(cmd)
	output, err := cmd.CombinedOutput()
	if logOutput || err != nil {
		fmt.Println(string(output))
	}
	return err
}

func DeleteAllOf(opts Options, kind string) error {
	cmd := exec.Command("kubectl")
	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}
	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	} else {
		return errors.New("Namespace is required for delete --all")
	}
	cmd.Args = append(cmd.Args, "delete", kind, "--all")
	fmt.Println(cmd)
	output, err := cmd.CombinedOutput()
	if logOutput || err != nil {
		fmt.Println(string(output))
	}
	return err
}

func WaitForCondition(opts Options, condition string, args ...string) error {

	cmd := exec.Command("kubectl")

	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}

	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}

	cmd.Args = append(cmd.Args, "wait", "--for", "condition="+condition)
	cmd.Args = append(cmd.Args, args...)

	output, err := cmd.CombinedOutput()

	if logOutput || err != nil {
		fmt.Println(string(output))
	}

	return err
}

// Exec executes a command against a Cassandra pod and the cassandra container in
// particular. This does not currently handle pipes.
func Exec(opts Options, pod string, args ...string) (string, error) {
	cmd := exec.Command("kubectl")

	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}

	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}

	cmd.Args = append(cmd.Args, "exec", "-i", pod, "-c", "cassandra", "--")
	cmd.Args = append(cmd.Args, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Println(cmd)

	err := cmd.Run()

	if logOutput {
		fmt.Println(stderr.String())
		fmt.Println(stdout.String())
	}

	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}

type ClusterInfoOptions struct {
	Options

	Namespaces []string

	OutputDirectory string
}

func DumpClusterInfo(opts ClusterInfoOptions) error {
	cmd := exec.Command("kubectl", "cluster-info", "dump")

	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}

	if len(opts.Namespaces) > 0 {
		cmd.Args = append(cmd.Args, "--namespaces", strings.Join(opts.Namespaces, ","))
	}

	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}

	cmd.Args = append(cmd.Args, "-o", "yaml")

	dir, err := filepath.Abs(opts.OutputDirectory)
	if err != nil {
		return err
	}

	cmd.Args = append(cmd.Args, "--output-directory", dir)

	output, err := cmd.CombinedOutput()

	if logOutput || err != nil {
		fmt.Println(string(output))
	}

	return err
}

func Get(opts Options, args ...string) (string, error) {
	cmd := exec.Command("kubectl", "get")
	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}
	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}

	cmd.Args = append(cmd.Args, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Println(cmd)

	err := cmd.Run()

	if logOutput || err != nil {
		fmt.Println(stderr.String())
		fmt.Println(stdout.String())
	}

	if err != nil {
		return stdout.String() + stderr.String(), err
	}
	return stdout.String(), nil
}

func RolloutStatus(ctx context.Context, opts Options, kind, name string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "rollout", "status")
	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}
	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}
	cmd.Args = append(cmd.Args, kind, name)
	fmt.Println(cmd)
	output, err := cmd.CombinedOutput()
	if logOutput || err != nil {
		fmt.Println(string(output))
	}
	return err
}

func Logs(opts Options, args ...string) (string, error) {
	cmd := exec.Command("kubectl", "logs")
	if len(opts.Context) > 0 {
		cmd.Args = append(cmd.Args, "--context", opts.Context)
	}
	if len(opts.Namespace) > 0 {
		cmd.Args = append(cmd.Args, "-n", opts.Namespace)
	}
	cmd.Args = append(cmd.Args, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Println(cmd)

	err := cmd.Run()

	if logOutput || err != nil {
		fmt.Println(stderr.String())
		fmt.Println(stdout.String())
	}

	if err != nil {
		return stdout.String() + stderr.String(), err
	}
	return stdout.String(), nil
}
