package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

func LogApi(handle func(http.ResponseWriter, *http.Request, *log.Logger), logger *log.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, logger)
	}
}

func start(args ...string) (pid int, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{os.Stdin,
			os.Stdout, os.Stderr}
		p, err := os.StartProcess(args[0], args, &procAttr)
		if err != nil {
			return -1, err
		}
		pid = p.Pid;
		go func() {
			p.Wait()
			p.Release()
		}()
		return pid, nil
	}
	return -1, err
}

func query(pid string) (err error) {

	ps_cmd := "ps -q " + pid
	cmd := exec.Command("bash", "-c", ps_cmd)
	var stderr bytes.Buffer
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	} else {
		line_num := strings.Count(out.String(), "\n")
		if line_num == 2 {
			return nil
		} else {
			return errors.New("process not found")
		}
	}
}

func stop(pid string) (err error) {

	kill_cmd := "kill -9 " + pid
	cmd := exec.Command("bash", "-c", kill_cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func Start(w http.ResponseWriter, r *http.Request, l *log.Logger) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Printf("An error occurred while reading response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	p, err := start("bash", "-c", string(body))
	if err != nil {
		l.Printf("Process error: %v", err)
		w.Write([]byte("error"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmp := fmt.Sprintf("%v", p)
	l.Printf("start process " + tmp)
	w.Write([]byte(tmp))
	w.WriteHeader(http.StatusOK)
}

func Query(w http.ResponseWriter, r *http.Request, l *log.Logger) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Printf("An error occurred while reading response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	pid, err := strconv.Atoi(string(body))
	if err != nil {
		l.Printf("pid %s", body);
		w.Write([]byte("pid error"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if pid <= 0 {
		l.Printf("pid %s", body);
		w.Write([]byte("pid error"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = query(string(body))
	if err != nil {
		l.Printf("Process not found: %v", err)
		w.Write([]byte("Process not found"))
		w.WriteHeader(http.StatusOK)
		return
	}
	l.Printf("Process %s running\n", body)
	w.Write([]byte("running"))
	w.WriteHeader(http.StatusOK)
}

func Stop(w http.ResponseWriter, r *http.Request, l *log.Logger) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Printf("An error occurred while reading response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	err = stop(string(body))
	if err != nil {
		l.Printf("Process not found: %v", err)
		w.Write([]byte("process not found"))
		w.WriteHeader(http.StatusOK)
		return
	}
	l.Printf("stop process %s", body)
	w.Write([]byte("stoped"))
	w.WriteHeader(http.StatusOK)
}

