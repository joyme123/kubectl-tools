package remote

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/joyme123/kubectl-tools/tools"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	utilexec "k8s.io/client-go/util/exec"
)

// Run ...
func Run(kube *KubeRequest, t tools.Tool, cmd []string) error {
	exitCode, err := PodExecuteCommand(ExecCommandRequest{
		KubeRequest: kube,
		Command:     []string{"which", cmd[0]},
	})
	if err != nil {
		return err
	}

	// cmd doesn't exist
	if exitCode == 1 {
		src, err := tools.GetLocalPath(t)
		if err != nil {
			return err
		}
		dst := "/tmp/" + t.Name
		exitCode, err := PodUploadFile(UploadFileRequest{
			KubeRequest: kube,
			Src:         src,
			Dst:         dst,
		})
		if err != nil {
			return err
		}
		if exitCode != 0 {
			return fmt.Errorf("upload file failed")
		}
		cmd[0] = dst
	}

	exitCode, err = PodExecuteCommand(ExecCommandRequest{
		KubeRequest: kube,
		Command:     cmd,
		StdOut:      os.Stdout,
		StdErr:      os.Stderr,
	})
	if err != nil {
		return fmt.Errorf("run command %s failed: %w", strings.Join(cmd, " "), err)
	}
	log.Info("exit with code: %d", exitCode)
	return nil
}

type KubeRequest struct {
	Clientset  *kubernetes.Clientset
	RestConfig *rest.Config
	Namespace  string
	Pod        string
	Container  string
}

type ExecCommandRequest struct {
	*KubeRequest
	Command []string
	StdIn   io.Reader
	StdOut  io.Writer
	StdErr  io.Writer
}

type UploadFileRequest struct {
	*KubeRequest
	Src string
	Dst string
}

func (w *NopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type NopWriter struct {
}

func (w *Writer) Write(p []byte) (n int, err error) {
	str := string(p)
	if len(str) > 0 {
		w.Output += str
	}
	return len(str), nil
}

type Writer struct {
	Output string
}

func PodUploadFile(req UploadFileRequest) (int, error) {
	stdOut := new(Writer)
	stdErr := new(Writer)

	log.Debugf("uploading file from: '%s' to '%s'", req.Src, req.Dst)

	fileContent, err := ioutil.ReadFile(req.Src)
	if err != nil {
		return 0, err
	}

	log.Debugf("read '%s' to memory, file size: '%d'", req.Src, len(fileContent))

	destFileName := path.Base(req.Dst)
	tarFile, err := WrapAsTar(destFileName, fileContent)
	if err != nil {
		return 0, err
	}

	log.Debugf("formatted '%s' as tar, tar size: '%d'", req.Src, len(tarFile))

	stdIn := bytes.NewReader(tarFile)

	tarCmd := []string{"tar", "-xf", "-"}

	destDir := path.Dir(req.Dst)
	if len(destDir) > 0 {
		tarCmd = append(tarCmd, "-C", destDir)
	}

	log.Debugf("executing tar: '%v'", tarCmd)

	execTarRequest := ExecCommandRequest{
		KubeRequest: &KubeRequest{
			Clientset:  req.Clientset,
			RestConfig: req.RestConfig,
			Namespace:  req.Namespace,
			Pod:        req.Pod,
			Container:  req.Container,
		},
		Command: tarCmd,
		StdIn:   stdIn,
		StdOut:  stdOut,
		StdErr:  stdErr,
	}

	exitCode, err := PodExecuteCommand(execTarRequest)

	log.Debugf("done uploading file, exitCode: '%d', stdOut: '%s', stdErr: '%s'",
		exitCode, stdOut.Output, stdErr.Output)

	return exitCode, err
}

func PodExecuteCommand(req ExecCommandRequest) (int, error) {

	execRequest := req.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(req.Pod).
		Namespace(req.Namespace).
		SubResource("exec")

	execRequest.VersionedParams(&corev1.PodExecOptions{
		Container: req.Container,
		Command:   req.Command,
		Stdin:     req.StdIn != nil,
		Stdout:    req.StdOut != nil,
		Stderr:    req.StdErr != nil,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(req.RestConfig, "POST", execRequest.URL())
	if err != nil {
		return 0, err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  req.StdIn,
		Stdout: req.StdOut,
		Stderr: req.StdErr,
		Tty:    false,
	})

	var exitCode = 0

	if err != nil {
		if exitErr, ok := err.(utilexec.ExitError); ok && exitErr.Exited() {
			exitCode = exitErr.ExitStatus()
			err = nil
		}
	}

	return exitCode, err
}

func WrapAsTar(fileNameOnTar string, fileContent []byte) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name: fileNameOnTar,
		Mode: 0755,
		Size: int64(len(fileContent)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}

	if _, err := tw.Write(fileContent); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
