package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/caicloud/logging-admin/pkg/util/graceful"
	"github.com/caicloud/logging-admin/pkg/util/osutil"

	"github.com/caicloud/nirvana/log"
	"gopkg.in/fsnotify/fsnotify.v1"
)

var (
	filebeatExecutablePath = osutil.Getenv("FB_EXE_PATH", "filebeat")
	srcConfigPath          = osutil.Getenv("SRC_CONFIG_PATH", "/config/filebeat-output.yml")
	dstConfigPath          = osutil.Getenv("DST_CONFIG_PATH", "/etc/filebeat/filebeat.yml")
)

// When configmap being created for the first time, following events received:
// INFO  1206-09:38:39.496+00 main.go:41 | Event: "/config/..2018_12_06_09_38_39.944532540": CREATE
// INFO  1206-09:38:39.496+00 main.go:41 | Event: "/config/..2018_12_06_09_38_39.944532540": CHMOD
// INFO  1206-09:38:39.497+00 main.go:41 | Event: "/config/filebeat-output.yml": CREATE
// INFO  1206-09:38:39.497+00 main.go:41 | Event: "/config/..data_tmp": RENAME
// INFO  1206-09:38:39.497+00 main.go:41 | Event: "/config/..data": CREATE
// INFO  1206-09:38:39.497+00 main.go:41 | Event: "/config/..2018_12_06_09_37_32.878326343": REMOVE
// When configmap being modified, following events received:
// INFO  1206-09:42:56.488+00 main.go:41 | Event: "/config/..2018_12_06_09_42_56.160544363": CREATE
// INFO  1206-09:42:56.488+00 main.go:41 | Event: "/config/..2018_12_06_09_42_56.160544363": CHMOD
// INFO  1206-09:42:56.488+00 main.go:41 | Event: "/config/..data_tmp": RENAME
// INFO  1206-09:42:56.488+00 main.go:41 | Event: "/config/..data": CREATE
// INFO  1206-09:42:56.488+00 main.go:41 | Event: "/config/..2018_12_06_09_38_39.944532540": REMOVE
func watchFileChange(path string, reloadCh chan<- struct{}) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := w.Add(path); err != nil {
		return err
	}

	for {
		select {
		case ev := <-w.Events:
			log.Infoln("Event:", ev.String())
			if ev.Op&fsnotify.Create == fsnotify.Create {
				// ref_link [https://github.com/jimmidyson/configmap-reload/issues/6#issuecomment-355203620]
				//ConfigMap volumes use an atomic writer. You could familarize
				//yourself with the mechanic how atomic writes are implemented.
				//In the end you could check if the actual change you do in 				//your ConfigMap results in the rename of the ..data-symlink (step 9).
				if filepath.Base(ev.Name) == "..data" {
					log.Infoln("Configmap updated")
					reloadCh <- struct{}{}
				}
			}
		case err := <-w.Errors:
			log.Errorf("Watch error: %v", err)
		}
	}
}

func start(cmd *exec.Cmd) (<-chan struct{}, error) {
	cmd.Start()
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	exited := make(chan struct{})
	go func(ch chan struct{}) {
		cmd.Wait()
		close(ch)
	}(exited)

	return exited, nil
}

func stop(cmd *exec.Cmd, exited <-chan struct{}) error {
	log.Infoln("Send TERM signal")
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	select {
	case <-exited:
		return nil
	case <-time.After(60 * time.Second):
		log.Infoln("Kill Process")
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
	}

	<-exited
	return nil
}

func run(stopCh <-chan struct{}) error {
	reloadCh := make(chan struct{}, 1)
	started := false
	cmd := newCmd()
	var exited <-chan struct{}

	go watchFileChange(filepath.Dir(srcConfigPath), reloadCh)

	if err := applyChange(); err == nil {
		reloadCh <- struct{}{}
	} else {
		log.Errorf("Error generate config file: %v", err)
		log.Infoln("Filebeat will not start until configmap being updated")
	}

	check := time.Tick(10 * time.Second)
	for {
		select {
		case <-stopCh:
			log.Infoln("Wait filebeat shutdown")
			if err := cmd.Wait(); err != nil {
				return fmt.Errorf("filebeat quit with error: %v", err)
			}
			return nil
		case <-reloadCh:
			log.Infoln("Reload")
			if err := applyChange(); err != nil {
				log.Errorln("Error apply change:", err)
				continue
			}

			var err error
			if !started {
				if exited, err = start(cmd); err != nil {
					return fmt.Errorf("error run filebeat: %v", err)
				}
				log.Infoln("Filebeat start")
				started = true
			} else {
				if err = stop(cmd, exited); err != nil {
					return fmt.Errorf("filebeat quit with error: %v", err)
				}
				log.Infoln("Filebeat quit")

				cmd = newCmd()
				if exited, err = start(cmd); err != nil {
					return fmt.Errorf("error run filebeat: %v", err)
				}
			}
		case <-check:
			if started {
				if cmd != nil && cmd.ProcessState.Exited() {
					log.Fatalln("Filebeat has unexpectedly exited: %v", cmd.ProcessState.ExitCode())
					os.Exit(1)
				}
			}
		}
	}
}

func applyChange() error {
	outputCfgData, err := ioutil.ReadFile(srcConfigPath)
	if err != nil {
		return err
	}

	tmplData, err := ioutil.ReadFile("/etc/filebeat/filebeat.yml.tpl")
	if err != nil {
		return err
	}

	cfg := map[string]interface{}{}
	if err := yaml.Unmarshal(outputCfgData, &cfg); err != nil {
		return fmt.Errorf("error decode output config yaml: %v", err)
	}

	t, err := template.New("filebeat").Parse(string(tmplData))
	if err != nil {
		return err
	}

	f, err := os.OpenFile(dstConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := t.Execute(f, cfg); err != nil {
		return fmt.Errorf("error rendor template: %v", err)
	}

	generated, _ := ioutil.ReadFile(dstConfigPath)
	fmt.Println(string(generated))

	return nil
}

var (
	fbArgs []string
)

func newCmd() *exec.Cmd {
	log.Infof("Will run filebeat with command: %v %v", filebeatExecutablePath, fbArgs)
	cmd := exec.Command(filebeatExecutablePath, fbArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}

func main() {
	fbArgs = os.Args[1:]
	os.Args = os.Args[:1]
	flag.Parse()

	closeCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := run(closeCh); err != nil {
			log.Fatalln("Error run keeper:", err)
		}
		wg.Done()
	}()
	go graceful.HandleSignal(closeCh, nil)
	wg.Wait()
}
