/*
 *  Copyright 2020 F5 Networks
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package events

import (
	"bufio"
	"github.com/tevino/abool"
	"strings"
	"sync/atomic"
	"testing"
)

func TestEventFlowWith4Workers(t *testing.T) {
	wasStarted := false
	var workers int64 = 0
	wasReloaded := false

	onStart := Trigger{
		Name: "on-start",
		Function: func(Message) error {
			wasStarted = true
			return nil
		},
	}
	onWorkerStart := Trigger{
		Name: "on-worker-start",
		Function: func(Message) error {
			atomic.AddInt64(&workers, 1)
			return nil
		},
	}

	onReload := Trigger{
		Name: "on-reload",
		Function: func(Message) error {
			wasReloaded = true
			return nil
		},
	}
	onWorkersExit := Trigger{
		Name: "on-workers-exit",
		Function: func(Message) error {
			atomic.AddInt64(&workers, -1)
			return nil
		},
	}

	Init()

	GlobalEvents.NginxStart.AddTrigger(&onStart)
	GlobalEvents.NginxWorkerStart.AddTrigger(&onWorkerStart)
	GlobalEvents.NginxReload.AddTrigger(&onReload)
	GlobalEvents.NginxWorkerExit.AddTrigger(&onWorkersExit)

	startLogs := `4975#4975: using the "epoll" event method 
4975#4975: nginx/1.17.9 (nginx-plus-r21) 
4975#4975: built by gcc 7.5.0 (Ubuntu 7.5.0-3ubuntu1~18.04) 
4975#4975: OS: Linux 5.3.0-46-generic 
4975#4975: getrlimit(RLIMIT_NOFILE): 1048576:1048576 
4975#4975: start worker processes
4975#4975: start worker process 4976  
4975#4975: start worker process 4977  
4975#4975: start worker process 4978  
4975#4975: start worker process 4980`

	reloadLogs := `4975#4975: signal 1 (SIGHUP) received from 4945, reconfiguring 
4975#4975: reconfiguring              
4975#4975: using the "epoll" event method 
4975#4975: start worker processes     
4975#4975: start worker process 4993  
4975#4975: start worker process 4995  
4975#4975: start worker process 4996  
4975#4975: start worker process 4997
4980#4980: gracefully shutting down   
4976#4976: gracefully shutting down   
4977#4977: gracefully shutting down   
4978#4978: gracefully shutting down   
4976#4976: exiting                    
4980#4980: exiting                    
4978#4978: exiting                    
4977#4977: exiting                    
4980#4980: exit                       
4976#4976: exit                       
4978#4978: exit                       
4977#4977: exit                       
4975#4975: signal 17 (SIGCHLD) received from 4977 
4975#4975: worker process 4977 exited with code 0 
4975#4975: worker process 4980 exited with code 0 
4975#4975: signal 29 (SIGIO) received 
4975#4975: signal 17 (SIGCHLD) received from 4976 
4975#4975: worker process 4976 exited with code 0 
4975#4975: signal 29 (SIGIO) received 
4975#4975: signal 17 (SIGCHLD) received from 4978 
4975#4975: worker process 4978 exited with code 0`

	exitLogs := `4975#4975: signal 29 (SIGIO) received 
4995#4995: signal 2 (SIGINT) received, exiting 
4975#4975: signal 2 (SIGINT) received, exiting 
4997#4997: signal 2 (SIGINT) received, exiting 
4996#4996: signal 2 (SIGINT) received, exiting 
4993#4993: signal 2 (SIGINT) received, exiting 
4995#4995: epoll_wait() failed (4: Interrupted system call) 
4996#4996: epoll_wait() failed (4: Interrupted system call) 
4997#4997: epoll_wait() failed (4: Interrupted system call) 
4995#4995: exiting                    
4996#4996: exiting                    
received signal interrupt                    
process stopped due to: interrupt            
4997#4997: exiting                    
4993#4993: epoll_wait() failed (4: Interrupted system call) 
4993#4993: exiting                    
4997#4997: exit                       
4995#4995: exit                       
4996#4996: exit                       
4993#4993: exit                       
killing nginx with signal 15      
4975#4975: signal 15 (SIGTERM) received from 4945, exiting 
4975#4975: signal 17 (SIGCHLD) received from 4995 
4975#4975: worker process 4995 exited with code 0 
4975#4975: signal 29 (SIGIO) received 
4975#4975: signal 17 (SIGCHLD) received from 4996 
4975#4975: worker process 4996 exited with code 0
4975#4975: signal 29 (SIGIO) received
4975#4975: signal 17 (SIGCHLD) received from 4993
4975#4975: worker process 4993 exited with code 0
4975#4975: signal 29 (SIGIO) received
4975#4975: signal 17 (SIGCHLD) received from 4997
4975#4975: worker process 4997 exited with code 0
4975#4975: exit`

	reader := strings.NewReader(startLogs)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		msg := strings.TrimRight(scanner.Text(), " \t")
		GlobalEvents.ParseForTriggerableEvent(msg)
	}

	if !wasStarted {
		t.Error("no start event emitted")
	}
	if workers != 4 {
		t.Errorf("incorrect number of workers: %d", workers)
	}

	GlobalEvents.ReloadStarted = abool.NewBool(true)

	reader = strings.NewReader(reloadLogs)
	scanner = bufio.NewScanner(reader)

	for scanner.Scan() {
		msg := strings.TrimRight(scanner.Text(), " \t")
		GlobalEvents.ParseForTriggerableEvent(msg)
	}

	if !wasReloaded {
		t.Error("No reload event emitted")
	}

	reader = strings.NewReader(exitLogs)
	scanner = bufio.NewScanner(reader)

	for scanner.Scan() {
		msg := strings.TrimRight(scanner.Text(), " \t")
		GlobalEvents.ParseForTriggerableEvent(msg)
	}

	if workers != 0 {
		t.Errorf("incorrect number of workers: %d", workers)
	}
}
